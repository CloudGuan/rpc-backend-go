package main

import "fmt"

//所有的IDL的type的类型
const (
	IDL_TYPE_SEQ = iota
	IDL_TYPE_MAP
	IDL_TYPE_SET
	IDL_TYPE_I8
	IDL_TYPE_I16
	IDL_TYPE_I32
	IDL_TYPE_I64
	IDL_TYPE_UI8
	IDL_TYPE_UI16
	IDL_TYPE_UI32
	IDL_TYPE_UI64
	IDL_TYPE_STRING
	IDL_TYPE_BOOL
	IDL_TYPE_FLOAT
	IDL_TYPE_DOUBLE
	IDL_TYPE_VOID
)

var (
	idl2go map[string]Generator
)

//@title init default package
func init() {
	idl2go = make(map[string]Generator)
	//初始化各种检查工具
	idl2go["seq"] = &SeqType{}
	idl2go["map"] = &MapType{}
	idl2go["set"] = &SetType{}
	//初始化 默认类型
	dtype := &DefaultType{}
	idl2go["i8"] = dtype
	idl2go["i16"] = dtype
	idl2go["i32"] = dtype
	idl2go["i64"] = dtype
	idl2go["ui8"] = dtype
	idl2go["ui16"] = dtype
	idl2go["ui32"] = dtype
	idl2go["ui64"] = dtype
	idl2go["string"] = dtype
	idl2go["bool"] = dtype
	idl2go["float"] = dtype
	idl2go["double"] = dtype

	fmt.Println("idl2go init finish ")
}

//@title 基础检查工具用来实现各种类型的检查
type Generator interface {
	TypeValid(string, *ArgNode) bool
	GenVar(v *ArgNode) (string, error)
}
