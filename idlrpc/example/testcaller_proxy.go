// Generated by the go idl tools. DO NOT EDIT 2022-03-17 11:34:04
// source: TestCaller

package example

import (
	"github.com/CloudGuan/rpc-backend-go/idlrpc"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/example/pbdata"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/transport"
	"google.golang.org/protobuf/proto"
)

// TestCallerProxy define stub.ProxyStub
type TestCallerProxy struct {
	idlrpc.ProxyBase
}

func TestCallerProxyCreator(trans transport.ITransport) idlrpc.IProxy {
	if trans == nil {
		return nil
	}

	if trans.IsClose() {
		return nil
	}

	srvProxy := &TestCallerProxy{}
	srvProxy.SetTransport(trans)

	return srvProxy
}

// GetUUID define function
func (sp *TestCallerProxy) GetUUID() uint64 {
	return SrvUUID
}

func (sp *TestCallerProxy) GetSrvName() string {
	return SrvName
}

func (sp *TestCallerProxy) GetSignature(methid uint32) string {
	var sign string
	switch methid {
	case 1:
		sign = "SetInfo"
	case 2:
		sign = "GetInfo"
	}
	return sign
}

func (sp *TestCallerProxy) IsOneWay(methodid uint32) (isoneway bool) {
	switch methodid {
	case 1:
		isoneway = false
	case 2:
		isoneway = false
	default:
		isoneway = false
	}
	return
}
func (sp *TestCallerProxy) SetInfo(_1 string) (err error) {

	rpc := sp.GetRpc()
	if rpc == nil {
		//TODO add define error
		return
	}
	pbarg := &pbdata.TestCaller_SetInfoArgs{}
	pbarg.Arg1 = _1
	_, err = rpc.Call(sp, 1, 1000, 0, pbarg)
	if err != nil {
		return
	}

	return
}
func (sp *TestCallerProxy) GetInfo() (ret1 string, err error) {

	rpc := sp.GetRpc()
	if rpc == nil {
		//TODO add define error
		return
	}
	pbarg := &pbdata.TestCaller_GetInfoArgs{}
	respMsg, err := rpc.Call(sp, 2, 1000, 0, pbarg)
	if err != nil {
		return
	}

	//如果是oneway 的方法 不用检测返回值序列化，相当于传统的调用
	pbret := &pbdata.TestCaller_GetInfoRet{}
	err = proto.Unmarshal(respMsg, pbret)
	if err != nil {
		return
	}
	ret1 = pbret.Ret1

	return
}
