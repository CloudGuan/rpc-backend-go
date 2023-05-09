package idlrpc

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/common"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/logger"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/proxy"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/errors"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/log"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/transport"
	"google.golang.org/protobuf/proto"
)

var (
	stackTrace bool
)

type (
	proxyFactoryMap map[uint64]ProxyCreator
	stubFactoryMap  map[uint64]StubCreator

	//callkey will get stub call from ctx
	callkey struct{}

	rpcImpl struct {
		opt            *Options
		proxyMgr       *ProxyManager
		proxyCallMgr   *proxy.ProxyCallManager
		stubMgr        *StubManager
		serviceFactory stubFactoryMap
		logger         log.ILogger //logger handle
		status         int32       // rpc status
	}
)

func (r *rpcImpl) Init(opts ...Option) error {
	r.opt = &Options{
		stackTrace: false,
		callTrace:  false,
		ctx:        context.Background(),
	}
	for _, o := range opts {
		o(r.opt)
	}
	r.logger = r.opt.logger
	stackTrace = r.opt.stackTrace
	return nil
}

func (r *rpcImpl) Start() error {
	if r.logger == nil {
		r.logger = &logger.NullLogger{}
	}
	r.stubMgr.Init(r.logger)
	logger.SetLogger(r.logger)
	r.status = RpcRunning
	r.logger.Info("[Rpc] ===== rpc frame work start working =====")
	return nil
}

func (r *rpcImpl) Tick() error {

	if atomic.LoadInt32(&r.status) != RpcRunning {
		return errors.ErrRpcClosed
	}

	if r.stubMgr != nil {
		r.stubMgr.Tick()
	}
	return nil
}

func (r *rpcImpl) ShutDown() error {
	//check status
	if atomic.LoadInt32(&r.status) != RpcRunning {
		return errors.ErrRpcClosed
	}
	//set status
	atomic.StoreInt32(&r.status, RpcClosed)

	//close all service
	r.stubMgr.UnInit()
	return nil
}

func (r *rpcImpl) Options() *Options {
	return r.opt
}

func (r *rpcImpl) OnMessage(trans transport.ITransport, ctx context.Context) error {
	//read bytes from transport until invalided bytes
	if atomic.LoadInt32(&r.status) != RpcRunning {
		return errors.ErrRpcClosed
	}
	for {
		headers, mLen, err := trans.Peek(protocol.RpcHeadSize)
		if mLen != 8 || err != nil {
			return err
		}

		header := protocol.ReadHeader(headers)
		if header == nil {
			trans.Close()
			return errors.ErrInvalidProto
		}

		// 校验长度
		if header.Length == 0 {
			trans.Close()
			return errors.ErrInvalidProto
		}

		// 校验类型
		if header.Type >= protocol.RpcProtocolMax || header.Type <= protocol.RpcInvalidMsg {
			//协议头校验不对
			trans.Close()
			return errors.ErrInvalidProto
		}

		//body message not arrived yet
		if header.Length > trans.Size() {
			return nil
		}

		//TODO add context usage
		switch header.Type {
		case protocol.RequestMsg:
			if err = r.onCall(trans); err != nil {
				r.logger.Warn("[Rpc] Execution of the rpc request failed, error %v", err)
			}
		case protocol.ResponseMsg:
			if err = r.onReturn(trans); err != nil {
				r.logger.Info("[Rpc] Execution of the rpc response failed,  error %v", err)
			}
		case protocol.ProxyRequestMsg:
			if err = r.onProxyCall(trans); err != nil {
				r.logger.Info("[Rpc] Execution of the proxy call failed, error %v", err)
			}
		case protocol.ProxyResponseMsg:
			if err = r.onProxyReturn(trans); err != nil {
				r.logger.Info("[Rpc] Execution of the proxy response failed, error %v", err)
			}
		case protocol.RpcTimeout:
			if err = r.onOutsideConnTimeout(trans); err != nil {
				r.logger.Info("[Rpc] Execution of the heartbeat notification failed, error: %v", err)
			}
		case protocol.NotRpcMsg:
			break
		default:
			r.logger.Warn("[Rpc] An illegal protocol request %d was received from %s:%d", header.Type, trans.RemoteAddr(), trans.GlobalIndex())
			return errors.ErrInvalidProto
		}
	}
}

