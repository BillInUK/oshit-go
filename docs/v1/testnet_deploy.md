
# 测试环境

## 网络配置

| 节点 | ip | port | 版本 |
| --- | ---- | --- | --- |
| postgres | 172.31.3.251 | 5432 | 16.2 | 
| redis | 172.31.3.251 | 6379 | 7.0.15 |
| nacos | 172.31.3.251 | 8848 | 3.1.1 |
| kafka | 172.31.3.251 | 9902 | 4.2.0 |

## nginx配置

* 外网nginx配置

```
beta.testnet.oshit.io
```

* 反向代理配置

```
location /meme/base/api/v1 {
    proxy_pass http://172.31.48.185:80/base/api/v1;
}

location /meme/reward/api/v1 {
    proxy_pass http://172.31.48.251:80/reward/api/v1;
}

location /meme/snap/api/v1 {
    proxy_pass http://172.31.48.22:80/snap/api/v1;
}
```

* 反向代理内外网映射

| 服务 | 外网 | 内网 | 本地 |
| ---- | --- | ---- | ---- | 
| base |  https://beta.testnet.oshit.io/meme/base/api/v1 | http://172.31.48.185:80/base/api/v1 | http://127.0.0.1:1100/base/ |
| reward |  https://beta.testnet.oshit.io/meme/reward/api/v1 |  http://172.31.48.251:80/reward/api/v1 | http://127.0.0.1:1100/reward/ | 
| pos |  https://beta.testnet.oshit.io/meme/snap/api/v1 | http://172.31.48.22:80/snap/api/v1 | http://127.0.0.1:1100/snap/ |


* url路径

base
```
https://beta.testnet.oshit.io/meme/base/api/v1
```

reward
```
https://beta.testnet.oshit.io/meme/reward/api/v1
```

pos
```
https://beta.testnet.oshit.io/meme/snap/api/v1
```

## 部署路径

所有的服务都部署在对应服务器的 /data/dist/micro-server 目录下，目录下的文件目录结构为

```
├── bin
│   ├── micro-server            # 可执行文件
│   └── server.sh               # 启动脚本
├── etc                         # 配置文件路径
│   ├── application.yaml        # 应用程序配置文件
│   └── dubbo.yaml              # dubbo 配置文件
└── log                         # 日志目录
    └── server.pid
```

## 部署命令

* base

部署
```
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o micro-server base.go && \
scp -o StrictHostKeyChecking=no micro-server testnet-micro-common-1:/data/dist/micro-server/ & \
wait
```

* reward

部署
```
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o micro-server reward.go setup.go && \
scp -o StrictHostKeyChecking=no micro-server testnet-micro-reward-1:/data/dist/micro-server/ & \
wait
```

* pos

部署
```
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o micro-server pos.go setup.go && \
scp -o StrictHostKeyChecking=no micro-server testnet-micro-pos-1:/data/dist/micro-server/ & \
wait
```

## 启动

* 启动脚本

启动脚本路径位于 deploy/scripts/server.sh

* 启动命令

```shell
./bin/server.sh start
```

* 关闭
```shell
./bin/server.sh stop
```