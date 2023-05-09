package idlrpc

import (
	"context"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/common"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/errors"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/log"
	"sync"
	"sync/atomic"
)

// SvcUuid service uuid type
type SvcUuid uint64

// ServiceCache service storage struct
type ServiceCache map[SvcUuid]*stubWrapper

// StubManager stub manager, manager registered service
type StubManager struct {
	stubCallId CallUuid     //stub call uuid
	svcMaps    ServiceCache //service
	rwlock     sync.RWMutex //read write lock
	logger     log.ILogger  //logger
}

func newStubManager() *StubManager {
	return &StubManager{
		1,
		make(ServiceCache, common.DefaultServiceCache),
		sync.RWMutex{},
		nil,
	}
}

func (m *StubManager) Init(logger log.ILogger) {
	m.logger = logger
}

func (m *StubManager) GeneUuid() CallUuid {
	return CallUuid(atomic.AddUint32((*uint32)(&m.stubCallId), 1))
}

func (m *StubManager) Add(ctx context.Context, impl IStub) (err error) {
	if impl == nil {
		//In theory, it will not enter this branch forever
		err = errors.NewRpcError(errors.CommErr, "service impl is nil!")
		return
	}

	//lock
	m.rwlock.Lock()
	defer m.rwlock.Unlock()

	//check repeated add
	_, ok := m.svcMaps[impl.GetUUID()]
	if ok {
		err = errors.NewRpcError(errors.ServiceHasExist, "service %s has exits in this programe", impl.GetUUID())
		m.logger.Error("[Service] %s,%d,0 service has been added to this programe", impl.GetServiceName(), impl.GetUUID())
		return
	}

	//create stub instance
	sb := newStubWrapper(impl, m.logger)
	if sb == nil {
		err = errors.NewRpcError(errors.CommErr, "service %s create instance error", impl.GetServiceName())
		m.logger.Error("[Service] %s,%d,0 create service instance error!", impl.GetServiceName(), impl.GetUUID())
		return
	}
	//call init funciton
	err = sb.init(ctx)
	if err != nil {
		return
	}
	//add to map
	m.svcMaps[impl.GetUUID()] = sb
	//start loop
	sb.start()

	m.logger.Info("[Service] %s, %d, 0 service add to rpc framework successful !", impl.GetServiceName(), impl.GetUUID())
	return
}

func (m *StubManager) Tick() {
	// read lock
	m.rwlock.RLock()
	// defer unlock
	defer m.rwlock.RUnlock()

	for _, v := range m.svcMaps {
		if v.isValid() {
			v.tick()
		}
	}
}

func (m *StubManager) Get(uuid SvcUuid) *stubWrapper {
	//read lock
	m.rwlock.RLock()
	defer m.rwlock.RUnlock()

	v, ok := m.svcMaps[uuid]
	if !ok {
		return nil
	}

	//check close status
	if !v.isValid() {
		m.logger.Warn("[Service] %s, %d,0  service has been closed !", v.srvImp.GetServiceName(), v.srvImp.GetUUID())
		return nil
	}
	return v
}

func (m *StubManager) UnInit() {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()

	for _, v := range m.svcMaps {
		v.close()
	}

	m.svcMaps = nil
}
