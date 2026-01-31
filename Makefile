.PHONY: build run test clean

# 构建所有服务
build:
	@echo "Building all services..."
	cd app/account/api && go build -o ../../../bin/account-api ./account.go
	cd app/account/rpc && go build -o ../../../bin/account-rpc ./account.go
	cd app/order/api && go build -o ../../../bin/order-api ./order.go
	cd app/order/rpc && go build -o ../../../bin/order-rpc ./order.go

# 运行所有服务
run: build
	@echo "Starting all services..."
	./bin/account-api &
	./bin/account-rpc &
	./bin/order-api &
	./bin/order-rpc &
	@echo "All services started"

# 清理
clean:
	rm -rf bin/*
	find . -name "*.out" -delete

# 测试
test:
	go test ./...

# Docker 构建
docker-build:
	docker build -f deploy/dockerfiles/Dockerfile-accountapi -t account-api:latest .
	docker build -f deploy/dockerfiles/Dockerfile-accountrpc -t account-rpc:latest .
	docker build -f deploy/dockerfiles/Dockerfile-orderapi -t order-api:latest .
	docker build -f deploy/dockerfiles/Dockerfile-orderrpc -t order-rpc:latest .