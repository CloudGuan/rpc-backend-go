package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	toolVersion = "0.0.3"
)

type ImputFiles []string

var ProtocExec string    // protoc 可执行程序路径
var OutDir string        // 输出目录
var IdlDir string        //idl文件搜索根路径
var usrDir string        //用户路径
var updateService string //指定更新的服务名称
var gVersion string      //版本号
var gInputFile ImputFiles

//impl Value Interface

func (in *ImputFiles) String() string {
	return strings.Join(*in, ":")
}

func (in *ImputFiles) Set(str string) error {
	*in = append(*in, str)
	return nil
}

func printHelp() {
	bin := os.Args[0]
	if i := strings.LastIndex(bin, "/"); i != -1 {
		bin = bin[i+1:]
	}

	fmt.Printf("Usage %s -I example.idl -O impl/goservice", bin)
	flag.PrintDefaults()
}
func main() {

	flag.Usage = printHelp //解析出错时候 调用的默认方法
	flag.Var(&gInputFile, "I", "Input idle files, support multiple files !!!")
	flag.StringVar(&IdlDir, "idlpath", "", "default idl path")
	flag.StringVar(&OutDir, "out", "", "set out put dir，default is current！！！")
	flag.StringVar(&usrDir, "usr", "", "set usr impl dir, default is current path")
	flag.StringVar(&gImplPath, "impl", "impl", "Set your impl path")
	flag.StringVar(&updateService, "service", "", "Specify a service")
	flag.StringVar(&ProtocExec, "proto_dir", "protoc", "set protoc exec dir")
	flag.StringVar(&gVersion, "ver", "v0.3.3", "rpc-backend-go version")

	//flag.StringVar(&InputFile, "input", "", "set input file ")
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	// 转换为绝对路径
	OutDir, err = DealInputPath(OutDir)
	if err != nil {
		os.Exit(-1)
	}

	IdlDir, err = filepath.Abs(IdlDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	if len(updateService) != 0 {
		updateService = strings.ToLower(updateService)
	}

	// pre deal IdlDir
	IdlDir = strings.Replace(IdlDir, "\\", "/", -1)
	if IdlDir[len(IdlDir)-1] == '/' {
		IdlDir = IdlDir[:]
	}

	// pre deal usr Dir
	if usrDir == "" || len(usrDir) == 0 {
		usrDir = IdlDir
	} else {
		if usrDir, err = DealInputPath(usrDir); err != nil {
			os.Exit(-1)
		}
	}

	fmt.Println(dir)
	fmt.Printf("cur working: %s usr working: %s \n", IdlDir, usrDir)
	for _, v := range gInputFile {
		basename := filepath.Base(v)
		fmt.Printf("load idl json %s/%s \n", IdlDir, basename)
		if err = GenGoIdleService(IdlDir, basename); err != nil {
			fmt.Printf("gen idl %s file error  %s !!!\n ", v, err.Error())
			os.Exit(-1)
		}
	}
}
