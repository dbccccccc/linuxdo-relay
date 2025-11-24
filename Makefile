.PHONY: help build test clean run docker-build docker-push install-tools lint fmt

# Variables
BINARY_NAME=linuxdo-relay
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

help: ## 显示帮助信息
	@echo "LinuxDo-Relay Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## 构建项目
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/server

build-all: ## 构建所有平台
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/server
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/server
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe ./cmd/server

test: ## 运行测试
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage: test ## 运行测试并生成覆盖率报告
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## 清理构建文件
	rm -f $(BINARY_NAME)*
	rm -f coverage.out coverage.html
	rm -rf dist/

run: ## 运行项目
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/server
	./$(BINARY_NAME)

install-tools: ## 安装开发工具
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint

lint: ## 运行代码检查
	golangci-lint run --timeout=5m

fmt: ## 格式化代码
	$(GOFMT) ./...

mod-tidy: ## 整理依赖
	$(GOMOD) tidy

mod-download: ## 下载依赖
	$(GOMOD) download

docker-build: ## 构建 Docker 镜像
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-build-latest: ## 构建 Docker 镜像（latest）
	docker build -t $(BINARY_NAME):latest .

docker-push: ## 推送 Docker 镜像
	docker push $(BINARY_NAME):$(VERSION)

docker-run: ## 运行 Docker 容器
	docker-compose up -d

docker-stop: ## 停止 Docker 容器
	docker-compose down

docker-logs: ## 查看 Docker 日志
	docker-compose logs -f

web-install: ## 安装前端依赖
	cd web && npm ci

web-dev: ## 运行前端开发服务器
	cd web && npm run dev

web-build: ## 构建前端
	cd web && npm run build

web-test: ## 运行前端测试
	cd web && npm run test

all: clean mod-download build test ## 完整构建流程

dev: ## 启动开发环境
	docker-compose up -d postgres redis
	@echo "Waiting for services to start..."
	@sleep 3
	@echo "Services ready. Run 'make run' to start the application."

.DEFAULT_GOAL := help
