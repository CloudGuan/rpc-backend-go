package transport

import "github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"

const (
	TRANS_WORKING = iota
	TRANS_CLOSED
)

// ITransport rpc-backend-go, message binary transporter interface
// recv protocol binary from network, application consume it
// recv protocol binary from application, network consume it
type ITransport interface {
	Write(pkg []byte, length int) (int, error) //往缓冲区写入
	Read(pkg []byte, length int) (int, error)  //从缓冲区读取，并且移除数据
	Peek(length int) ([]byte, int, error)      //获取头部的定长数据 但是不移除他们
	Send(pkg []byte) error                     //发送数据包 多协程调用
	Close()                                    //关闭
	Size() uint32                              //返回读队列长度
	IsClose() bool                             //查询状态
	GetID() uint32                             //获取Trans ID
	SetID(transID uint32)                      //设置Trans ID
	LocalAddr() string                         // LocalAddr returns the local network address.
	RemoteAddr() string                        // RemoteAddr returns the remote network address.
	GlobalIndex() protocol.GlobalIndexType     // outside global index id
	Heartbeat() error                          // 触发一次心跳逻辑
}
