// Generated by the go idl tools. DO NOT EDIT 2022-03-17 11:27:42
// source: box_example
package example

import "context"

type TestCallerImpl struct {
	name string
}

func NewTestCaller() *TestCallerImpl {
	service := &TestCallerImpl{}
	return service
}

func (sp *TestCallerImpl) GetUUID() uint64 {
	return 8590810067174448481
}

func (sp *TestCallerImpl) GetNickName() string {
	return "8590810067174448481"
}

func (sp *TestCallerImpl) OnAfterFork(ctx context.Context) bool {
	//add this service to framework, do not call rpc method in this function
	return true
}

func (sp *TestCallerImpl) OnTick() bool {
	//tick function in main goroutine
	return true
}

func (sp *TestCallerImpl) OnBeforeDestroy() bool {
	//tick function in main goroutine
	return true
}
func (sp *TestCallerImpl) SetInfo(ctx context.Context, _1 string) (err error) {
	sp.name = _1
	return
}

func (sp *TestCallerImpl) GetInfo(ctx context.Context) (ret1 string, err error) {
	panic("test case panic!")
}