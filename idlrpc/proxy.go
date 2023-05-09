package idlrpc

import (
	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/proxy"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/transport"
	"sync"
	"sync/atomic"
)

const (
	ProxyConnected = iota
	ProxyDisconnected
)

type ProxyId proxy.ProxyUuid

// ProxyCreator proxy creator, one client may connect to differently service
type ProxyCreator func(transport.ITransport) IProxy

// ProxyBase proxy common struct impl
type ProxyBase struct {
	id          ProxyId                  // proxy instance id, only set by create
	targetId    uint32                   // last response service id
	status      int32                    // status of proxy's, maybe ticked out or lost heartbeat
	globalIndex protocol.GlobalIndexType // real proxy id
	trans       transport.ITransport     // transport ref
	rpc         IRpc                     // rpc framework instance
	rw          sync.Mutex
}

// SetTransport set network message transport
// called by creator
func (base *ProxyBase) SetTransport(trans transport.ITransport) {
	base.trans = trans
}

func (base *ProxyBase) SetID(id ProxyId) {
	base.id = id
}

func (base *ProxyBase) GetID() ProxyId {
	return base.id
}

func (base *ProxyBase) SetTargetID(id uint32) {
	base.targetId = id
}

func (base *ProxyBase) GetTargetID() uint32 {
	return base.targetId
}

func (base *ProxyBase) GetTransport() transport.ITransport {
	return base.trans
}

func (base *ProxyBase) GetGlobalIndex() protocol.GlobalIndexType {
	return base.globalIndex
}

func (base *ProxyBase) SetGlobalIndex(index protocol.GlobalIndexType) {
	base.globalIndex = index
}

func (base *ProxyBase) IsConnected() bool {
	if base == nil {
		return false
	}

	if base.trans == nil {
		return false
	}

	// 是否是有效连接，新增踢人和心跳协议
	if atomic.LoadInt32(&base.status) != ProxyConnected {
		return false
	}

	return !base.trans.IsClose()
}

func (base *ProxyBase) SetRpc(r IRpc) {
	base.rw.Lock()
	defer base.rw.Unlock()
	base.rpc = r
}

func (base *ProxyBase) GetRpc() IRpc {
	return base.rpc
}

func (base *ProxyBase) close() {
	atomic.StoreInt32(&base.status, ProxyDisconnected)
}

// IProxy  user's idl proxy interface, using as client to call remote function;
// rpc framework set transport to it
type IProxy interface {

	// GetUUID service uuid
	// @detail get service uuid id which generated by rpc-repo
	GetUUID() uint64

	// GetSignature
	// get method human-readable name
	GetSignature(uint32) string

	// GetID
	//get proxy instance ID, priority delivery service instance
	GetID() ProxyId

	// SetID
	// set proxy instance ID
	SetID(ProxyId)

	// SetTargetID
	// last return service id, maybe some service box need
	SetTargetID(uint32)

	// GetTargetID
	// return last return service id
	GetTargetID() uint32

	// GetGlobalIndex
	// return proxy global index
	GetGlobalIndex() protocol.GlobalIndexType

	// SetGlobalIndex
	// set proxy global index
	SetGlobalIndex(indexType protocol.GlobalIndexType)

	// IsOneWay oneway
	//@detail is one way function which not wait for return
	//@param method id
	IsOneWay(uint32) bool

	// GetSrvName
	//@detail get proxy name
	GetSrvName() string

	// SetTransport
	// @detail set transport
	SetTransport(transport transport.ITransport)

	// GetTransport
	// return transport
	GetTransport() transport.ITransport

	// IsConnected
	// check transport valid state
	IsConnected() bool

	// SetRpc set rpc instance to proxy
	SetRpc(IRpc)

	// GetRpc get rpc instance from proxy
	GetRpc() IRpc

	// close will close this proxy without closing the connection of network
	close()
}