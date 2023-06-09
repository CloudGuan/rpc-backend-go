package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
)

var (
	idlsttp *template.Template
)

func init() {
	temp, err := template.New("typeletter").Funcs(funcsMap).Parse(typeletter)
	if err != nil {
		fmt.Printf("init struct letter error %v !!!!\n", err)
		return
	}

	temp, err = temp.Parse(stblock)
	if err != nil {
		fmt.Printf("parse struct bolck letter error %v !!!!\n", err)
		return
	}

	idlsttp, err = temp.Parse(stletter)
	if err != nil {
		fmt.Printf("parse struct letter error %v !!!!\n", err)
		return
	}
}

func GenGoData(idlname string, Structs []*StructNode, Enums []*EnumNode) error {
	if idlsttp == nil {
		return errors.New("idl struct template is nil !!!!")
	}

	type StGen struct {
		IdlName string
		Structs []*StructNode
		Enums   []*EnumNode
	}

	stgen := &StGen{
		idlname,
		Structs,
		Enums,
	}

	filename := strings.ToLower(idlname) + ".struct.go"

	if FileExits(filename) {
		//存在的话删了重新生成
		err := os.Remove(filename)
		if err != nil {
			fmt.Printf("delete %s idl common file  error %v !!! \n", filename, err)
			return err
		}
	}

	//生成go.mod
	if err := GenIdlGoMod(idlname); err != nil {
		return err
	}

	//打开文件

	filehanld, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0765)
	if err != nil {
		fmt.Printf("open %s idl struct file error %v !!!!", filename, err)
		return err
	}
	//关闭文件
	defer filehanld.Close()

	err = idlsttp.Execute(filehanld, stgen)
	if err != nil {
		return err
	}
	return nil
}

//生成 go mod 文件 每个对应的file都是需要
func GenIdlGoMod(idlname string) error {
	filename := "go.mod"
	if FileExits(filename) {
		//存在的话删了重新生成
		err := os.Remove(filename)
		if err != nil {
			fmt.Printf("delete %s idl common file  error %v !!! \n", filename, err)
			return err
		}
	}

	hfile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0765)
	if err != nil {
		fmt.Printf("Open Idl %s Common Struct Error  %v!!!!", filename, err)
		return err
	}

	defer hfile.Close()

	//没啥东西直接写就好了
	var codec bytes.Buffer

	codec.WriteString("// Machine generated code\n\n")
	codec.WriteString("// Code generated by go-idl-tool. DO NOT EDIT.\n")
	codec.WriteString(fmt.Sprintf("//date: %s \n//idltool version: %s \n//source: %s \n \n", TimeFormat(), toolVersion, idlname))

	codec.WriteString(`module ` + idlname + "/idldata")
	codec.WriteString("\n")
	codec.WriteString(`go 1.16`)
	codec.WriteString("\n\n")
	codec.WriteString(`require (
	github.com/golang/protobuf v1.4.3 // indirect
	google.golang.org/protobuf v1.23.0
)`)
	hfile.WriteString(codec.String())

	return nil
}
