package main

import (
	"errors"
	"fmt"
)

//@title set dict 这种复合结构的key value
type ComplexNode struct {
	IdlType  string `json:"IdlType"`
	IsStruct bool   `json:"isStruct"`
	GoType   string `json:"type"`
}

//idl 解析脚本 method 使用
type ArgNode struct {
	IdlType  string       `json:"IdlType"`
	Index    uint32       `json:"index"`
	IsStruct bool         `json:"isStruct"`
	IsEnum   bool         `json:"isEnum"`
	DeclName string       `json:"decl_name"`
	GoType   string       `json:"type"`
	Key      *ComplexNode `json:"key"`
	Value    *ComplexNode `json:"value"`
}

//idl struct 使用相关
type FieldNode struct {
	IdlType  string       `json:"IdlType"`
	Name     string       `json:"name"`
	Index    uint32       `json:"index"`
	IsStruct bool         `json:"isStruct"`
	IsEnum   bool         `json:"isEnum"`
	GoType   string       `json:"type"`
	Key      *ComplexNode `json:"key"`
	Value    *ComplexNode `json:"value"`
}

type MethodNode struct {
	Arguments []*ArgNode `json:"arguments"`
	Index     uint32     `json:"index"`
	Name      string     `json:"name"`
	Noexcept  bool       `json:"noexcept"`
	IsOneway  bool       `json:"oneway"`
	RetType   *ArgNode   `json:"retType"` //只有在oneway的時候有才没有返回值 不做类型检查
	Retry     uint32     `json:"retry"`
	TimeOut   uint32     `json:"timeout"`
}

type ServiceNode struct {
	LoadType string        `json:"loadType"`
	MaxInst  uint32        `json:"maxInst"`
	Methods  []*MethodNode `json:"methods"`
	Name     string        `json:"name"`
	SrvType  string        `json:"type"`
	Uuid     string        `json:"uuid"`
}

// idl 协议解析 enum 相关
type EnumFieldNode struct {
	Name  string `json:"name"`
	Value int32  `json:"value"`
}

// 枚举相关的json结构
type EnumNode struct {
	Name   string           `json:"name"`
	Fields []*EnumFieldNode `json:"fields"`
}

type StructNode struct {
	Name   string       `json:"name"`
	Fields []*FieldNode `json:"fields"`
}

type IdlJsonNode struct {
	ServiceNames []string       `json:serviceNames`
	Services     []*ServiceNode `json:"services"`
	StructNames  []string       `json:"structNames"`
	Structs      []*StructNode  `json:"structs"`
	EnumNames    []string       `json:"enumNames"`
	Enums        []*EnumNode    `json:"enums"`
	MaxInst      uint32         `json:"maxInst"`
	IdlName      string         `json:"idlname"`
}

func (ij *IdlJsonNode) IsValid() bool {
	if ij == nil {
		return false
	}

	if len(ij.ServiceNames) != len(ij.Services) {
		return false
	}
	//添加自己的检查代码
	return true
}

//@add for debug
func (s *ServiceNode) String() string {
	return fmt.Sprintf("Uuid %s type %s max %d Name %s srvType %s \n method %v \n", s.Uuid, s.LoadType, s.MaxInst, s.Name, s.SrvType, s.Methods)
}

func (m *MethodNode) String() string {
	return fmt.Sprintf("Index %d Name %s excpet %t oneway %t args %v", m.Index, m.Name, m.Noexcept, m.IsOneway, m.Arguments)
}

func (v *ArgNode) String() string {
	return fmt.Sprintf("index %d idltype %s gotype %s struct %t \n", v.Index, v.IdlType, v.GoType, v.IsStruct)
}

func checkVarVailid(own string, v *ArgNode) error {
	itype, ok := idl2go[v.IdlType]
	if !ok {
		return errors.New("Can't find type " + v.IdlType)
	}

	if itype.TypeValid(own, v) == false {
		errinfo := fmt.Sprintf("method %s param %d type %s check error !!!", own, v.Index, v.IdlType)
		return errors.New(errinfo)
	}

	return nil
}

func (m *MethodNode) IsValid() error {
	if m == nil {
		return errors.New("method node is nil !!!!")
	}

	// 检查方法名字
	if m.Name == "" {
		return errors.New("function Name is invalid ")
	}
	// 检查参数
	for _, param := range m.Arguments {
		err := checkVarVailid(m.Name, param)
		if err != nil {
			return err
		}
	}

	// 检查返回值
	err := checkVarVailid(m.Name, m.RetType)
	if err != nil {
		return err
	}

	return nil
}
