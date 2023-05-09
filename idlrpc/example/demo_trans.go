package example

import (
	"errors"
	"sync/atomic"

	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
)

// TransportRing 双队列，一读一写
type TransportRing struct {
	transID    uint32
	isclose    uint32
	recvbuffer []byte
	sendchan   chan []byte
}

func NewTransportRing() *TransportRing {
	return &TransportRing{
		//recvchan: make(chan []byte, 1024*1024),
		sendchan: make(chan []byte, 1024*1024),
		isclose:  0,
	}
}

func (trans *TransportRing) LocalAddr() string {
	return ""
}

func (trans *TransportRing) RemoteAddr() string {
	return ""
}

func (trans *TransportRing) GlobalIndex() protocol.GlobalIndexType {
	return 0
}

func (trans *TransportRing) Write(pkg []byte, length int) (int, error) {
	trans.recvbuffer = append(trans.recvbuffer, pkg...)
	return len(trans.recvbuffer), nil
}

func (trans *TransportRing) Read(pkg []byte, length int) (int, error) {
	if len(trans.recvbuffer) < length {
		return 0, nil
	}
	real := copy(pkg, trans.recvbuffer[:length])
	if real != length {
		return 0, nil
	}
	trans.recvbuffer = trans.recvbuffer[length:]
	return length, nil
}
func (trans *TransportRing) Peek(length int) ([]byte, int, error) {
	if trans == nil {
		return nil, 0, errors.New("Messages Trans Error !")
	}

	if len(trans.recvbuffer) < length {
		return nil, len(trans.recvbuffer), nil
	}
	return trans.recvbuffer[:length], length, nil
}

func (trans *TransportRing) Send(pkg []byte) error {
	if atomic.LoadUint32(&trans.isclose) == 0 {
		trans.sendchan <- pkg
	}
	return nil
}

func (trans *TransportRing) PopSend() []byte {
	return <-trans.sendchan
}

func (trans *TransportRing) Size() uint32 {
	return uint32(len(trans.recvbuffer))
}

func (trans *TransportRing) Close() {
	atomic.StoreUint32(&trans.isclose, 1)
}
func (trans *TransportRing) IsClose() bool {
	return atomic.LoadUint32(&trans.isclose) == 1
}

func (trans *TransportRing) GetID() uint32 {
	return trans.transID
}

func (trans *TransportRing) SetID(id uint32) {
	trans.transID = id
}

func (trans *TransportRing) Heartbeat() error {
	return nil
}
