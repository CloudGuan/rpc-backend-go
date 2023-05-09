package idlrpc

import (
	"context"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/common"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/errors"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/log"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/transport"
)

// StubCallQueue service remote call queue
type stubCallQueue chan *StubCall
type stopSign chan struct{}

// stubWrapper user stub wrapper
type stubWrapper struct {
	isClose   int32           //is this service not service again, 0 not, 1 closed
	srvImp    IStub           //stub interface user implemenet
	wg        sync.WaitGroup  //worker goroutine waiter
	callQueue stubCallQueue   //rpc remote call queue
	stopCh    stopSign        //stop signal channel, maybe be replaced with context cancel function
	logger    log.ILogger     //logger instance
	ctx       context.Context //graceful close single
}

// newStubWrapper create stubbase while service register
func newStubWrapper(impl IStub, logger log.ILogger) *stubWrapper {
	if impl == nil {
		panic("[IStub] register invalid service ")
		return nil
	}

	return &stubWrapper{
		isClose:   0,
		srvImp:    impl,
		wg:        sync.WaitGroup{},
		callQueue: make(stubCallQueue, common.DefaultCallCache),
		stopCh:    make(stopSign),
		logger:    logger,
	}
}

// Init service, after register
func (s *stubWrapper) init(ctx context.Context) (err error) {
	// add recover function to avoid throwing panic in OnTick function
	defer func() {
		if r := recover(); r != nil {
			if s.srvImp == nil {
				//In theory, it will not enter this branch forever
				s.logger.Warn("[Service] 0,0,0 service implement interface is nil !")
				err = errors.ErrServicePanic
				return
			}
			err = errors.NewRpcError(errors.CommErr, "service %s throw panic while executing the initialization function", s.srvImp.GetServiceName())
			s.logger.Warn("[Service] %s,%d,0 service throw exception %v on init ", s.srvImp.GetServiceName(), s.srvImp.GetUUID(), r)
			if stackTrace {
				s.logger.Error("trace back: %s", string(debug.Stack()))
			}
		}
	}()

	if !s.srvImp.OnAfterFork(ctx) {
		err = errors.ErrServiceInit
	}
	return
}

// start start service working loop
func (s *stubWrapper) start() {
	if s.srvImp == nil {
		s.logger.Error("[Service] service implement interface is invalid")
		panic("invalid service")
	}

	num := s.srvImp.GetMutipleNum()
	if num == 0 {
		num = common.DefaultServiceWorker
	}

	//start worker
	for i := uint32(0); i < num; i++ {
		s.wg.Add(1)
		go s.loop()
	}
}

// loop rpc remote call worker, read stub call from message queue
func (s *stubWrapper) loop() {
	defer func() {
		//maybe panic
		if r := recover(); r != nil {
			s.logger.Error("[Service] %s,%d,0 service method call panic!!", s.srvImp.GetServiceName(), s.srvImp.GetUUID())
			if stackTrace {
				s.logger.Error("trace back: %s", string(debug.Stack()))
			}
		}
		s.wg.Done()
	}()

	for {

		// The try-receive operation here is to
		// try to exit the sender goroutine as
		// early as possible. Try-receive and
		// try-send select blocks are specially
		// optimized by the standard Go
		// compiler, so they are very efficient.
		select {
		case <-s.stopCh:
			return
		default:
		}

		//emmmmmmmm
		select {
		case <-s.stopCh:
			return
		case call := <-s.callQueue:
			// stub call has been set to nil while shutdown rpc framework
			if call == nil {
				return
			}

			err := s.doCallMethod(call)
			//TODO: destroy stub call
			if err != nil {
				//TODO: add record of error code
				return
			}
		}
	}
}

func (s *stubWrapper) doCallMethod(stubCall *StubCall) (err error) {
	if stubCall == nil {
		s.logger.Error("[Service] %s,%d,0 stub call pointer is invalid", s.srvImp.GetServiceName(), s.srvImp.GetUUID())
		return errors.ErrStubCallInvalid
	}
	//recover function, not break loop
	defer func() {
		//recover panic
		if r := recover(); r != nil {
			if !s.srvImp.IsOneWay(stubCall.MethodID()) {
				//build rpc response, notify client exception
				var pkg []byte
				if info, ok := r.(errors.RpcPanicInfo); ok {
					pkg = info.Pkg
				}
				resp := protocol.BuildException(stubCall.CallID(), pkg)
				respData, pkgLen := protocol.PackRespMsg(resp)
				if respData == nil || pkgLen == 0 {
					s.logger.Error("[Service] %s,%d,0 serialize response bytes error !", s.srvImp.GetServiceName(), s.srvImp.GetUUID())
					return
				} else {
					s.logger.Error("[Service] %s,%d,0 runtime error: %v ", s.srvImp.GetServiceName(), stubCall.MethodID(), r)
					if stackTrace {
						s.logger.Error("trace back: %s", string(debug.Stack()))
					}
				}
				err = stubCall.doRet(respData)
				if err != nil {
					return
				}
			}
		}
	}()

	if stubCall.globalID == InvalidGlobalIndex {
		return s.rpcCall(stubCall)
	} else {
		return s.rpcProxyCall(stubCall)
	}
}

