package proxy

import (
	common2 "github.com/CloudGuan/rpc-backend-go/idlrpc/internal/common"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/logger"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/errors"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
	"sync"
	"sync/atomic"
)

type CallMap map[uint32]*ProxyCall

// ProxyCallManager manager proxy call for multi goroutine
type ProxyCallManager struct {
	callMap CallMap
	callID  uint32
	rwMutex sync.RWMutex //loc
}

func NewCallManager() *ProxyCallManager {
	return &ProxyCallManager{
		make(map[uint32]*ProxyCall),
		1,
		sync.RWMutex{},
	}
}

func (pcm *ProxyCallManager) GenCallID() uint32 {
	return atomic.AddUint32(&pcm.callID, 1)
}

func (pcm *ProxyCallManager) CreateProxyCall(proxyId ProxyUuid, timeOut uint32, retryTime int32, globalIndex protocol.GlobalIndexType) *ProxyCall {
	if retryTime > common2.MaxRetryTime {
		retryTime = common2.MaxRetryTime
	}
	pc := &ProxyCall{
		ProxyId:     proxyId,
		CallID:      pcm.GenCallID(),
		timeOut:     timeOut,
		retryTime:   retryTime,
		globalIndex: globalIndex,
		Ch:          make(chan []byte, 1),
	}
	return pc
}

func (pcm *ProxyCallManager) Add(pc *ProxyCall) error {
	// lock map
	pcm.rwMutex.Lock()
	defer pcm.rwMutex.Unlock()
	//check proxy call id repetition
	_, ok := pcm.callMap[pc.CallID]
	if ok {
		logger.Error("[IProxy] %d,0,0 proxy call has existed ", pc.CallID)
		return errors.NewRpcError(errors.CommErr, "proxy call %d is exist!!", pc.CallID)
	}

	pcm.callMap[pc.CallID] = pc
	return nil
}

func (pcm *ProxyCallManager) Get(callId uint32) *ProxyCall {
	//lock
	pcm.rwMutex.RLock()
	defer pcm.rwMutex.RUnlock()

	pc, ok := pcm.callMap[callId]
	if !ok {
		logger.Warn("[IProxy] %d,0,0 is not exist", callId)
		return nil
	}
	return pc
}

func (pcm *ProxyCallManager) Destroy(callId uint32) {
	// lock map
	pcm.rwMutex.Lock()
	defer pcm.rwMutex.Unlock()

	delete(pcm.callMap, callId)
}
