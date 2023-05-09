package idlrpc

import "context"

type ServiceStatus uint32

const (
	SERVICE_RESOLVED ServiceStatus = iota + 1 //ready for servicing
	SERVICE_UPDATING                          //stop receive message, wait for update
)

type (
	//IStub rpc stub of service's side
	IStub interface {
		//GetUUID service uuid, generate by rpc-repo
		GetUUID() SvcUuid
		//GetServiceName service name
		GetServiceName() string
		//GetSignature method human-readable name
		GetSignature(methodId uint32) string
		//GetMutipleNum service goroutine num, read from idl keyword multiple
		GetMutipleNum() uint32
		//IsOneWay one way function, not send response
		IsOneWay(methodId uint32) bool
		// Call will trigger service's method
		// return encode response buffer data and errors
		Call(context.Context, uint32, []byte) ([]byte, error)
		//OnAfterFork invoke by framework after spawned
		//init service data or load db data in this function
		OnAfterFork(ctx context.Context) bool
		//OnBeforeDestroy invoke by framework before uninitialized
		//clean service data in this function
		OnBeforeDestroy() bool
		//OnTick tick by service manager in logic tick
		OnTick() bool
		//GetStatus get service status
		//TODO: for hot update
		GetStatus() ServiceStatus
		//SetStatus set service status
		SetStatus(status ServiceStatus)
	}
)

//StubCreator stub factory
type StubCreator func(v interface{}) IStub
