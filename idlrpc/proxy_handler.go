package idlrpc

import (
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/transport"
)

type (
	IProxyHandler interface {
		OnRelay(trans transport.ITransport, header *protocol.RpcMsgHeader) error
	}
)
