package common

//some global var

var (
	DefaultTimeOut       uint32 = 2000 // to second
	DefaultCallCache     uint32 = 128
	DefaultServiceWorker uint32 = 1
	DefaultServiceCache  uint32 = 8
	MaxRetryTime         int32  = 5 // max retry time TODO read from config
)

const InvalidStubId = 0
