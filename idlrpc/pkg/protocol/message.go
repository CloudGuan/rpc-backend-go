package protocol

type (
	RequestPackage struct {
		Header *RpcCallHeader //请求协议头
		Buffer []byte         //协议提二进制数据， 可以对接pb 或者自己的结构体
	}
	ProxyRequestPackage struct {
		Header *RpcProxyCallHeader
		Buffer []byte
	}
	ResponsePackage struct {
		Header *RpcCallRetHeader //回包结构体
		Buffer []byte            //协议体二进制数据， 可以对接pb 或者自己的结构体
	}
	ProxyRespPackage struct {
		Header *RpcProxyCallRetHeader
		Buffer []byte
	}

	RpcSubPackage struct {
		Header *RpcSubHeader
		Buffer []byte
	}

	RpcPubPackage struct {
		Header *RpcPubHeader
		Buffer []byte
	}

	RpcCancelPackage struct {
		Header *RpcCancelSubHeader
		Buffer []byte
	}

	RpcPingPackage struct {
		Header *RpcPingHeader
	}

	RpcPongPackage struct {
		Header *RpcPongHeader
	}

	RpcTimeoutPackage struct {
		Header *RpcTimeoutHeader
	}

	RpcLoggedOutPackage struct {
		Header *RpcLoggedOutHeader
	}
)
