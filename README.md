# RPC go 框架

> 本工程是基于[RPC Frontend](https://gitee.com/dennis-kk/rpc-frontend) 产生的配置生成RPC框架，框架本身不包含socket通信相关代码，但可以快速与已有的网络通信框架集成。主要实现了[RPC Frontend](https://gitee.com/dennis-kk/rpc-frontend) 自定义的协议通信，可以与[c++](https://gitee.com/dennis-kk/rpc-backend-cpp) , [c#](https://gitee.com/dennis-kk/rpc-backend-csharp) 以及[Lua](https://gitee.com/dennis-kk/rpc-backend-lua) 框架进行通信。

## 依赖文件

- go 1.13+
- windows： visual studio 2019 
- linux: gcc9+
- protobuf 3.10.0+

## 目录结构说明

```shell
└─idlrpc
    ├─common 		#公共函数
    ├─idl2go		#idl协议 开发脚手架生成工具
    ├─net			#网络相关是实列代码，可以替换为你自己的网络库
    ├─protocol		
    ├─service
    ├─stub
    ├─stubcall
    └─transport		#传输工具，里面包含一个使用范例，需要你根据自身情况定制
```

## 使用流程

- 获取goimports, protoc-gen-go, go-rpc 工具包，并编译代码生成工具

  ```shell
  go get golang.org/x/tools/cmd/goimports
  go install golang.org/x/tools/cmd/goimports 
  go install google.golang.org/protobuf/cmd/protoc-gen-go
  #更新idlrpc 仓库
  go get -u gitee.com/dennis-kk/rpc-go-backend/idlrpc
  go install gitee.com/dennis-kk/rpc-go-backend/idlrpc/idl2go
  ```

- 编写 idl 文件，idl文件语法定义参考[RPC Frontend](https://gitee.com/dennis-kk/rpc-frontend) 

  ```txt
  struct Data {
      i32 field1
      string field2
      seq<string> field3
      set<string> field4
      dict<i64,string> field5
      bool field6
      float field7
      double field8    
  }
  
  service Service {
      void method1(Data,string)
      string method2(i8,set<string>,ui64)        
  }
  ```

- 使用rpc frontend 工具生成go语言接口描述文件，与protobuf的接口描述文件

  ```shell
  ./rpc-frontend -f example.idl -t go 
  ./rpc-frontend -f example.idl -t protobuf
  ```

  生成完毕后你会得到两份文件分别为: example.idl.go.json  与 example.idl.protobuf.json

- 调用golang_pb_layer.py 生成go语言适用的protobuf协议 

  ```shell
  ./golang_pb_layer.py example.idl.go.json example.idl.protbuf.json
  ```

- 调用编译好的goidltool工具生成开发脚手架以及代码

  ```shell
  ./idl2go -out your_path -I example.idl.go.json -idlpath outpath
  ```

## 生成结构说明

生成完成后，你可以在你指定的目录下看到如下的结构：

```shell 
F:.
├─idldata #idl 定义结构体文件，
│  │  example.struct.go	#接口参数，调用和被调用的自定义结构体均在里面
│  │  go.mod
│  │
│  └─pbdata
│          example.service.pb.go 	#pb协议，发送用不必关心
│
├─login	#以你自定义的服务名为单位，你有多少个服务就有多少个文件夹
│  │  go.mod
│  │
│  ├─impl
│  │      login_impl.go #接口定义文件
│  │
│  ├─proxy
│  │      login_proxy.go #客户端proxy文件，如果要作为调用者则使用这个文件的结构
│  │
│  ├─stub
│  │      login_stub.go
│  │
│  └─usr
│          go.mod
│          login_usr.go	#实现文件，开发者实际需要实现的文件
```

## 使用

新建一个空工程，使用go mod 初始话工程，并且添加依赖文件, 下文均使用example 文件为例，如果你需要导入多个包，按照例子添加即可。

```shell
module rpc-repo/gomock

go 1.15

require (
	gitee.com/dennis-kk/rpc-go-backend v0.0.0
	example/idldata v0.0.0
	example/servicedynamic v0.0.0
	example/servicedynamic_impl v0.0.0
)

replace (
	gitee.com/dennis-kk/rpc-go-backend => ../../src/go_service/rpc-go-backend
	example/idldata => ../../src/go_service/example/idldata
	example/servicedynamic => ../../src/go_service/example/servicedynamic
	example/servicedynamic_impl => ../../usr/go/example/servicedynamic_impl
)
```

**特别说明**：对于rpc-go-backend的引入，可以采用示例中的使用本地仓库的模式，也是讲最后的版本号修改为last或者指定版本来获取。

对于你自己服务，如果你已经发到了github并且打上了tag可以直接使用版本号下载依赖，如果只是本地测试，可以模仿本例使用相对路径。

### 代码骨架

#### 服务器代码骨架

- 创建实现实例

```go
dynamicptr = &servicedynamic_impl.ServiceDynamicClient{}
```

- 注册stub与实例到service

```go
err := idlrpc.RegisterService(serviceptr.GetUUID(), servicestub.NewServiceStub(), serviceptr)
	if err != nil {
		fmt.Printf("Load Service error: %v !!!\n", err)
		return
	}
```

- 启动rpc

```go
idlrpc.Start()
```

- 在主协程或者你自己的协程启动rpc主循环

```go
	for {
		//update rpc
		idlrpc.Tick()
	}
```

#### 客户端代码骨架

- 创建一个Transport实例，这里按照你的分配规则添加，可以随机可以指定连接

  ```go
  utiltool.GetRandomTrans()
  ```

- 创建客户端实列

  ```go
  Dynamicclient := dynamicproxy.NewServiceDynamicProxy(utiltool.GetRandomTrans())
  ```

- 启动rpc框架

  ```go
  idlrpc.Start()
  ```

- rpc主循环

  ```go
  idlrpc.Tick()
  ```

- 一次rpc调用代码

  ```go
  res,err := Dynamicclient.Method2(0, data.Field4, 0)
  ```