func (r *rpcImpl) OnProxyMessage(trans transport.ITransport, ph IProxyHandler) error {
	//read bytes from transport until invalided bytes
	if atomic.LoadInt32(&r.status) != RpcRunning {
		return errors.ErrRpcClosed
	}
	for {
		headers, mLen, err := trans.Peek(protocol.RpcHeadSize)
		if mLen != 8 || err != nil {
			return err
		}

		header := protocol.ReadHeader(headers)
		if header == nil {
			trans.Close()
			r.logger.Error("[RPC] parse rpc header error from addr %s global index %d !", trans.RemoteAddr(), trans.GlobalIndex())
			return errors.ErrInvalidProto
		}

		//校验长度
		if header.Length == 0 {
			trans.Close()
			r.logger.Error("[RPC] %s illegal rpc header length ", trans.RemoteAddr())
			return errors.ErrInvalidProto
		}

		//校验类型
		if header.Type >= protocol.RpcProtocolMax || header.Type <= protocol.RpcInvalidMsg {
			trans.Close()
			r.logger.Error("[RPC] %s illegal rpc protocol type %d ", trans.RemoteAddr(), header.Type)
			return errors.ErrInvalidProto
		}

		//body message not arrived yet
		if header.Length > trans.Size() {
			return nil
		}

		if err := ph.OnRelay(trans, header); err != nil {
			r.logger.Warn("[Rpc] proxy call error %v", err)
		}
	}
}

func (r *rpcImpl) RegisterService(service IService) error {
	if r == nil {
		r.logger.Warn("[Rpc] rpc framework not init yet!")
		return errors.NewRpcError(errors.CommErr, "stub manager is invalid")
	}

	if service == nil {
		r.logger.Warn("[Rpc] pass invalid service interface to rpc framework")
		return errors.NewRpcError(errors.CommErr, "service interface is invalid")
	}

	//check stub get creator
	creator, ok := r.serviceFactory[service.GetUUID()]
	if !ok {
		r.logger.Warn("[Rpc] service %d not register to rpc framework!", service.GetUUID())
		return errors.NewServiceNotExist(service.GetUUID())
	}

	svcStub := creator(service)
	if svcStub == nil {
		r.logger.Warn("[Rpc] service %d creator stub error!", service.GetUUID())
		return errors.NewRpcError(errors.CommErr, "creat stub error !")
	}

	//try add to stub manager
	err := r.stubMgr.Add(r.opt.ctx, svcStub)
	if err != nil {
		r.logger.Warn("[Rpc] register %s service to framework error !", svcStub.GetServiceName())
		return err
	}
	return nil
}

func (r *rpcImpl) Call(srvProxy IProxy, methodId, timeout uint32, retry int32, message proto.Message) (buffer []byte, err error) {
	//get proxy manager
	if r.proxyMgr == nil {
		return nil, errors.NewRpcError(errors.CommErr, "proxy manager is invalid")
	}

	if srvProxy == nil {
		r.logger.Warn("[Rpc] pass invalid proxy id to rpc framework")
		return nil, errors.ErrProxyInvalid
	}

	proxyCall := r.proxyCallMgr.CreateProxyCall(proxy.ProxyUuid(srvProxy.GetID()), timeout, retry, srvProxy.GetGlobalIndex())
	if proxyCall == nil {
		return nil, errors.ErrProxyInvalid
	}
	proxyCall.MethodId = methodId

	err = r.proxyCallMgr.Add(proxyCall)
	if err != nil {
		return nil, err
	}
	defer func() {
		//always destroy
		r.proxyCallMgr.Destroy(proxyCall.CallID)
	}()

	// parameters serialize data, message may be nil
	pkg, err := proto.Marshal(message)
	if err != nil {
		return
	}

	var packData []byte

	if srvProxy.GetGlobalIndex() == InvalidGlobalIndex {

		// wrapper rpc call request
		reqPb := &protocol.RequestPackage{
			Header: &protocol.RpcCallHeader{
				RpcMsgHeader: protocol.RpcMsgHeader{
					Length: uint32(protocol.CallHeadSize + len(pkg)),
					Type:   protocol.RequestMsg,
				},
				ServiceUUID: srvProxy.GetUUID(),
				ServerID:    srvProxy.GetTargetID(),
				CallID:      proxyCall.CallID,
				MethodID:    methodId,
			},
			Buffer: pkg,
		}

		packData, _ = protocol.PackReqMsg(reqPb)
		//startT := time.Now()

	} else {
		reqPb := &protocol.ProxyRequestPackage{
			Header: &protocol.RpcProxyCallHeader{
				RpcMsgHeader: protocol.RpcMsgHeader{
					Length: uint32(protocol.ProxyCallHeadSize + len(pkg)),
					Type:   protocol.ProxyRequestMsg,
				},
				ServiceUUID: srvProxy.GetUUID(),
				ServerID:    srvProxy.GetTargetID(),
				CallID:      proxyCall.CallID,
				MethodID:    methodId,
				GlobalIndex: srvProxy.GetGlobalIndex(),
			},
			Buffer: pkg,
		}
		if srvProxy.IsOneWay(methodId) {
			reqPb.Header.OneWay = 1
		} else {
			reqPb.Header.OneWay = 0
		}
		packData, _ = protocol.PackProxyReqMsg(reqPb)
	}

	buffer, err = callMethod(r, srvProxy, proxyCall, methodId, packData)

	//one way, not care about remote return
	if srvProxy.IsOneWay(methodId) {
		return nil, err
	}
	return
}

