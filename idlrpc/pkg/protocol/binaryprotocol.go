package protocol

import (
	"bytes"
	"encoding/binary"
)

type binaryProtocol struct {
}

func (bp *binaryProtocol) ReadHeader(pkg []byte, header *RpcMsgHeader) bool {
	if header == nil {
		return false
	}

	hsize := binary.Size(header)
	if hsize == -1 {
		return false
	}

	if len(pkg) < hsize {
		return false
	}

	header.Length = binary.BigEndian.Uint32(pkg[0:4])
	header.Type = binary.BigEndian.Uint32(pkg[4:8])
	return true
}

func (bp *binaryProtocol) ParsePlatoHeader(pkg []byte, header interface{}) bool {
	reader := bytes.NewReader(pkg)
	if err := binary.Read(reader, binary.BigEndian, header); err != nil {
		return false
	}
	return true
}

// ParseReqMsg parse quest package
func (bp *binaryProtocol) ParseReqMsg(pkg []byte, header *RpcCallHeader) bool {
	if header == nil {
		return false
	}

	if CallHeadSize > len(pkg) {
		return false
	}

	reader := bytes.NewBuffer(pkg)
	err := binary.Read(reader, binary.BigEndian, header)
	if err != nil {
		//TODO record log
		return false
	}

	return true
}

func (bp *binaryProtocol) ParseProxyReqMsg(pkg []byte, header *RpcProxyCallHeader) bool {
	if header == nil {
		return false
	}

	if ProxyCallHeadSize > len(pkg) {
		return false
	}

	reader := bytes.NewBuffer(pkg)
	err := binary.Read(reader, binary.BigEndian, header)
	if err != nil {
		//TODO record log
		return false
	}
	return true
}

func (bp *binaryProtocol) ParseRespMsg(pkg []byte, header *RpcCallRetHeader) bool {
	if header == nil {
		return false
	}

	if RespHeadSize > len(pkg) {
		return false
	}

	reader := bytes.NewBuffer(pkg)
	err := binary.Read(reader, binary.BigEndian, header)
	if err != nil {
		//TODO record log
		return false
	}

	return true
}

func (bp *binaryProtocol) ParseProxyRespMsg(pkg []byte, header *RpcProxyCallRetHeader) bool {
	if header == nil {
		return false
	}

	//hsize := binary.Size(header)
	if ProxyRetHeadSize > len(pkg) {
		return false
	}

	reader := bytes.NewBuffer(pkg)
	err := binary.Read(reader, binary.BigEndian, header)
	if err != nil {
		//TODO record log
		return false
	}
	return true
}

func (bp *binaryProtocol) ParseSubMsg(pkg []byte, header *RpcSubHeader) bool {
	if header == nil {
		return false
	}

	//hsize := binary.Size(header)
	if SubHeaderSize > len(pkg) {
		return false
	}

	reader := bytes.NewBuffer(pkg)
	if err := binary.Read(reader, binary.BigEndian, header); err != nil {
		return false
	}
	return true
}

func (bp *binaryProtocol) ParsePubMsg(pkg []byte, header *RpcPubHeader) bool {
	if header == nil {
		return false
	}

	//hsize := binary.Size(header)
	if PubHeaderSize > len(pkg) {
		return false
	}

	reader := bytes.NewBuffer(pkg)
	if err := binary.Read(reader, binary.BigEndian, header); err != nil {
		return false
	}
	return true
}

func (bp *binaryProtocol) ParseCancelMsg(pkg []byte, header *RpcCancelSubHeader) bool {
	if header == nil {
		return false
	}

	//hsize := binary.Size(header)
	if CancelHeaderSize > len(pkg) {
		return false
	}

	reader := bytes.NewBuffer(pkg)
	if err := binary.Read(reader, binary.BigEndian, header); err != nil {
		return false
	}
	return true
}

// PackPlatoMsg 通用消息打包函数，使用反射实现
func (bp *binaryProtocol) PackPlatoMsg(header interface{}, body []byte, length int) ([]byte, int) {
	if header == nil {
		return nil, 0
	}

	pkg := make([]byte, 0, length)
	writer := bytes.NewBuffer(pkg)

	err := binary.Write(writer, binary.BigEndian, header)
	if err != nil {
		return nil, 0
	}

	pkg = writer.Bytes()
	if len(body) > 0 {
		pkg = append(pkg, body...)
	}

	return pkg, len(pkg)
}

func (bp *binaryProtocol) PackRespMsg(resp *ResponsePackage) ([]byte, int) {

	totallen := RespHeadSize + len(resp.Buffer)
	pkg := make([]byte, totallen)

	binary.BigEndian.PutUint32(pkg[0:], resp.Header.Length)
	binary.BigEndian.PutUint32(pkg[4:], resp.Header.Type)
	binary.BigEndian.PutUint32(pkg[8:], resp.Header.ServerID)
	//FIXME wkk 把这个改成64位了 记得配合修改
	binary.BigEndian.PutUint32(pkg[12:], resp.Header.CallID)
	binary.BigEndian.PutUint32(pkg[16:], resp.Header.ErrorCode)
	copy(pkg[RespHeadSize:], resp.Buffer)
	return pkg, totallen
}

