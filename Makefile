.PHONY: build build-base build-reward build-pos clean

# 交叉编译目标：Linux amd64（部署到服务器）
GOOS   := linux
GOARCH := amd64
OUTDIR := ./bin

build: build-base build-reward build-pos

build-base:
	@echo "Building base..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(OUTDIR)/base ./app/base/api

build-reward:
	@echo "Building reward..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(OUTDIR)/reward ./app/reward/api

build-pos:
	@echo "Building pos..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(OUTDIR)/pos ./app/pos/api

clean:
	rm -rf $(OUTDIR)
