package idlrpc

import (
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/transport"
)

// CallUuid  stub call uuid generate by manager
type CallUuid uint32

// StubCall remote call data
type StubCall struct {
	uuid      CallUuid                 //uuid, generate by manager
	srvUuid   uint64                   // service uuid
	srvInstID uint32                   // service instance id
	callID    uint32                   // proxy call id
	methodID  uint32                   // method id
	globalID  protocol.GlobalIndexType // global index id for transport
	oneWay    uint16                   // is one way method
	buffer    []byte                   // serialize body
	trans     transport.ITransport     //remote client socket channel
}

// newStubCall create stub call by manager
func newStubCall(trans transport.ITransport, req *protocol.RequestPackage, uuid CallUuid) *StubCall {
	return &StubCall{
		uuid:      uuid,
		srvUuid:   req.Header.ServiceUUID,
		srvInstID: req.Header.ServerID,
		callID:    req.Header.CallID,
		methodID:  req.Header.MethodID,
		buffer:    req.Buffer,
		trans:     trans,
	}
}

func newStubCallWithProxy(trans transport.ITransport, req *protocol.ProxyRequestPackage, uuid CallUuid) *StubCall {
	return &StubCall{
		uuid:      uuid,
		srvUuid:   req.Header.ServiceUUID,
		srvInstID: req.Header.ServerID,
		callID:    req.Header.CallID,
		methodID:  req.Header.MethodID,
		globalID:  req.Header.GlobalIndex,
		oneWay:    req.Header.OneWay,
		buffer:    req.Buffer,
		trans:     trans,
	}
}

// MethodID rpc method call id
func (sc *StubCall) MethodID() uint32 {
	return sc.methodID
}

// CallID rpc remote proxyCall call id
func (sc *StubCall) CallID() uint32 {
	return sc.callID
}

// GetUUID stub call uuid
func (sc *StubCall) GetUUID() CallUuid {
	return sc.uuid
}

func (sc *StubCall) GetServiceUUID() uint64 {
	return sc.srvUuid
}

func (sc *StubCall) GlobalIndex() protocol.GlobalIndexType {
	return sc.globalID
}

func (sc *StubCall) GetBody() []byte {
	return sc.buffer
}

func (sc *StubCall) doRet(msg []byte) error {
	if sc.trans.IsClose() {
		//TODO add common error
		//log2.Error("[Service] %s,%d,0 service method %s call raise panic %v, but transport has been closed", s.srvImp.GetServiceName(), s.srvImp.GetUUID(), stubCall.MethodID(), r)
		return nil
	}
	err := sc.trans.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