func (bp *binaryProtocol) PackProxyRespMsg(resp *ProxyRespPackage) ([]byte, int) {
	if resp == nil {
		return nil, 0
	}

	totalLen := ProxyRetHeadSize + len(resp.Buffer)
	pkg := make([]byte, totalLen)
	binary.BigEndian.PutUint32(pkg[0:], resp.Header.Length)
	binary.BigEndian.PutUint32(pkg[4:], resp.Header.Type)
	binary.BigEndian.PutUint32(pkg[8:], resp.Header.ServerID)
	binary.BigEndian.PutUint32(pkg[12:], resp.Header.CallID)
	binary.BigEndian.PutUint32(pkg[16:], resp.Header.ErrorCode)
	binary.BigEndian.PutUint32(pkg[20:], uint32(resp.Header.GlobalIndex))
	copy(pkg[ProxyRetHeadSize:], resp.Buffer)

	return pkg, totalLen
}

func (bp *binaryProtocol) PackReqMsg(req *RequestPackage) ([]byte, int) {
	if req == nil {
		return nil, 0
	}

	totallen := CallHeadSize + len(req.Buffer)
	pkg := make([]byte, totallen)

	binary.BigEndian.PutUint32(pkg[0:], req.Header.Length)
	binary.BigEndian.PutUint32(pkg[4:], req.Header.Type)
	binary.BigEndian.PutUint64(pkg[8:], req.Header.ServiceUUID)
	binary.BigEndian.PutUint32(pkg[16:], req.Header.ServerID)
	//FIXME wkk 把这个改成64位了 记得配合修改
	binary.BigEndian.PutUint32(pkg[20:], req.Header.CallID)
	binary.BigEndian.PutUint32(pkg[24:], req.Header.MethodID)

	copy(pkg[CallHeadSize:], req.Buffer)
	return pkg, totallen
}

func (bp *binaryProtocol) PackProxyReqMsg(req *ProxyRequestPackage) ([]byte, int) {
	if req == nil {
		return nil, 0
	}

	totallen := ProxyCallHeadSize + len(req.Buffer)
	pkg := make([]byte, totallen)

	binary.BigEndian.PutUint32(pkg[0:], req.Header.Length)
	binary.BigEndian.PutUint32(pkg[4:], req.Header.Type)
	binary.BigEndian.PutUint64(pkg[8:], req.Header.ServiceUUID)
	binary.BigEndian.PutUint32(pkg[16:], req.Header.ServerID)
	binary.BigEndian.PutUint32(pkg[20:], req.Header.CallID)
	binary.BigEndian.PutUint32(pkg[24:], req.Header.MethodID)
	binary.BigEndian.PutUint32(pkg[28:], uint32(req.Header.GlobalIndex))
	binary.BigEndian.PutUint16(pkg[32:], req.Header.OneWay)

	copy(pkg[ProxyCallHeadSize:], req.Buffer)
	return pkg, totallen
}

func (bp *binaryProtocol) PackSubMsg(msg *RpcSubPackage) ([]byte, int) {
	if msg == nil {
		return nil, 0
	}

	totallen := SubHeaderSize + len(msg.Buffer)
	pkg := make([]byte, totallen)

	binary.BigEndian.PutUint32(pkg[0:], msg.Header.Length)
	binary.BigEndian.PutUint32(pkg[4:], msg.Header.Type)
	copy(msg.Header.SubId[:], pkg[8:])
	binary.BigEndian.PutUint32(pkg[24:], msg.Header.ProxyId)
	binary.BigEndian.PutUint64(pkg[28:], msg.Header.ServiceUUID)
	binary.BigEndian.PutUint32(pkg[36:], msg.Header.ServiceID)
	binary.BigEndian.PutUint32(pkg[40:], msg.Header.NameLen)
	binary.BigEndian.PutUint32(pkg[44:], msg.Header.DataLen)

	copy(pkg[SubHeaderSize:], msg.Buffer)
	return pkg, totallen
}

func (bp *binaryProtocol) PackPubMsg(msg *RpcPubPackage) ([]byte, int) {
	if msg == nil {
		return nil, 0
	}

	totallen := PubHeaderSize + len(msg.Buffer)
	pkg := make([]byte, totallen)

	binary.BigEndian.PutUint32(pkg[0:], msg.Header.Length)
	binary.BigEndian.PutUint32(pkg[4:], msg.Header.Type)
	copy(msg.Header.SubId[:], pkg[8:])
	binary.BigEndian.PutUint32(pkg[24:], msg.Header.ProxyId)
	binary.BigEndian.PutUint32(pkg[28:], msg.Header.ValueLen)

	copy(pkg[PubHeaderSize:], msg.Buffer)
	return pkg, totallen
}

func (bp *binaryProtocol) PackCancelMsg(msg *RpcCancelPackage) ([]byte, int) {
	if msg == nil {
		return nil, 0
	}

	totallen := CancelHeaderSize + len(msg.Buffer)
	pkg := make([]byte, totallen)

	binary.BigEndian.PutUint32(pkg[0:], msg.Header.Length)
	binary.BigEndian.PutUint32(pkg[4:], msg.Header.Type)
	copy(msg.Header.SubId[:], pkg[8:])

	copy(pkg[CancelHeaderSize:], msg.Buffer)
	return pkg, totallen
}
