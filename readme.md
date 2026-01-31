# 生成必要的auth token和key

```
1. 生成一个48字节的强随机字符串（长度远大于32字符）用来做 NACOS_AUTH_TOKEN
openssl rand -base64 36

2. 生成身份标识键 (KEY)，使用一个随机的UUID
openssl rand -hex 8

# 3. 生成身份标识值 (VALUE)，使用更强的随机字符串（可以更长一些）
openssl rand -base64 24

```

# 启动nacos

```
docker run --name nacos-standalone-derby \
    -e MODE=standalone \
    -e NACOS_AUTH_TOKEN=pHVWBPFlwQ9EYMntyyzIQPweJOdU7BerkjGFsV6aohFBGjxW \
    -e NACOS_AUTH_IDENTITY_KEY=4bb5bd15aa29f5e8 \
    -e NACOS_AUTH_IDENTITY_VALUE=+2a+Nwzd038BV5pEQDANara7fTwVbMXt \
    -p 8080:8080 \
    -p 8848:8848 \
    -p 9848:9848 \
    -d nacos/nacos-server:latest
```

# 进入控制台 
http://127.0.0.1:8080/index.html#/register
```
lJPQwjjO9k
```


# 使用protocol buf

```
# 1. 安装 protobuf 编译器 (protoc)
brew install protobuf
# 验证安装
protoc --version

# 2. 安装标准的 Go 语言插件 protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

```

## 安装关键的 Dubbo-go 插件

```
# 安装 Dubbo-go 的 Triple 协议代码生成器
# 请确保你的 go.mod 已经引入了 dubbo-go，它会提供这个插件
go install github.com/dubbogo/protoc-gen-go-triple/v3@v3.0.2

# 安装后，确保 `$GOPATH/bin`（通常是 `~/go/bin`）在你的系统 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc # 持久化到 zsh
```

## 编写你的 .proto 文件
```
// account.proto
syntax = "proto3";

package account; // 包名，在Go中会作为生成代码的包名

option go_package = "oshit-go/app/account/rpc/pb;account"; // 非常重要！指定生成Go代码的导入路径和包别名

// 定义请求和响应消息
message GetUserRequest {
  int64 user_id = 1;
}

message User {
  int64 id = 1;
  string username = 2;
  string email = 3;
}

message CreateUserRequest {
  string username = 1;
  string email = 2;
  string password = 3;
}

// 定义服务接口
service AccountService {
  rpc GetUser (GetUserRequest) returns (User) {}
  rpc CreateUser (CreateUserRequest) returns (User) {}
  rpc ValidateToken (google.protobuf.StringValue) returns (User) {} // 基本类型建议使用包装类型
}
```

## 第 4 步：执行代码生成命令

```
# 切换到你的 .proto 文件所在目录
cd oshit-go/app/account/rpc

# 执行 protoc 命令，同时生成 .pb.go 和 .triple.go 文件
protoc --go_out=. --go-triple_out=. ./proto/account.proto

# 如果你的 .proto 文件在其它目录，例如 ./proto/account.proto，命令如下：
# protoc --go_out=. --go-triple_out=. ./proto/account.proto
```