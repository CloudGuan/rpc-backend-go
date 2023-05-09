package proxy

import (
	common2 "github.com/CloudGuan/rpc-backend-go/idlrpc/internal/common"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
	"sync/atomic"
)

type ProxyUuid uint32

// ProxyCall  rpc-proxy call information of remote call description
type ProxyCall struct {
	CallID      uint32                   //proxy call uuid, do not modify
	ProxyId     ProxyUuid                //proxy instance id
	ReqData     []byte                   //serizaled data of rpc request
	Ch          chan []byte              //response notify channel
	errCode     uint32                   //remote call error code, IDL_SUCCESS
	timeOut     uint32                   //method timeout
	globalIndex protocol.GlobalIndexType // proxy call global index
	retryTime   int32                    //left retry time
	MethodId    uint32                   // method id 提供给日志使用
}

// DecRetryTime decrease retry time, read & write in worker goroutine
func (pc *ProxyCall) DecRetryTime() {
	pc.retryTime--
}

// GetRetryTime get left retry time
func (pc *ProxyCall) GetRetryTime() int32 {
	return pc.retryTime
}

func (pc *ProxyCall) SetErrorCode(errCode uint32) {
	atomic.StoreUint32(&pc.errCode, errCode)
}

func (pc ProxyCall) GetErrorCode() uint32 {
	return atomic.LoadUint32(&pc.errCode)
}

func (pc *ProxyCall) DoRet(body []byte) {
	pc.Ch <- body
}

func (pc *ProxyCall) GlobalIndex() protocol.GlobalIndexType {
	return pc.globalIndex
}

func (pc *ProxyCall) SetGlobalIndex(index protocol.GlobalIndexType) {
	pc.globalIndex = index
}

// GetTimeOut timeout millisecond
func (pc *ProxyCall) GetTimeOut() uint32 {
	if pc.timeOut == 0 {
		return common2.DefaultTimeOut
	}
	return pc.timeOut
}
