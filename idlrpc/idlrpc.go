// Package idlrpc rpc-backend-go framework interface
package idlrpc

import (
	"context"

	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/transport"
	"google.golang.org/protobuf/proto"
)

const (
	RpcNotInit = iota
	RpcRunning
	RpcClosed
)

type (
	IRpc interface {
		Init(...Option) error
		// Start rpc framework,init data and start some goroutine
		Start() error
		// Tick Calls the main loop in the user master coroutine。
		Tick() error
		// ShutDown close rpc framework close
		ShutDown() error
		// Options options get rpc options
		Options() *Options
		// OnMessage deal network message
		OnMessage(trans transport.ITransport, ctx context.Context) error
		// OnProxyMessage trans proxy message to inside service
		OnProxyMessage(tran transport.ITransport, ph IProxyHandler) error
		// RegisterService register user impl service struct to framework
		RegisterService(service IService) error
		// Call service proxy call remote sync
		//Return resp unmarshalled proto buffer and exec result
		Call(proxyId IProxy, methodId, timeout uint32, retry int32, message proto.Message) ([]byte, error)
		// GetProxyFromPeer get proxy by stub call
		GetProxyFromPeer(ctx context.Context, uuid uint64) (IProxy, error)
		// GetServiceProxy get service proxy
		GetServiceProxy(uuid uint64, trans transport.ITransport) (IProxy, error)
		// AddProxyCreator add proxy creator
		AddProxyCreator(uuid uint64, pc ProxyCreator) error
		// AddStubCreator add stub creator
		AddStubCreator(uuid uint64, bc StubCreator) error
	}
)
