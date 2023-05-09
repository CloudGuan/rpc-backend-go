package protocol

import "sync"

const (
	PACKAGE_ERROR = iota
	PACKAGE_FULL
	PACKAGE_LESS
)

var (
	curprotocol Protocol
	once        sync.Once
)

type Protocol interface {
	ReadHeader(pkg []byte, header *RpcMsgHeader) bool
	ParsePlatoHeader(pkg []byte, header interface{}) bool
	ParseReqMsg(pkg []byte, header *RpcCallHeader) bool
	ParseProxyReqMsg([]byte, *RpcProxyCallHeader) bool
	ParseRespMsg(pkg []byte, header *RpcCallRetHeader) bool
	ParseProxyRespMsg([]byte, *RpcProxyCallRetHeader) bool
	ParseSubMsg([]byte, *RpcSubHeader) bool
	ParsePubMsg([]byte, *RpcPubHeader) bool
	ParseCancelMsg([]byte, *RpcCancelSubHeader) bool
	PackPlatoMsg(interface{}, []byte, int) ([]byte, int)
	PackRespMsg(resp *ResponsePackage) ([]byte, int)
	PackReqMsg(req *RequestPackage) ([]byte, int)
	PackProxyReqMsg(req *ProxyRequestPackage) ([]byte, int)
	PackProxyRespMsg(resp *ProxyRespPackage) ([]byte, int)
	PackSubMsg(header *RpcSubPackage) ([]byte, int)
	PackPubMsg(header *RpcPubPackage) ([]byte, int)
	PackCancelMsg(header *RpcCancelPackage) ([]byte, int)
}

// 初始化，对接多种协议格式可以
func init() {
	once.Do(func() {
		curprotocol = &binaryProtocol{}
	})
}

func SetProtocol(cusProto Protocol) {
	curprotocol = cusProto
}

func ReadHeader(pkg []byte) *RpcMsgHeader {
	if curprotocol == nil {
		return nil
	}

	header := &RpcMsgHeader{}
	if curprotocol.ReadHeader(pkg, header) == false {
		return nil
	}

	return header
}

func ReadCallHeader(pkg []byte) *RpcCallHeader {
	if curprotocol == nil {
		return nil
	}

	header := &RpcCallHeader{}
	if curprotocol.ParseReqMsg(pkg, header) == false {
		return nil
	}
	return header
}

func ReadProxyCallHeader(pkg []byte) *RpcProxyCallHeader {
	if curprotocol == nil {
		return nil
	}
	header := &RpcProxyCallHeader{}
	//sif cur protocol.Parse
	if curprotocol.ParseProxyReqMsg(pkg, header) == false {
		return nil
	}
	return header
}

func ReadRetHeader(pkg []byte) *RpcCallRetHeader {
	if curprotocol == nil {
		return nil
	}

	header := &RpcCallRetHeader{}
	if curprotocol.ParseRespMsg(pkg, header) == false {
		return nil
	}
	return header
}

func ReadProxyRetHeader(pkg []byte) *RpcProxyCallRetHeader {
	if curprotocol == nil {
		return nil
	}

	header := &RpcProxyCallRetHeader{}
	if curprotocol.ParseProxyRespMsg(pkg, header) == false {
		return nil
	}
	return header
}

func ReadPingHeader(pkg []byte) *RpcPingHeader {
	if curprotocol == nil {
		return nil
	}
	// check size
	if PingHeaderSize > len(pkg) {
		return nil
	}

	header := &RpcPingHeader{}
	if curprotocol.ParsePlatoHeader(pkg, header) == false {
		return nil
	}

	return header
}

func ReadTimeoutHeader(pkg []byte) *RpcTimeoutHeader {
	if curprotocol == nil {
		return nil
	}
	if TimeoutHeaderSize > len(pkg) {
		return nil
	}

	header := &RpcTimeoutHeader{}
	if curprotocol.ParsePlatoHeader(pkg, header) == false {
		return nil
	}

	return header
}

func ReadLoggedOutHeader(pkg []byte) *RpcLoggedOutHeader {
	if curprotocol == nil {
		return nil
	}

	if LoggedOutHeaderSize > len(pkg) {
		return nil
	}

	header := &RpcLoggedOutHeader{}

	if curprotocol.ParsePlatoHeader(pkg, header) == false {
		return nil
	}

	return header
}

func PackRespMsg(resp *ResponsePackage) ([]byte, int) {
	return curprotocol.PackRespMsg(resp)
}

func PackProxyRespMsg(resp *ProxyRespPackage) ([]byte, int) {
	return curprotocol.PackProxyRespMsg(resp)
}

func PackReqMsg(req *RequestPackage) ([]byte, int) {
	return curprotocol.PackReqMsg(req)
}

func PackProxyReqMsg(req *ProxyRequestPackage) ([]byte, int) {
	return curprotocol.PackProxyReqMsg(req)
}

func PackPingMsg(resp *RpcPingPackage) ([]byte, int) {
	return curprotocol.PackPlatoMsg(resp.Header, nil, int(resp.Header.Length))
}

func PackPongMsg(resp *RpcPongPackage) ([]byte, int) {
	return curprotocol.PackPlatoMsg(resp.Header, nil, int(resp.Header.Length))
}

func PackTimeMsg(resp *RpcTimeoutPackage) ([]byte, int) {
	return curprotocol.PackPlatoMsg(resp.Header, nil, int(resp.Header.Length))
}

func PackLoggedOutMsg(resp *RpcLoggedOutPackage) ([]byte, int) {
	return curprotocol.PackPlatoMsg(resp.Header, nil, int(resp.Header.Length))
}
