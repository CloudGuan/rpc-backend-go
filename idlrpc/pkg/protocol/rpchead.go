package protocol

import (
	"encoding/binary"

	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/common"
)

const (
	RpcInvalidMsg    uint32 = iota //0 无效类型
	RequestMsg                     // 1 调用请求
	ResponseMsg                    // 2 调用返回
	ProxyRequestMsg                // 3 proxy模式调用请求
	ProxyResponseMsg               // 4 proxy 模式请求回调
	NotRpcMsg                      // 5 非RPC协议
	RpcEventSub                    // 6 订阅事件
	RpcEventPub                    // 7 发布事件
	RpcEventCancel                 // 8 取消订阅
	RpcCallAlias                   // 9 调用声明
	RpcPing                        // 10 发送ping请求, 外部客户端到网关
	RpcPong                        // 11 返回pong请求, 网关返回外部客户端
	RpcTimeout                     // 12 ping pong 超时, 网关同步内部服务器
	RpcLoggedOut                   // 13 外部连接强制下线操作
	RpcProtocolMax                 // 10 最大有效协议范围
)

const (
	IDL_SUCCESS uint32 = iota + 1
	IDL_SERVICE_NOT_FOUND
	IDL_SERVICE_ERROR
	IDL_RPC_TIME_OUT
	IDL_RPC_LIMIT
	//IDL_SERVICE_EXCEPTION
)

var (
	RpcHeadSize         int
	CallHeadSize        int
	RespHeadSize        int
	ProxyCallHeadSize   int
	ProxyRetHeadSize    int
	SubHeaderSize       int
	PubHeaderSize       int
	CancelHeaderSize    int
	PingHeaderSize      int
	PongHeaderSize      int
	TimeoutHeaderSize   int
	LoggedOutHeaderSize int
)

// RpcMsgHeader 协议包头 协议类型 协议长度
type (
	GlobalIndexType uint32 //global index type
	RpcMsgHeader    struct {
		Length uint32 //整个包长，包含头
		Type   uint32 //调用 返回 非rpc 请求
	}
	RpcCallHeader struct {
		RpcMsgHeader
		ServiceUUID uint64 //服务UUID
		ServerID    uint32 //服务器实例ID
		CallID      uint32 //代理调用id
		MethodID    uint32 //方法id
	}

	RpcProxyCallHeader struct {
		RpcMsgHeader
		ServiceUUID uint64          //服务UUID
		ServerID    uint32          //服务器实例ID
		CallID      uint32          //代理调用id
		MethodID    uint32          //方法id
		GlobalIndex GlobalIndexType //代理节点标识
		OneWay      uint16          // 是否是one way节点
	}

	RpcCallRetHeader struct {
		RpcMsgHeader
		ServerID  uint32
		CallID    uint32
		ErrorCode uint32
	}

	RpcProxyCallRetHeader struct {
		RpcMsgHeader
		ServerID    uint32          //服务实例id
		CallID      uint32          //调用id对端赋值
		ErrorCode   uint32          //错误代码
		GlobalIndex GlobalIndexType //代理节点标识
	}

	// RpcSubHeader 订阅事件消息头
	RpcSubHeader struct {
		RpcMsgHeader
		SubId       [16]byte //订阅id
		ProxyId     uint32   //代理id
		ServiceUUID uint64   //服务uid
		ServiceID   uint32   //服务势力id
		NameLen     uint32   //事件名称长度
		DataLen     uint32   //数目长度
	}

	//RpcPubHeader 发布事件消息头
	RpcPubHeader struct {
		RpcMsgHeader
		SubId    [16]byte //订阅id
		ProxyId  uint32   //代理id
		ValueLen uint32   //附带数据长度
	}

	//RpcCancelSubHeader 取消订阅消息头
	RpcCancelSubHeader struct {
		RpcMsgHeader
		SubId [16]byte //订阅id
	}

	// RpcPingHeader 心跳包，client --> gateway
	RpcPingHeader struct {
		RpcMsgHeader
		PingId uint64 //ping id, 客户端第一次连接网关时候，赋值一次，可以用于定时器轮询
	}

	// RpcPongHeader 心跳包，gateway --> client
	RpcPongHeader struct {
		RpcMsgHeader
		PingId uint64 //ping id, 客户端第一次连接网关时候，赋值一次，可以用于定时器轮询
	}

	// RpcTimeoutHeader 断线时，网关通知内部服务，外部连接断开
	RpcTimeoutHeader struct {
		RpcMsgHeader
		GlobalIndexId GlobalIndexType // 网关唯一 id
	}

	// RpcLoggedOutHeader 强制用户下线
	RpcLoggedOutHeader struct {
		RpcMsgHeader
		GlobalIndexId GlobalIndexType
	}
)