func (r *rpcImpl) GetProxyFromPeer(ctx context.Context, uuid uint64) (srvProxy IProxy, err error) {
	if r == nil {
		r.logger.Warn("[Rpc] rpc frame work not init!")
		return nil, nil
	}

	if ctx == nil {
		r.logger.Error("[Rpc] invalid call ctx!")
		return nil, errors.ErrStubCallInvalid
	}

	stubCall := ctx.Value(callkey{}).(*StubCall)
	if stubCall == nil {
		return nil, errors.ErrStubCallInvalid
	}

	srvProxy = r.proxyMgr.getOrCreateProxy(uuid, stubCall.globalID, stubCall.trans)
	if srvProxy == nil {
		err = errors.ErrProxyInvalid
		return
	}
	srvProxy.SetRpc(r)
	return
}

// GetServiceProxy try get service by transport, rpc framework will create proxy while it not exits
func (r *rpcImpl) GetServiceProxy(uuid uint64, trans transport.ITransport) (IProxy, error) {
	if r == nil {
		r.logger.Warn("[Rpc] rpc frame work not init!")
		return nil, errors.ErrRpcNotInit
	}

	if trans == nil || trans.IsClose() {
		return nil, errors.ErrTransClose
	}

	srvProxy := r.proxyMgr.getOrCreateProxy(uuid, trans.GlobalIndex(), trans)
	srvProxy.SetRpc(r)
	return srvProxy, nil
}

func (r *rpcImpl) GetProxyByID(id ProxyId) (IProxy, error) {
	if r == nil {
		r.logger.Warn("[Rpc] rpc frame work not init!")
		return nil, errors.ErrRpcNotInit
	}

	srvProxy, err := r.proxyMgr.Get(id)
	if err != nil {
		return nil, err
	}
	srvProxy.SetRpc(r)
	return srvProxy, nil
}

func (r *rpcImpl) DestroyProxy(id ProxyId) error {
	return r.proxyMgr.Destroy(id)
}

func (r *rpcImpl) AddProxyCreator(uuid uint64, pc ProxyCreator) error {
	r.proxyMgr.addCreator(uuid, pc)
	return nil
}

func (r *rpcImpl) AddStubCreator(uuid uint64, bc StubCreator) error {
	r.serviceFactory[uuid] = bc
	return nil
}

// ============================= tool function ==============================

