//text 语法模板定义
package main

import (
	"strings"
	"text/template"
	"time"
)

var (
	idlpackagename string
	//工具函数
	funcsMap = template.FuncMap{"idltime": TimeFormat,
		"stfieldup":  StFieldFormat,
		"getidlname": GetIdlName,
		"tolower":    strings.ToLower,
		"isupper":    StartWithUppercase,
		"pbfiled":    DealPbStructField,
	}
)

const (
	//@title 类型定义模板
	typeletter = `{{- define "typelet"}}
{{- if eq .GoType "map"}}
{{- if .Value}} map[{{.Key.GoType}}]{{if .Value.IsStruct}}{{"*"}}{{getidlname}}{{end}}{{.Value.GoType}}
{{- else}} map[{{.Key.GoType}}]bool{{end}}
{{- else if eq .GoType "[]"}} []{{if .Key.IsStruct}}{{"*"}}{{getidlname}}{{end}}{{.Key.GoType}}
{{- else}} {{if .IsStruct}}*{{getidlname}}{{else if .IsEnum}}{{getidlname}}{{end}}{{.GoType}}{{end}}
{{- end}}`

	//@title 结构体生成
	stblock = `{{- define "stblock"}}type {{.Name}} struct{
{{- range $index,$elem := .Fields}}` + "\n\t{{stfieldup .Name}} {{block \"typelet\" $elem}}{{end}} `json:\"{{.Name}}\" bson:\"{{.Name}}\"`\n" +
		`{{- end}}
}
{{end}}
`
	//@title 结构体文件生成模板
	stletter = `
// Generated by the go idl tools. DO NOT EDIT {{idltime}}
// source: {{.IdlName}}

package idldata
import "{{.IdlName}}/idldata/pbdata"

//test hhh 
{{range .Enums}}
type {{.Name}} int32
{{$ename := .Name}}
const (
	{{- range .Fields}}
	{{$ename}}_{{.Name}} {{$ename}} = {{.Value}}
	{{- end}}
)
{{end}}

{{range .Structs}}
{{- block "stblock" .}}{{end}}

func(this *{{.Name}}) SerializeToPb() *pbdata.{{.Name}}{
	if this == nil {
		return nil
	}

	pbobj := &pbdata.{{.Name}}{
	}
	{{- range .Fields }}
	{{- $fn := stfieldup .Name}}
	{{- $pbn := pbfiled .Name}}
	{{- if eq .IdlType "i8" "i16"}}
	pbobj.{{$pbn}} = int32(this.{{$fn}})
	{{- else if eq .IdlType "ui8" "ui16"}}
	pbobj.{{$pbn}} = uint32(this.{{$fn}})
	{{- else if eq .IdlType "set" }}
	for k, _ := range this.{{$fn}} {
		pbobj.{{$pbn}} = append(pbobj.{{$pbn}}, k)
	}
	{{- else if eq .IdlType "seq"}}
	for _,v := range this.{{$fn}} {
		{{- if .Key.IsStruct}}
		pbobj.{{$pbn}} = append(pbobj.{{$pbn}}, v.SerializeToPb())
		{{- else}}
		pbobj.{{$pbn}} = append(pbobj.{{$pbn}}, v)
		{{- end}}
	}
	{{- else if eq .IdlType "dict"}}
	if pbobj.{{$pbn}} == nil{
		pbobj.{{$pbn}}=make(map[{{.Key.GoType}}]{{if .Value.IsStruct}}*pbdata.{{end}}{{.Value.GoType}})
	}
	for k,v := range this.{{$fn}} {
		{{- if .Value.IsStruct}}
		pbobj.{{$pbn}}[k] = v.SerializeToPb()
		{{- else}}
		pbobj.{{$pbn}}[k] = v
		{{- end}}
	}
	{{- else if .IsStruct}}
	pbobj.{{$pbn}} = this.{{$fn}}.SerializeToPb()
	{{- else if .IsEnum}}
	pbobj.{{$pbn}} = pbdata.{{.GoType}}(this.{{$fn}})
	{{- else}}
	pbobj.{{$pbn}} = this.{{$fn}}
	{{- end}}
	{{- end}}
	return pbobj
}

func (this *{{.Name}}) ParseFromPb(pbobj *pbdata.{{.Name}}) {
	if pbobj == nil {
		return
	}
	{{- range .Fields }}
	{{- $fn := stfieldup .Name}}
	{{- $pbn := pbfiled .Name}}
	{{- if eq .IdlType "i8" "i16" "ui8" "ui16"}}
	this.{{$fn}} = {{.GoType}}(pbobj.{{$pbn}})
	{{- else if eq .IdlType "set" }}
	if this.{{$fn}} == nil{
		this.{{$fn}}=make(map[{{.Key.GoType}}]bool)
	}
	for _,v := range pbobj.{{$pbn}} {
		this.{{$fn}}[v] = true
	}
	{{- else if eq .IdlType "seq"}}
	for _,v := range pbobj.{{$pbn}} {
		{{- if .Key.IsStruct}}
		obj := &{{.Key.GoType}}{}
		obj.ParseFromPb(v)
		this.{{$fn}} = append(this.{{$fn}}, obj)
		{{- else}}
		this.{{$fn}} = append(this.{{$fn}}, {{.Key.GoType}}(v))
		{{- end}}
	}
	{{- else if eq .IdlType "dict"}}
	if this.{{$fn}} == nil{
		this.{{$fn}} = make(map[{{.Key.GoType}}]{{if .Value.IsStruct}}*{{end}}{{.Value.GoType}})
	}
	for k,v := range pbobj.{{$pbn}} {
		{{- if .Value.IsStruct}}
		temp := &{{.Value.GoType}}{}
		temp.ParseFromPb(v)
		this.{{$fn}}[k] = temp
		{{- else}}
		this.{{$fn}}[k] = {{.Value.GoType}}(v)
		{{- end}}
	}
	{{- else if .IsStruct}}
	this.{{$fn}} = &{{.GoType}}{}
	this.{{$fn}}.ParseFromPb(pbobj.{{$pbn}})
	{{- else if .IsEnum}}
	this.{{$fn}} = {{.GoType}}(pbobj.{{$pbn}})
	{{- else}}
	this.{{$fn}} = pbobj.{{$pbn}}
	{{- end}}
	{{- end}}
}
{{end}}
`

	//@title 函数模板
	funcletter = `{{define "methods"}}{{range .}}
	{{stfieldup .Name}}(context.Context
{{- range $index,$elem := .Arguments}}
{{- print ","}} 
{{- if ne $elem.GoType "void" }}
{{- block "typelet" $elem}}{{end}}
{{- end}}
{{- end}})({{if .RetType}}{{if ne .RetType.GoType "void" }}{{block "typelet" .RetType}}{{end}},{{end}}error{{end}})
{{- end}}
{{- end}}
`
)

// TimeFormat 生成返回值类型
func TimeFormat() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// StFieldFormat go 语言用大小写来控制 是不是暴露给包外面使用 所以 需要把首字符转换为大写
func StFieldFormat(srcfield string) string {
	idx := strings.Index(srcfield, "_")
	if idx == -1 {
		return strings.ToUpper(srcfield[:1]) + srcfield[1:]
	}

	fls := strings.Split(srcfield, "_")
	res := ""
	for _, fl := range fls {
		if fl == "" {
			if res == "" {
				res = "X"
			} else {
				res = res + "_"
			}
		} else if (fl[0] >= 'a' && fl[0] <= 'z') || (fl[0] >= 'A' && fl[0] <= 'Z') {
			res = res + strings.ToUpper(fl[:1]) + fl[1:]
		} else {
			res = res + "_" + fl
		}
	}
	return res
}

func GetIdlName() string {
	if len(idlpackagename) != 0 {
		return idlpackagename + "."
	}
	return ""
}
