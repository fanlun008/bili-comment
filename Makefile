# Makefile for bili-comment

.PHONY: build clean install test help

# 变量定义
BINARY_NAME=bili-comment
BUILD_DIR=./build
MAIN_FILE=./main.go

# 默认目标
all: build

# 构建二进制文件
build:
	@echo "构建 $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "构建完成: $(BINARY_NAME)"

# 构建到指定目录
build-to-dir:
	@echo "构建 $(BINARY_NAME) 到 $(BUILD_DIR)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 交叉编译
build-all: build-linux build-windows build-darwin

build-linux:
	@echo "构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)

build-windows:
	@echo "构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)

build-darwin:
	@echo "构建 macOS 版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@echo "清理完成"

# 安装到系统路径
install: build
	@echo "安装 $(BINARY_NAME) 到 /usr/local/bin..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "安装完成"

# 运行测试
test:
	@echo "运行测试..."
	@go test -v ./...

# 格式化代码
fmt:
	@echo "格式化代码..."
	@go fmt ./...

# 检查代码
vet:
	@echo "检查代码..."
	@go vet ./...

# 更新依赖
deps:
	@echo "更新依赖..."
	@go mod tidy
	@go mod download

# 运行示例
example: build
	@echo "运行示例..."
	@./$(BINARY_NAME) --help

# 显示帮助
help:
	@echo "可用的命令:"
	@echo "  build          构建二进制文件"
	@echo "  build-to-dir   构建到 build 目录"
	@echo "  build-all      交叉编译所有平台"
	@echo "  build-linux    构建 Linux 版本"
	@echo "  build-windows  构建 Windows 版本"
	@echo "  build-darwin   构建 macOS 版本"
	@echo "  clean          清理构建文件"
	@echo "  install        安装到系统路径"
	@echo "  test           运行测试"
	@echo "  fmt            格式化代码"
	@echo "  vet            检查代码"
	@echo "  deps           更新依赖"
	@echo "  example        运行示例"
	@echo "  help           显示帮助信息"
