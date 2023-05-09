package errors

import (
	gerror "errors"
	"fmt"
)

type RpcErrorCode uint32

const (
	RpcNoErr RpcErrorCode = iota
	CommErr
	TransportClosed
	ProxyNotExist
	ProxyDisconnected
	ProxyTimeout
	ProxyCallNoFound
	StubcallInvalid
	ServiceShutdown
	ServiceNotExist
	ServiceHasExist
	MethodException
	ServicePanic
	SERVICE_NOT_FOUND
	FUNCTION_NOT_FOUND
)

var (
	ErrRpcNotFound  = gerror.New("rpc service or method not found")
	ErrRpcTimeOut   = gerror.New("rpc proxy call time out")
	ErrRpcRet       = gerror.New("rpc method execute return error")
	ErrRpcNotInit   = gerror.New("rpc framework not init")
	ErrRpcClosed    = gerror.New("rpc has been closed")
	ErrInvalidProto = gerror.New("invalid rpc protocol")
	ErrServiceInit  = gerror.New("initialize service error")
)

var (
	ErrTransClose      = &RpcError{TransportClosed, "Transport has been closed"}
	ErrProxyInvalid    = &RpcError{ProxyNotExist, "Proxy Is Invalid"}
	ErrStubCallInvalid = &RpcError{StubcallInvalid, "stub's call invalid"}
	ErrServicePanic    = &RpcError{ServicePanic, "service exec panic"}
	ErrIllegalReq      = &RpcError{errCode: CommErr, errStr: "invalid request message!"}
	ErrIllegalProto    = &RpcError{errCode: CommErr, errStr: "rpc protocol message buffer error !"}
)

type RpcError struct {
	errCode RpcErrorCode
	errStr  string
}

func (re *RpcError) Error() string {
	return re.errStr
}

func (re *RpcError) Code() RpcErrorCode {
	return re.errCode
}

func NewRpcError(code RpcErrorCode, format string, args ...interface{}) *RpcError {
	return &RpcError{
		errCode: code,
		errStr:  fmt.Sprintf(format, args),
	}
}

func NewProxyNotExit(proxyid uint32) *RpcError {
	return &RpcError{
		ProxyNotExist,
		fmt.Sprintf("Proxy %d Not Exist", proxyid),
	}
}

func NewProxyDisconnected(uuid uint64, id uint32, name string) *RpcError {
	return &RpcError{
		ProxyDisconnected,
		fmt.Sprintf("%s's proxy uuid:%d id:%d has been disconnected", name, uuid, id),
	}
}

func NewMethodExecError(service, method string) *RpcError {
	return &RpcError{
		MethodException,
		fmt.Sprintf("%::%s execute error !", service, method),
	}
}

func NewServiceNotExist(uuid uint64) *RpcError {
	return &RpcError{
		ServiceNotExist,
		fmt.Sprintf("service %d not exist", uuid),
	}
}

func NewProxyNotFound(callid uint32) *RpcError {
	return &RpcError{
		ProxyCallNoFound,
		fmt.Sprintf("proxy call %d not exits", callid),
	}
}

type RpcPanicInfo struct {
	Info interface{}
	Pkg  []byte
}

func (info RpcPanicInfo) String() string {
	return fmt.Sprintf("RPC CALL PANIC: %v", info.Info)
}
