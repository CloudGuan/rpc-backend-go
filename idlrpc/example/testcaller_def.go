// Generated by the go idl tools. DO NOT EDIT 2022-03-17 11:34:04
// source: TestCaller
package example

import (
	"context"

	"github.com/CloudGuan/rpc-backend-go/idlrpc"
)

const (
	SrvUUID = 8590810067174448481
	SrvName = "TestCaller"
)

type ITestCaller interface {
	idlrpc.IService
	SetInfo(context.Context, string) error
	GetInfo(context.Context) (string, error)
}
