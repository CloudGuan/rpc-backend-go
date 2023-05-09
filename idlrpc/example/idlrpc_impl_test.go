package example

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/CloudGuan/rpc-backend-go/idlrpc"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/example/pbdata"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/logger"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
	"google.golang.org/protobuf/proto"
)

type (
	testApp struct {
		rpc    idlrpc.IRpc
		cancel chan bool
		wg     sync.WaitGroup
	}
)

func (t *testApp) init() error {
	t.cancel = make(chan bool)
	t.wg = sync.WaitGroup{}

	t.rpc = idlrpc.CreateRpcFramework()
	err := t.rpc.Init(idlrpc.WithLogger(&logger.DefaultLogger{}), idlrpc.WithStackTrace(true))
	return err
}

func (t *testApp) start() {
	_ = t.rpc.Start()
	go func() {
		t.wg.Add(1)
		defer t.wg.Done()

		timer := time.NewTimer(10 * time.Millisecond)
		for {
			select {
			case <-timer.C:
				t.rpc.Tick()
				timer.Reset(10 * time.Millisecond)
			case <-t.cancel:
				return
			}
		}
	}()
}

func (t *testApp) stop() {
	t.cancel <- true
	t.wg.Wait()
	_ = t.rpc.ShutDown()
}

func TestRpcInit(t *testing.T) {
	app := testApp{}
	if err := app.init(); err != nil {
		t.Fatal(err)
	}
	app.start()
	time.Sleep(1 * time.Second)
	app.stop()
}

func TestAddService(t *testing.T) {
	app := testApp{}
	if err := app.init(); err != nil {
		t.Fatal(err)
	}
	if err := app.rpc.AddStubCreator(SrvUUID, TestCallerStubCreator); err != nil {
		t.Fatal(err)
	}
	app.start()
	if err := app.rpc.RegisterService(NewTestCaller()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	app.stop()
}

func TestSendMessage(t *testing.T) {
	app := testApp{}
	trans := NewTransportRing()
	caller := NewTestCaller()

	if err := app.init(); err != nil {
		t.Fatal(err)
	}
	if err := app.rpc.AddStubCreator(SrvUUID, TestCallerStubCreator); err != nil {
		t.Fatal(err)
	}
	app.start()
	if err := app.rpc.RegisterService(caller); err != nil {
		t.Fatal(err)
	}

	pbarg := &pbdata.TestCaller_SetInfoArgs{}
	pbarg.Arg1 = "hello"

	// parameters serialize data, message may be nil
	pkg, err := proto.Marshal(pbarg)
	if err != nil {
		return
	}
	// wrapper rpc call request
	reqPb := &protocol.RequestPackage{
		Header: &protocol.RpcCallHeader{
			RpcMsgHeader: protocol.RpcMsgHeader{
				Length: uint32(protocol.CallHeadSize + len(pkg)),
				Type:   protocol.RequestMsg,
			},
			ServiceUUID: SrvUUID,
			ServerID:    1,
			CallID:      1,
			MethodID:    1,
		},
		Buffer: pkg,
	}

	reqData, _ := protocol.PackReqMsg(reqPb)
	_, _ = trans.Write(reqData, len(reqData))

	_ = app.rpc.OnMessage(trans, context.Background())
	time.Sleep(2 * time.Second)
	if caller.name != "hello" {
		t.Error("call method1 failed! ")
	}
	app.stop()
}

func TestPeerPanic(t *testing.T) {
	app := testApp{}
	trans := NewTransportRing()
	caller := NewTestCaller()

	if err := app.init(); err != nil {
		t.Fatal(err)
	}
	if err := app.rpc.AddStubCreator(SrvUUID, TestCallerStubCreator); err != nil {
		t.Fatal(err)
	}
	app.start()
	if err := app.rpc.RegisterService(caller); err != nil {
		t.Fatal(err)
	}

	pbarg := &pbdata.TestCaller_GetInfoArgs{}

	// parameters serialize data, message may be nil
	pkg, err := proto.Marshal(pbarg)
	if err != nil {
		return
	}
	// wrapper rpc call request
	reqPb := &protocol.RequestPackage{
		Header: &protocol.RpcCallHeader{
			RpcMsgHeader: protocol.RpcMsgHeader{
				Length: uint32(protocol.CallHeadSize + len(pkg)),
				Type:   protocol.RequestMsg,
			},
			ServiceUUID: SrvUUID,
			ServerID:    1,
			CallID:      1,
			MethodID:    2,
		},
		Buffer: pkg,
	}

	reqData, _ := protocol.PackReqMsg(reqPb)
	_, _ = trans.Write(reqData, len(reqData))
	_ = app.rpc.OnMessage(trans, context.Background())
	time.Sleep(2 * time.Second)
	app.stop()
}

func TestRetryRpcCall(t *testing.T) {
	app := testApp{}
	trans := NewTransportRing()
	//caller := NewTestCaller()

	if err := app.init(); err != nil {
		t.Fatal(err)
	}
	if err := app.rpc.AddProxyCreator(SrvUUID, TestCallerProxyCreator); err != nil {
		t.Fatal(err)
	}
	app.start()

	// get proxy
	pInterface, err := app.rpc.GetServiceProxy(SrvUUID, trans)
	if err != nil {
		t.Fatal(err)
	}
	sp := pInterface.(*TestCallerProxy)
	if sp == nil {
		t.Fatal("get caller proxy error !!!")
	}

	_, err = sp.GetInfo()
	if err == nil {
		t.Fatalf("unexception error return")
	}

	app.stop()
}
