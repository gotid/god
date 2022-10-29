## 如何使用

1. 定义一个配置结构，如下：

```go
package demo

type DemoConfig struct {
Host string `json:",default=0.0.0.0"`
Port int
LogMode string `json:",options=[file,console]"`
}
```

2. 编写 yaml 或 json 配置文件：

- yaml 示例

```yaml
# 大部分字段是可选的，或拥有默认值
Port: 8080
# 可以使用环境变量
LogMode: ${LOG_MODE}
```

3. 加载配置文件至结构体

```go
package demo

// 遇错退出
var c config.Config
conf.MustLoad(configFile, &c)

// 或自行处理加载错误
var c config.Config
if err := conf.Load(configFile, &c); err != nil {
log.Fatal(err)
}

// 启用环境变量加载
var c config.Config
conf.MustLoad(configFile, &c, conf.UseEnv())
```