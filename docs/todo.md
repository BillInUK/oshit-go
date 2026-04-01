# 数据库

* 旧数据表到新数据库表迁移

* 不需要长期保留数据的表使用ttl

* 流水表采用分表
* * 流水表分表
* * 旧流水表里面错误的流水需要修正，比方说campaign那块的流水，之前记录有错误，看看是否可以修正

* 新版本的金额看看是否可以使用uint256插件来存放，或者大整型进行存放，避免精度丢失等问题

* 新版本使用postgres 18，然后要周期性的备份数据库到从aws的S3

# api迁移

* 原本的内网服务之间调用改

* 原本的内网服务之间调用改用dubbo调用

* 新版本的API需要单独搞个域名，这样方便前端测试和使用

* 需要用AI给前端改1个版本，让前端可以无缝的交接到新版本的API

* 新版本的API对于金额展示一律使用字面值进行展示，避免前端过度适配

*

# kafka的使用

* rocketmq替换成kafka
* 如果可以，可以规范kafka的签名

# solana rpc的使用


# 内网ALB的使用

* 现在内网用的还是NLB，我更希望能替换成ALB，这样业务更加方便


# 生成proto的步骤

```
  关于生成 proto 的步骤（供参考）：

  # 1. 修改 .proto 文件

  # 2. 在 proto 文件所在目录执行（paths=source_relative 让输出放在当前目录）
  cd app/pb/base
  protoc --go_out=. --go_opt=paths=source_relative base.proto

  # 注意：base.triple.go 是手写的，只定义服务方法（RPC），
  # 不涉及消息字段，所以消息字段变更时不需要动它。
  # 只有新增/删除 RPC 方法时才需要同步修改 base.triple.go。
```