func (r *rpcImpl) onCall(trans transport.ITransport) error {
	//read trans header
	pkg := make([]byte, protocol.CallHeadSize)
	if mLen, err := trans.Read(pkg[:], protocol.CallHeadSize); mLen != protocol.CallHeadSize || err != nil {
		r.logger.Warn("[Rpc] parse rpc message error !")
		return errors.ErrIllegalProto
	}
	// read protocol header
	msgHeader := protocol.ReadCallHeader(pkg)
	if msgHeader == nil {
		r.logger.Warn("[Rpc] read req protocol head error !")
		return errors.ErrIllegalReq
	}

	mLen := int(msgHeader.Length) - protocol.CallHeadSize

	reqMsg := &protocol.RequestPackage{
		Header: msgHeader,
		Buffer: make([]byte, mLen),
	}

	rLen, err := trans.Read(reqMsg.Buffer[:], mLen)
	if err != nil || rLen != mLen {
		return errors.ErrIllegalProto
	}

	srvStub := r.stubMgr.Get(SvcUuid(msgHeader.ServiceUUID))
	if srvStub == nil {
		notFound(trans, msgHeader)
		return errors.NewServiceNotExist(msgHeader.ServiceUUID)
	}

	callUuid := r.stubMgr.GeneUuid()

	//create stub call
	stubCall := newStubCall(trans, reqMsg, callUuid)
	if stubCall == nil {
		r.logger.Warn("[Rpc] %d,%d,%d create stub call error!", reqMsg.Header.ServiceUUID, reqMsg.Header.MethodID, reqMsg.Header.CallID)
		return errors.ErrStubCallInvalid
	}

	err = srvStub.doCallService(trans, stubCall)
	if err != nil {
		r.logger.Warn("[Rpc] %d,%d,%d service all error !", reqMsg.Header.ServiceUUID, reqMsg.Header.MethodID, reqMsg.Header.CallID)
		return err
	}
	return nil
}

func (r *rpcImpl) onProxyCall(trans transport.ITransport) error {
	//read trans header
	pkg := make([]byte, protocol.ProxyCallHeadSize)
	if mLen, err := trans.Read(pkg[:], protocol.ProxyCallHeadSize); mLen != protocol.ProxyCallHeadSize || err != nil {
		r.logger.Warn("[Rpc] parse rpc proxy message error !")
		return errors.ErrIllegalProto
	}

	// read protocol header
	msgHeader := protocol.ReadProxyCallHeader(pkg)
	if msgHeader == nil {
		r.logger.Warn("[Rpc] read req protocol head error !")
		return errors.ErrIllegalReq
	}
	mLen := int(msgHeader.Length) - protocol.ProxyCallHeadSize
	reqMsg := &protocol.ProxyRequestPackage{
		Header: msgHeader,
		Buffer: make([]byte, mLen),
	}

	rLen, err := trans.Read(reqMsg.Buffer[:], mLen)
	if err != nil || rLen != mLen {
		return errors.ErrIllegalProto
	}

	srvStub := r.stubMgr.Get(SvcUuid(msgHeader.ServiceUUID))
	if srvStub == nil {
		//notFound(trans, msgHeader)
		notFoundReturnProxy(trans, msgHeader)
		return errors.NewServiceNotExist(msgHeader.ServiceUUID)
	}

	callUuid := r.stubMgr.GeneUuid()
	stubCall := newStubCallWithProxy(trans, reqMsg, callUuid)
	err = srvStub.doCallService(trans, stubCall)
	if err != nil {
		r.logger.Warn("[Rpc] %d,%d,%d service all error !", reqMsg.Header.ServiceUUID, reqMsg.Header.MethodID, reqMsg.Header.CallID)
		return err
	}
	return nil
}

func (r *rpcImpl) onReturn(trans transport.ITransport) error {
	pkg := make([]byte, protocol.RespHeadSize)
	if mLen, err := trans.Read(pkg, protocol.RespHeadSize); mLen != protocol.RespHeadSize || err != nil {
		r.logger.Warn("[Rpc] rpc return protocol error!")
		return errors.ErrIllegalProto
	}

	header := protocol.ReadRetHeader(pkg)
	if header == nil {
		return errors.ErrIllegalProto
	}

	//get resp data
	mLen := int(header.Length) - protocol.RespHeadSize
	resp := &protocol.ResponsePackage{
		Header: header,
		Buffer: make([]byte, mLen),
	}

	resLen, err := trans.Read(resp.Buffer, mLen)
	if err != nil || resLen != mLen {
		r.logger.Warn("[Rpc] %d proxy call protocol rev error !", header.CallID)
		return errors.ErrProxyInvalid
	}

	// 一定要 读取完整的消息结构才能返回错误，否则回出现消息错乱
	//get proxy call
	proxyCall := r.proxyCallMgr.Get(header.CallID)
	if proxyCall == nil {
		r.logger.Warn("[Rpc] %d proxy call not found", header.CallID)
		return errors.NewProxyNotFound(header.CallID)
	}

	srvProxy, err := r.proxyMgr.Get(ProxyId(proxyCall.ProxyId))
	if srvProxy != nil && err == nil {
		//flush proxy status
		switch header.ErrorCode {
		case protocol.IDL_SUCCESS:
			srvProxy.SetTargetID(header.ServerID)
		default:
			srvProxy.SetTargetID(common.InvalidStubId)
		}
	}

	proxyCall.SetErrorCode(header.ErrorCode)

	//always notify
	proxyCall.DoRet(resp.Buffer)
	return nil
}