func init() {
	RpcHeadSize = binary.Size(RpcMsgHeader{})
	CallHeadSize = binary.Size(RpcCallHeader{})
	RespHeadSize = binary.Size(RpcCallRetHeader{})
	ProxyCallHeadSize = binary.Size(RpcProxyCallHeader{})
	ProxyRetHeadSize = binary.Size(RpcProxyCallRetHeader{})
	SubHeaderSize = binary.Size(RpcSubHeader{})
	PubHeaderSize = binary.Size(RpcPubHeader{})
	CancelHeaderSize = binary.Size(RpcCancelSubHeader{})
	PingHeaderSize = binary.Size(RpcPingHeader{})
	PongHeaderSize = binary.Size(RpcPongHeader{})
	TimeoutHeaderSize = binary.Size(RpcTimeoutHeader{})
	LoggedOutHeaderSize = binary.Size(RpcLoggedOutHeader{})
}

func BuildRespHeader(resp *ResponsePackage, srvID uint32, callID uint32, errcode uint32) {
	if resp == nil {
		return
	}

	if resp.Header == nil {
		resp.Header = &RpcCallRetHeader{}
	}

	resp.Header.Type = ResponseMsg
	resp.Header.Length = uint32(RespHeadSize) + uint32(len(resp.Buffer))
	resp.Header.ServerID = srvID
	resp.Header.CallID = callID
	resp.Header.ErrorCode = errcode
}

func BuildProxyCallHeader(header *RpcCallHeader, globalIndex GlobalIndexType) *RpcProxyCallHeader {
	ph := &RpcProxyCallHeader{
		RpcMsgHeader{
			Length: header.Length - uint32(CallHeadSize) + uint32(ProxyCallHeadSize),
			Type:   ProxyRequestMsg,
		},
		header.ServiceUUID,
		header.ServerID,
		header.CallID,
		header.MethodID,
		globalIndex,
		0,
	}
	return ph
}

func BuildProxyRespHeader(resp *ProxyRespPackage, srvID uint32, callID uint32, errcode uint32, globalIndex GlobalIndexType) {
	if resp == nil {
		return
	}

	if resp.Header == nil {
		resp.Header = &RpcProxyCallRetHeader{}
	}

	resp.Header.Type = ProxyResponseMsg
	resp.Header.Length = uint32(ProxyRetHeadSize) + uint32(len(resp.Buffer))
	resp.Header.ServerID = srvID
	resp.Header.CallID = callID
	resp.Header.ErrorCode = errcode
	resp.Header.GlobalIndex = globalIndex
}

func BuildNotFound(req *RpcCallHeader) (resp *ResponsePackage) {
	if req == nil {
		return
	}
	resp = &ResponsePackage{
		Header: &RpcCallRetHeader{},
	}
	resp.Header.Type = ResponseMsg
	resp.Header.Length = uint32(RespHeadSize)
	resp.Header.ErrorCode = IDL_SERVICE_NOT_FOUND
	resp.Header.CallID = req.CallID
	resp.Header.ServerID = common.InvalidStubId
	return
}

func BuildProxyNotFound(req *RpcProxyCallHeader) (resp *ProxyRespPackage) {
	if req == nil {
		return
	}
	resp = &ProxyRespPackage{
		Header: &RpcProxyCallRetHeader{},
	}
	resp.Header.Type = ProxyResponseMsg
	resp.Header.Length = uint32(ProxyRetHeadSize)
	resp.Header.ErrorCode = IDL_SERVICE_NOT_FOUND
	resp.Header.CallID = req.CallID
	resp.Header.ServerID = common.InvalidStubId
	resp.Header.GlobalIndex = req.GlobalIndex
	return
}

// BuildException build run exception response
func BuildException(callID uint32, pkg []byte) (resp *ResponsePackage) {
	buflen := 0
	if pkg != nil {
		buflen = len(pkg)
	}
	resp = &ResponsePackage{
		&RpcCallRetHeader{
			RpcMsgHeader{
				uint32(RespHeadSize + buflen),
				ResponseMsg,
			},
			0,
			callID,
			IDL_SERVICE_ERROR,
		},
		pkg,
	}
	return resp
}

// BuildTimeOut TODO use static variables
func BuildTimeOut(callID uint32) (resp *ResponsePackage) {
	resp = &ResponsePackage{
		&RpcCallRetHeader{
			RpcMsgHeader{
				uint32(RespHeadSize),
				ResponseMsg,
			},
			0,
			callID,
			IDL_RPC_TIME_OUT,
		},
		nil,
	}

	return
}
