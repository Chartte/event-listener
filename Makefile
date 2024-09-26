# 定义变量
VERSION = 4.2
APP_NAME = event-listener
IMAGE_NAME = registry.cn-beijing.aliyuncs.com/kunpengcloud/$(APP_NAME)

# 可配置的变量，允许通过命令行覆盖
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0

# 默认目标，执行 `make` 时默认运行这个目标
all: build docker-build docker-push

# 编译 Go 程序
build:
	@echo "开始编译 Go 程序..."
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(APP_NAME) .
	@echo "Go 编译完成"

# 构建 Docker 镜像
docker-build: build
	@echo "开始构建 Docker 镜像..."
	docker build -t $(IMAGE_NAME):$(VERSION) .
	@echo "Docker 镜像构建完成"

# 推送 Docker 镜像
docker-push: docker-build
	@echo "开始推送 Docker 镜像..."
	docker push $(IMAGE_NAME):$(VERSION)
	@echo "Docker 镜像推送完成"

# 清理本地生成的可执行文件
clean:
	@echo "清理生成的文件..."
	rm -f $(APP_NAME)
	@echo "清理完成"