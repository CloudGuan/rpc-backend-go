package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

var gImplPath string

// Gen 用来固定的部分逻辑代码
type Gen struct {
	Name      string       //包名 也是服务名字
	Service   *ServiceNode //解析出来的节点
	HasStruct bool         //是否有公共文件结构体
	Idlname   string       //idl 包名字用于管理公共的结构体
}

func (g *Gen) GenHead() string {
	return fmt.Sprintf("//Package  %s Generate By idl-go-tool \n //date: %v ", g.Name, time.Now().Format("2006-01-02 15:04:05"))
}

func GenGoIdleService(jsonpath, idlfile string) error {
	fullptah := fmt.Sprintf("%s/%s", jsonpath, idlfile)
	jsonfile, err := os.Open(fullptah)
	if err != nil {
		fmt.Printf("read file %s error %v \n", idlfile, err)
		return err
	}

	idldesc := IdlJsonNode{}

	jsonbytes, _ := ioutil.ReadAll(jsonfile)
	json.Unmarshal(jsonbytes, &idldesc)

	if PathExits(OutDir) == false {
		err := os.MkdirAll(OutDir, 0755)
		if err != nil {
			return err
		}
	}

	err = os.Chdir(OutDir)
	if err != nil {
		return err
	}

	//检查目录
	idlpath := strings.ToLower(idldesc.IdlName)
	if err := CheckIdlDir(idlpath); err != nil {
		fmt.Printf("checkout idl %s dir errror %v \n", idlpath, err)
		return err
	}
	// 切换目录
	if err := os.Chdir(idlpath); err != nil {
		return err
	}
	//生成idl公共包数据 如果有的话
	if len(idldesc.Structs) != 0 {
		if err := GenIdlData(idldesc.IdlName, idldesc.Structs, idldesc.Enums); err != nil {
			return err
		}
	} else {
		//没有struct 也要生成pb
		if err := GenPbLayer(idldesc.IdlName); err != nil {
			return err
		}
		os.Chdir("idldata")
		err := GenIdlGoMod(idldesc.IdlName)
		if err != nil {
			return err
		}
		os.Chdir("..")
	}
	//fmt.Printf("%v \n", idldesc)
	GenService(&idldesc)

	//切换回原来的目录
	os.Chdir("..")
	return nil
}

func CheckIdlDir(idlname string) error {
	var err error
	if PathExits(idlname) == false {
		err = os.Mkdir(idlname, 0755)
		if err != nil {
			fmt.Printf("create srevice: %s  dir error: %v !!!\n", idlname, err)
			return err
		}
	}
	return nil
}

func GenIdlData(idlname string, idlstructs []*StructNode, idlenums []*EnumNode) error {
	//生成pb文件
	if err := GenPbLayer(idlname); err != nil {
		return err
	}

	//切换到data 目录
	if PathExits("idldata") == false {
		err := os.Mkdir("idldata", 0755)
		if err != nil {
			return err
		}
	}

	os.Chdir("idldata")
	//不管怎么样都滚回之前的目录
	defer os.Chdir("../")

	//生成go的结构体
	err := GenGoData(idlname, idlstructs, idlenums)
	if err != nil {
		return err
	}

	return nil
}

// GenPbLayer proto 文件会被外置json 脚本生成好这里只需要调用就好
//生成好的这里需要拷贝到对应目录
func GenPbLayer(idlname string) error {
	//检查文件是否存在
	pbpath := fmt.Sprintf("%s/%s.service.proto", IdlDir, strings.ToLower(idlname))
	if FileExits(pbpath) == false {
		curdir, _ := os.Getwd()
		fmt.Printf("cur path: %s", curdir)
		return errors.New(pbpath + " file not ")
	}

	pbcmd := fmt.Sprintf("%s --go_out=./ %s", ProtocExec, pbpath)
	//if outMsg, err := exec.Command(ProtocExec, "-I="+IdlDir, "--go_out=./", pbpath).Output(); err != nil {
	//	return fmt.Errorf("[GenPbLayer] run protoc cmd %s error %q \n", pbcmd, string(outMsg))
	//}

	cmder := exec.Command(ProtocExec, "-I="+IdlDir, "--go_out=./", pbpath)
	errBuffer := &bytes.Buffer{}
	cmder.Stderr = errBuffer
	if err := cmder.Run(); err != nil {
		return fmt.Errorf("[GenPbLayer] run protoc cmd %s error:\n  %s \n", pbcmd, errBuffer.String())
	}

	return nil
}

