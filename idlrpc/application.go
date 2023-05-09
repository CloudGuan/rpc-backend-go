package idlrpc

import (
	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/proxy"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/transport"
)

const (
	InvalidGlobalIndex = 0
)

func notFound(trans transport.ITransport, req *protocol.RpcCallHeader) {
	resp := protocol.BuildNotFound(req)
	if resp == nil {
		return
	}
	resppkg, pkglen := protocol.PackRespMsg(resp)
	if resppkg == nil || pkglen == 0 {
		//TODO 添加序列化错误
		return
	}
	err := trans.Send(resppkg)
	if err != nil {
		return
	}
}

func notFoundReturnProxy(trans transport.ITransport, proxyReq *protocol.RpcProxyCallHeader) {
	resp := protocol.BuildProxyNotFound(proxyReq)
	respPkg, pkgLen := protocol.PackProxyRespMsg(resp)
	if respPkg == nil || pkgLen == 0 {
		//TODO 添加序列化错误
		return
	}
	err := trans.Send(respPkg)
	if err != nil {
		return
	}
}

func CreateRpcFramework() IRpc {
	return &rpcImpl{
		proxyMgr:       newProxyManager(),
		proxyCallMgr:   proxy.NewCallManager(),
		stubMgr:        newStubManager(),
		serviceFactory: make(stubFactoryMap),
		logger:         nil,
		status:         RpcNotInit,
	}
}