func (r *rpcImpl) onProxyReturn(trans transport.ITransport) error {
	pkg := make([]byte, protocol.ProxyRetHeadSize)
	if mLen, err := trans.Read(pkg, protocol.ProxyRetHeadSize); mLen != protocol.ProxyRetHeadSize || err != nil {
		r.logger.Warn("[Rpc] rpc proxy protocol return error!")
		return errors.ErrIllegalProto
	}

	header := protocol.ReadProxyRetHeader(pkg)
	if header == nil {
		return errors.ErrIllegalProto
	}

	//get resp data
	mLen := int(header.Length) - protocol.ProxyRetHeadSize
	resp := &protocol.ProxyRespPackage{
		Header: header,
		Buffer: make([]byte, mLen),
	}

	resLen, err := trans.Read(resp.Buffer, mLen)
	if err != nil || resLen != mLen {
		r.logger.Warn("[Rpc] %d proxy call protocol rev error !", header.CallID)
		return errors.ErrProxyInvalid
	}

	//get proxy call
	proxyCall := r.proxyCallMgr.Get(header.CallID)
	if proxyCall == nil {
		r.logger.Warn("[Rpc] proxy call %d:%d not found", header.CallID, header.GlobalIndex)
		return errors.NewProxyNotFound(header.CallID)
	}

	srvProxy, err := r.proxyMgr.Get(ProxyId(proxyCall.ProxyId))
	if srvProxy != nil && err == nil {
		//flush proxy status
		switch header.ErrorCode {
		case protocol.IDL_SUCCESS:
			srvProxy.SetTargetID(header.ServerID)
		case protocol.IDL_SERVICE_NOT_FOUND:
			srvProxy.SetTargetID(common.InvalidStubId)
		case protocol.IDL_SERVICE_ERROR:
			srvProxy.SetTargetID(common.InvalidStubId)
		case protocol.IDL_RPC_TIME_OUT:
			srvProxy.SetTargetID(common.InvalidStubId)
		}
	}
	proxyCall.SetErrorCode(header.ErrorCode)
	//proxyCall.SetGlobalIndex(header.GlobalIndex)

	//always notify
	proxyCall.DoRet(resp.Buffer)

	return nil
}

// 外部连接超时
func (r *rpcImpl) onOutsideConnTimeout(trans transport.ITransport) error {
	// 读取解析协议
	pkg := make([]byte, protocol.TimeoutHeaderSize)
	if mLen, err := trans.Read(pkg, protocol.TimeoutHeaderSize); mLen != protocol.TimeoutHeaderSize || err != nil {
		r.logger.Warn("[Rpc] Protocol resolution fails because the protocol specifications are inconsistent or the packet is damaged! ")
		return errors.ErrIllegalProto
	}

	header := protocol.ReadTimeoutHeader(pkg)
	if header == nil {
		return errors.ErrIllegalProto
	}

	r.logger.Info("[Rpc] The external connection %d has been broken by interrupted signal", header.GlobalIndexId)
	// 查找proxy 设置下线状态，删除proxy
	return r.proxyMgr.closeOutsideProxy(header.GlobalIndexId)
}