func GenService(idljson *IdlJsonNode) {
	fmt.Printf(" services: %v \n", idljson.ServiceNames)
	if idljson.IsValid() == false {
		fmt.Printf(" service %v json is invalid!\n", idljson.ServiceNames)
		return
	}
	var err error
	for idx, v := range idljson.Services {
		//检查路径
		if idljson.ServiceNames[idx] == "" {
			continue
		}

		//如果制定了特定服务这里需要跳过其他服务 不更新
		if len(updateService) != 0 {
			if strings.ToLower(idljson.ServiceNames[idx]) != updateService {
				fmt.Printf("Skip idl: %s service: %s", idljson.IdlName, idljson.ServiceNames[idx])
				continue
			}
		}

		srvpath := strings.ToLower(idljson.ServiceNames[idx])
		if CheckSrvSubDir(srvpath) == false {
			fmt.Printf("generate Service %s working path error！！！\n", idljson.ServiceNames[idx])
			continue
		}
		//切换路径
		err = os.Chdir(srvpath)
		if err != nil {
			fmt.Printf("change to dir %s error !!! \n", idljson.ServiceNames[idx])
			continue
		}
		//生成data 引用文件
		gen := &Gen{
			Name:    idljson.ServiceNames[idx],
			Service: v,
			Idlname: idljson.IdlName,
		}

		if len(idljson.Structs) != 0 {
			gen.HasStruct = true
		}
		//生成mod文件
		err := GenImplMod(gen.Idlname, gen.Name)
		if err != nil {
			fmt.Printf("generate service %s:%s error %v !", gen.Idlname, gen.Name, err)
			return
		}
		//生成impl 文件
		GenServiceImpl(gen.Name, gen)
		//生成proxy文件
		GenProxyFile(gen.Idlname, gen)
		//生成stub文件
		StubGenFile(gen.Idlname, gen)
		//生成client文件 生相对路径绑在不接入包管理的情况下需要绑定
		GenClientImpl(gen.Idlname, gen)

		os.Chdir("../")

		//对所有的东西执行 go fmt 和 go import 去掉unused package
		if err = exec.Command("gofmt", "-l", "-w", "-s", "./"+srvpath).Run(); err != nil {
			curdir, _ := os.Getwd()
			fmt.Printf("go format service package : %s error %v cur dir：%s\n", srvpath, err, curdir)
		}
		if err = exec.Command("goimports", "-w", "./"+srvpath).Run(); err != nil {
			curdir, _ := os.Getwd()
			fmt.Printf("fix service package: %s error %v curdir %s \n", srvpath, err, curdir)
		}
	}
}

//这里默认应该在impl 目录 获取最后一层目录做测试
func CheckSrvSubDir(onesrv string) bool {
	//pwd, _ := os.Getwd()
	//paths := strings.Split(pwd, "\\")
	//if paths[len(paths)-1] != gImplPath {
	//	fmt.Printf("current working path is not implpath cur %s expect %s", pwd, gImplPath)
	//	return false
	//}
	var err error
	if PathExits(onesrv) == false {
		err = os.Mkdir(onesrv, 0755)
		if err != nil {
			fmt.Printf("make Service %s working path error %v !!!!", onesrv, err)
			return false
		}
	}
	return true
}
