package main

import (
	"errors"
	"fmt"
)

// DefaultType @title 通用类型类型 实现各种类型的生成以及实现接口
type DefaultType struct {
	nformat  string //普通格式化
	stformat string //针对结构体 格式化的时候传递指针 减少拷贝
}

type SeqType struct {
	nformat  string //普通格式化
	stformat string //针对结构体 格式化的时候传递指针 减少拷贝
}

type MapType struct {
	nformat  string //普通格式化
	stformat string //针对结构体 格式化的时候传递指针 减少拷贝
}

// SetType 不支持结构体
type SetType struct {
	nformat string //普通格式化
}

type VoidType struct {
	nformat  string //普通格式化
	stformat string //针对结构体 格式化的时候传递指针 减少拷贝
}

func (d *DefaultType) TypeValid(method string, v *ArgNode) bool {
	return true
}

// GenVar format index gotype
func (d *DefaultType) GenVar(v *ArgNode) (string, error) {
	if v.IsStruct {
		return fmt.Sprintf(d.stformat, v.Index, v.GoType), nil
	} else {
		return fmt.Sprintf(d.nformat, v.Index, v.GoType), nil
	}
}

func (s *SeqType) TypeValid(own string, v *ArgNode) bool {
	if s == nil {
		fmt.Println("seq var node is nil ")
		return false
	}

	if v.Key == nil || v.Value != nil {
		fmt.Printf("owner Name  %s, %d Seq key or value is not nil ", own, v.Index)
		return false
	}

	return true
}

// GenVar format index []gotype 针对结构体类型 数组也存储
func (s *SeqType) GenVar(v *ArgNode) (string, error) {
	//key 代表了数组的实际类型
	if v.Key == nil {
		return "", errors.New("key type is nill")
	}
	if v.Key.IsStruct {
		// %d []*%s
		return fmt.Sprintf(s.stformat, v.Index, v.Key.GoType), nil
	} else {
		// %d []%s
		return fmt.Sprintf(s.nformat, v.Index, v.Key.GoType), nil
	}
}

func (m *MapType) TypeValid(own string, v *ArgNode) bool {
	if m == nil {
		fmt.Println("map type is nil ")
	}

	//map require key value type
	if v.Key == nil || v.Value == nil {
		fmt.Printf("owner Name  %s, %d Seq key or value is not nil ", own, v.Index)
		return false
	}

	if v.IsStruct != false {
		fmt.Printf("own Name %s, %d Map Struct Type Generate Error !!!", own, v.Index)
		return false
	}

	return true
}

// GenVar map 对应格式生成
func (m *MapType) GenVar(v *ArgNode) (string, error) {

	if m == nil {
		return "", nil
	}

	if v.IsStruct == false {
		return "nil", errors.New("Map type gen tool not init")
	}

	if v.Key == nil || v.Value == nil {
		return "nil", errors.New("map keys or value is nil")
	}

	//警告map套用map的复杂数据结构 禁止map套map 以及map套用别的结构
	if v.Value.IdlType == "set" || v.Value.IdlType == "map" || v.Value.IdlType == "sep" {
		fmt.Printf("map value is not sample struct %d %s ", v.Index, v.IdlType)
		return "nil", errors.New("Map too confused")
	}

	if v.Value.IsStruct {
		return fmt.Sprintf(m.stformat, v.Key.GoType, v.Value.GoType), nil
	} else {
		return fmt.Sprintf(m.nformat, v.Key.GoType, v.Value.GoType), nil
	}
}

func (s *SetType) TypeValid(own string, v *ArgNode) bool {

	if s == nil {
		fmt.Println("Set Type Struct Is nil ")
		return false
	}

	if v.Key == nil {
		fmt.Printf("own Name %s, %d Set Struct Type Key is nil ", own, v.Index)
		return false
	}

	if v.Value != nil {
		fmt.Printf("own Name %s, %d Set Struct Type Value is not nil ", own, v.Index)
		return false
	}

	return true
}

// GenVar go map 自定义key 需要实现hash 所以生成时候不支持自定义数据结构
func (s *SetType) GenVar(v *ArgNode) (string, error) {
	if s == nil {
		return "nil", errors.New("set type gen is not init")
	}

	if v.Key == nil {
		return "", errors.New("Set type's key is nil")
	}

	//同理也是检查负责结构
	if v.Key.IsStruct {
		return "", errors.New("set not support Stuct Key")
	}

	//同理也不支持复杂结构
	if v.Key.IdlType == "set" || v.Key.IdlType == "map" || v.Key.IdlType == "seq" {
		return "", errors.New("set key type is too complex")
	}

	return fmt.Sprintf(s.nformat, v.Key.GoType), nil
}

func (vt *VoidType) TypeValid(method string, v *ArgNode) bool {

	if vt == nil {
		fmt.Println("Void type struct is nil")
		return false
	}

	if v.IsStruct {
		fmt.Printf("method %s, %d Void Struct Is Struct Type ", method, v.Index)
		return false
	}

	if v.Key != nil || v.Value != nil {
		return false
	}

	return true
}

// GenVar go 只有返回值支持空类型 不用生成代码
func (vt *VoidType) GenVar(v *ArgNode) (string, error) {
	return "", nil
}