func (s *stubWrapper) rpcCall(stubCall *StubCall) (err error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, callkey{}, stubCall)
	//not check transport first
	buffer, err := s.srvImp.Call(ctx, stubCall.MethodID(), stubCall.buffer)
	// not one-way function, send response
	if !s.srvImp.IsOneWay(stubCall.MethodID()) {
		execCode := protocol.IDL_SUCCESS
		if err != nil {
			//log error
			s.logger.Warn("[Service] %s,%d,%d method %s call error %v", s.srvImp.GetServiceName(), s.srvImp.GetUUID(), stubCall.CallID(), s.srvImp.GetSignature(stubCall.MethodID()), err)
			//set error
			execCode = protocol.IDL_SERVICE_ERROR
		}
		//Build response package
		resp := &protocol.ResponsePackage{
			Buffer: buffer,
		}
		//srvID compatible with cpp implement, must be zero in golang
		protocol.BuildRespHeader(resp, 0, stubCall.CallID(), execCode)
		respData, pkgLen := protocol.PackRespMsg(resp)
		if respData == nil || pkgLen == 0 {
			s.logger.Error("[Service] %s,%d,0 serialize response bytes error !", s.srvImp.GetServiceName(), s.srvImp.GetUUID())
			err = errors.NewMethodExecError(s.srvImp.GetServiceName(), s.srvImp.GetSignature(stubCall.MethodID()))
			return
		}
		err = stubCall.doRet(respData)
		if err != nil {
			s.logger.Warn("[Service] %s,%d,%d method %s send data error!", s.srvImp.GetServiceName(), s.srvImp.GetUUID(), stubCall.CallID(), s.srvImp.GetSignature(stubCall.MethodID()), err)
			return
		}
	}
	return
}

func (s *stubWrapper) rpcProxyCall(stubCall *StubCall) (err error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, callkey{}, stubCall)
	//not check transport first
	buffer, err := s.srvImp.Call(ctx, stubCall.MethodID(), stubCall.buffer)
	// not one-way function, send response
	if !s.srvImp.IsOneWay(stubCall.MethodID()) {
		execCode := protocol.IDL_SUCCESS
		if err != nil {
			//log error
			s.logger.Warn("[Service] %s,%d,%d method %s exec error %v", s.srvImp.GetServiceName(), s.srvImp.GetUUID(), stubCall.CallID(), s.srvImp.GetSignature(stubCall.MethodID()), err)
			//set error
			execCode = protocol.IDL_SERVICE_ERROR
		}
		//Build response package
		resp := &protocol.ProxyRespPackage{
			Buffer: buffer,
		}
		//srvID compatible with cpp implement, must be zero in golang
		protocol.BuildProxyRespHeader(resp, 0, stubCall.CallID(), execCode, stubCall.globalID)
		respData, pkgLen := protocol.PackProxyRespMsg(resp)
		if respData == nil || pkgLen == 0 {
			s.logger.Error("[Service] %s,%d,0 serialize response bytes error !", s.srvImp.GetServiceName(), s.srvImp.GetUUID())
			err = errors.NewMethodExecError(s.srvImp.GetServiceName(), s.srvImp.GetSignature(stubCall.MethodID()))
			return
		}
		err = stubCall.doRet(respData)
		if err != nil {
			s.logger.Warn("[Service] %s,%d,%d method %s send data error!", s.srvImp.GetServiceName(), s.srvImp.GetUUID(), stubCall.CallID(), s.srvImp.GetSignature(stubCall.MethodID()), err)
			return
		}
	}
	return nil
}

func (s *stubWrapper) tick() {
	// add recover function to avoid throwing panic in OnTick function
	defer func() {
		if r := recover(); r != nil {
			if s.srvImp == nil {
				//In theory, it will not enter this branch forever
				s.logger.Warn("[Service] 0,0,0 service implement interface is nil !")
				return
			}
			s.logger.Warn("[Service] %s,%d,0 service throw exception %v on tick ", s.srvImp.GetServiceName(), s.srvImp.GetUUID(), r)
		}
	}()
	//TODO check impl status
	s.srvImp.OnTick()
}

// close close service
func (s *stubWrapper) close() {
	//atomic check close status
	isClose := atomic.LoadInt32(&s.isClose)
	if isClose == 1 {
		s.logger.Warn("[Service] %s,%d,0 stub close multi times", s.srvImp.GetServiceName(), s.srvImp.GetUUID())
		return
	}
	//set close status
	if !atomic.CompareAndSwapInt32(&s.isClose, isClose, 1) {
		// set false, close status have been changed by other goroutine
		s.logger.Info("[Service] %s,%d,0 service has been closed by other goroutine!")
		return
	}
	close(s.callQueue)
	//close stop channel
	close(s.stopCh)
	s.wg.Wait()
	s.srvImp.OnBeforeDestroy()
}

// addCall add stubcall to service call queue
func (s *stubWrapper) addCall(call *StubCall) error {
	//check status
	if atomic.LoadInt32(&s.isClose) != 0 {
		s.logger.Warn("[Service] %s,%d,0 service has been shutdown while stub call %d", s.srvImp.GetServiceName(), s.srvImp.GetUUID(), call.CallID())
		return errors.NewRpcError(errors.ServiceShutdown, "service %s has shutdown ", s.srvImp.GetServiceName())
	}

	//check channel
	if s.callQueue == nil {
		return errors.NewRpcError(errors.ServiceShutdown, "service %s has shutdown ", s.srvImp.GetServiceName())
	}
	//add to callQueue
	s.callQueue <- call
	return nil
}

func (s *stubWrapper) isValid() bool {
	return atomic.LoadInt32(&s.isClose) == 0
}

func (s *stubWrapper) doCallService(trans transport.ITransport, stubCall *StubCall) error {
	//check valid
	if !s.isValid() {
		return errors.ErrTransClose
	}
	// add stub call to service
	err := s.addCall(stubCall)
	if err != nil {
		return err
	}
	return nil
}
