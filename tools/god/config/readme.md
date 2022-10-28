# 配置项管理

| 名称           | 是否可选 | 说明     |
|--------------|------|--------|
| namingFormat | YES  | 文件名格式符 |

# naming-format

`namingFormat` 和日期格式符（yyyy-MM-dd）类似，可在对代码文件名进行格式化。

## 格式符（godesigner）

格式符由 `go` 和 `designer` 组成，常见的三种命名格式如下：

* 小写：`godesigner`
* 驼峰：`goDesigner`
* 蛇式：`go_designer`

效果示例：
源字符串：user_center

| 格式符            | 格式结果           | 说明            |
|----------------|----------------|---------------|
| godesigner     | usercenter     | 小写            |
| goDesigner     | userCenter     | 驼峰            |
| go_designer    | user_center    | 蛇式            |
| Go#designer    | User#center    | 井字符分割，大写开头    |
| GODESIGNER     | USERCENTER     | 大写            |
| \_go#designer_ | \_user#center_ | 井字符分割，前后缀以下划线 |


错误示例：
* go
* gODesigner
* designer
* goDEsigner
* goDESigner
* goeSigner
* haha

# 使用方法
目前可在生成 rpc、model、api时，通过 `-style` 指定格式，如：
```shell
god rpc proto -src test.proto -dir . -style go_designer
```

```shell
god model mysql datasource -url="" -table="*" -dir . -style GoDesigner
```

```shell
god api go test.api -dir . -style godesigner
```

# 默认值 `godesigner`