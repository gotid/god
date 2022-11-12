# mapreduce

## 为什么需要 MapReduce

在实际的业务场景中，我们常常需要从不同的 rpc 服务中获取响应属性来组装成复杂对象。

比如要查询商品详情：

1. 商品服务 - 查询商品属性
2. 库存服务 - 查询库存属性
3. 价格服务 - 查询价格属性
4. 营销服务 - 查询营销属性

如果是串行调用的话，响应时间会随着 rpc 服务调用次数呈线性增长，故我们要优化性能一般会将串行改并行。

简单的场景下使用 `WaitGroup` 也能满足需求，但如果我们需要对 rpc 调用返回的数据进行校验、转换、汇总呢？继续使用 `WaitGroup` 就有点不从心了，go 的官方库没有此类工具（java 中有 CompleteFuture），我们依据了 MapReduce 的架构思想实现了进程内的数据批处理 MapReduce 并发工具类。

## 设计思路

我们尝试把自己代入到作者的角色，梳理一下并发工具可能的业务场景：

1. 查询商品详情：支持并发调用多个服务来组合产品属性，支持调用错误可以立即结束。
2. 商品详情页自动推荐用户卡券：支持并发校验卡券，校验失败自动剔除，返回全部卡券。

以上实际上都是在清理输入数据，针对数据处理有个非常经典的异步模式：生产者消费者模式。于是，我们可以抽象一下数据批处理的生命周期，大致可分为三个阶段：

![img](https://raw.githubusercontent.com/zeromicro/zero-doc/main/doc/images/mapreduce-serial-cn.png)

1. 数据生产 generate
2. 数据加工 mapper
3. 数据聚合 reducer

其中，数据生产是不可或缺的阶段，数据加工、数据聚合是可选阶段。生产和加工支持并发调用，聚合基本属于纯内存操作，单协程即可。

然后，思考一下不同阶段之间的数据应该如何流转，既然数据处理由 goroutine 执行，那么自然可以考虑用 channel 来实现 goroutine 之间的通信啦

![img](https://raw.githubusercontent.com/zeromicro/zero-doc/main/doc/images/mapreduce-cn.png)

 最后，如何实现随时终止流程呢？

在`goroutine` 中监听一个全局的结束 `channel` 和调用方提供的上下文 `ctx` 就行~！

## 简单示例

并行求平方和

```go
package main

import (
  "fmt"
  "log"
  
  "github.com/gotid/god/lib/mr"
)

func main() {
  val, err := mr.MapReduce(func(source chan <- any) {
    // 数据生产
    for i := 0; i < 10; i++ {
      source <- i
    }
  }, func(item any, writer mr.Writer, cancel func(error)) {
    // 数据加工
    i := item.(int)
    writer.Writer(i * i)
  }, func(pipe <-chan any, writer mr.Writer, cancel func(error)) {
    // 数据聚合
    var sum int
    for i := range(pipe) {
      sum += i.(int)
    }
    writer.Write(sum)
  })
  if err != nil {
    log.Fatal(err)
  }
  
  fmt.Println("结果：", val)
}
```

更多示例：[https://github.com/gotid/god/tree/master/examples/http](https://github.com/gotid/god/tree/master/examples/http)
