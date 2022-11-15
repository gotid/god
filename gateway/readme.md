# Gateway 网关

## 用法

- main.go

```go
var configFile = flag.String("f", "config.yaml", "配置文件")

func main() {
  flag.Parse()
  
  var c gateway.Config
  conf.MustLoad(*configFile, &c)
  gw := gateway.MustNewServer(c)
  defer gw.Stop()
  gw.Start()
}
```




- config.yaml

```yaml
Name: demo-gateway
Host: localhost
Port: 8888
Upstreams:
  - Grpc:
      Etcd:
        Hosts:
          - localhost:2379
        Key: hello.rpc
    # protoset 模式
    ProtoSets:
      - hello.pb
    # Mappings 也可以在 proto 选项中进行覆盖
    Mappings:
      - Method: get
        Path: /pingHello/:ping
        RpcPath: hello.Hello/Ping
  - Grpc:
      Endpoints:
        - localhost:8081
    # 反射模式，无需 ProtoSet 设置
    Mappings:
      - Method: post
        Path: /pingWorld
        RpcPath: world.World/Ping
```

## 生成 ProtoSet 文件

- 没有外部导入的示例命令

```shell
protoc --descriptor_set_out=hello.pb hello.proto
```

- 有外部导入的示例命令

```shell
protoc --include_imports --proto_path=. --descriptor_set_out=hello.pb hello.proto
```