// CallMethod proxy call helper
// rpc proxy call remote stub, create proxy call and wait for response
func callMethod(rpc *rpcImpl, pImpl IProxy, call *proxy.ProxyCall, methodId uint32, packData []byte) (buffer []byte, err error) {
	//pre-check
	if pImpl == nil {
		err = errors.ErrProxyInvalid
		rpc.logger.Error("[RpcProxy] 0,0,0 Invalid proxy for remote call")
		return
	}

	if !pImpl.IsConnected() {
		err = errors.ErrTransClose
		rpc.logger.Error("[RpcProxy] %s,%d,%d,%d IProxy connected has been closed, with method %s call", pImpl.GetSrvName(), pImpl.GetUUID(), pImpl.GetID(), pImpl.GetGlobalIndex(), pImpl.GetSignature(methodId))
		return
	}

	//get transport
	trans := pImpl.GetTransport()
	call.ReqData = packData
	err = trans.Send(call.ReqData)
	if err != nil {
		return nil, err
	}

	//do not wait for response while function is oneway
	if pImpl.IsOneWay(methodId) {
		return nil, nil
	}

	clicker := time.NewTimer(time.Duration(call.GetTimeOut()) * time.Millisecond)
	defer clicker.Stop()

	//wait for response and retry
	select {
	case buffer = <-call.Ch:
		if call.GetErrorCode() != protocol.IDL_RPC_TIME_OUT {
			// do nothing,rpc call successful or throw exception
			break
		}
		// time out error, retry
		buffer = retry(rpc, pImpl, call)
	case <-clicker.C:
		//add retry function code
		buffer = retry(rpc, pImpl, call)
	}

	errCode := call.GetErrorCode()
	switch errCode {
	case protocol.IDL_SUCCESS:
	case protocol.IDL_SERVICE_NOT_FOUND:
		rpc.logger.Warn("[Rpc] service %d method %s not found", pImpl.GetUUID(), pImpl.GetSignature(methodId))
		err = errors.ErrRpcNotFound
	case protocol.IDL_SERVICE_ERROR:
		rpc.logger.Warn("[Rpc] service %d method %s exec error", pImpl.GetUUID(), pImpl.GetSignature(methodId))
		err = errors.ErrRpcRet
	//case protocol.IDL_SERVICE_EXCEPTION:
	//	rpc.logger.Warn("[Rpc] service %d method %s exec panic", pImpl.GetUUID(), pImpl.GetSignature(methodId))
	//	err = errors.ErrRpcException
	case protocol.IDL_RPC_TIME_OUT:
		rpc.logger.Warn("[Rpc] service %d method %s exec timeout", pImpl.GetUUID(), pImpl.GetSignature(methodId))
		err = errors.ErrRpcTimeOut
	default:
	}
	return
}

// retry , send rpc call
func retry(rpc *rpcImpl, proxy IProxy, call *proxy.ProxyCall) (buffer []byte) {
	// check proxy is valid
	if !proxy.IsConnected() {
		uuid, id, name := proxy.GetUUID(), proxy.GetID(), proxy.GetSrvName()
		call.SetErrorCode(protocol.IDL_SERVICE_NOT_FOUND)
		rpc.logger.Warn("[IProxy] %d,%d,%s is invalid !", uuid, id, name)
		return
	}
	// 只要进入了这个函数，就一定是超时了
	call.SetErrorCode(protocol.IDL_RPC_TIME_OUT)
	rpc.logger.Warn("[IProxy] proxy call %d service %d:%q's method %q retry ", call.CallID, proxy.GetUUID(), proxy.GetSrvName(), proxy.GetSignature(call.MethodId))
	trans := proxy.GetTransport()
	for call.GetRetryTime() > 0 {
		call.DecRetryTime()
		err := trans.Send(call.ReqData)
		if err != nil {
			return
		}
		clicker := time.NewTimer(time.Duration(call.GetTimeOut()) * time.Millisecond)
		select {
		case buffer = <-call.Ch:
			// success
			errCode := call.GetErrorCode()
			if errCode == protocol.IDL_SUCCESS {
				clicker.Stop()
				return
			} else if errCode != protocol.IDL_RPC_TIME_OUT {
				clicker.Stop()
				return
			} //retry again
			clicker.Stop()
			rpc.logger.Warn("[IProxy] proxy call %d service %d:%q's method %q retry again ", call.CallID, proxy.GetUUID(), proxy.GetSrvName(), proxy.GetSignature(call.MethodId))
		case <-clicker.C:
			call.SetErrorCode(protocol.IDL_RPC_TIME_OUT)
			continue
		}
	}
	return
}
