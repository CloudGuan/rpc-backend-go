package idlrpc

type (
	//SdkCreateHandle rpc sdk creator, return sdk instance
	SdkCreateHandle func(...string) (ISDK, error)
	// PackageInfo service package information
	PackageInfo struct {
		ServiceUUID uint64          //service uuid
		Creator     SdkCreateHandle //Sdk Creator handle
	}

	// ISDK sdk interface
	ISDK interface {
		GetUuid() uint64
		GetNickName() string
		IsProxy() bool
		Register(IRpc) error
	}
)
