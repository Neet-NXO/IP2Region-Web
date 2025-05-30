# ip2region-web Makefile

# 应用名称
APP_NAME = ip2region-web
BUILD_DIR = build
FRONTEND_DIR = frontend
DIST_DIR = $(FRONTEND_DIR)/dist

# Go 构建配置
GO_BUILD_FLAGS = -ldflags="-s -w"
GO_VERSION = $(shell go version)

# 默认目标
all: build-all

# 检查环境
check-env:
	@echo "检查构建环境..."
	@echo "Go版本: $(GO_VERSION)"
	@which node > /dev/null || (echo "错误: 需要安装Node.js" && exit 1)
	@which npm > /dev/null || (echo "错误: 需要安装npm" && exit 1)
	@echo "环境检查通过"

# 下载Go依赖
deps:
	@echo "下载Go依赖..."
	go mod tidy
	go mod download

# 安装前端依赖
deps-frontend:
	@echo "安装前端依赖..."
	cd $(FRONTEND_DIR) && npm install

# 构建后端
build-backend:
	@echo "构建后端..."
	go build $(GO_BUILD_FLAGS) -o $(APP_NAME)

# 构建前端
build-frontend:
	@echo "构建前端..."
	cd $(FRONTEND_DIR) && npm run build

# 清理构建产物
clean:
	@echo "清理构建产物..."
	rm -f $(APP_NAME)
	rm -f $(APP_NAME).exe
	rm -rf $(DIST_DIR)
	rm -rf $(BUILD_DIR)

# 测试
test:
	@echo "运行测试..."
	go test -v ./...

# 代码格式化
fmt:
	@echo "格式化Go代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	@which golangci-lint > /dev/null || (echo "警告: 未安装golangci-lint，跳过代码检查" && exit 0)
	golangci-lint run

# 完整构建（后端+前端）
build-all: check-env deps deps-frontend build-frontend build-backend
	@echo "构建完成: $(APP_NAME)"
	@echo "前端文件: $(DIST_DIR)"
	@echo "使用方法: ./$(APP_NAME) -port=8080 -static=./$(DIST_DIR)"

# 开发模式
dev: deps
	@echo "启动开发模式..."
	go run main.go

# 开发模式（前端）
dev-frontend:
	@echo "启动前端开发服务器..."
	cd $(FRONTEND_DIR) && npm run dev

# 生产构建
prod: build-all
	@echo "生产构建完成"

# 帮助信息
help:
	@echo "可用的构建目标:"
	@echo "  all/build-all    - 完整构建（默认）"
	@echo "  build-backend    - 仅构建后端"
	@echo "  build-frontend   - 仅构建前端"
	@echo "  deps             - 下载Go依赖"
	@echo "  deps-frontend    - 安装前端依赖"
	@echo "  clean            - 清理构建产物"
	@echo "  test             - 运行测试"
	@echo "  fmt              - 格式化代码"
	@echo "  lint             - 代码检查"
	@echo "  dev              - 开发模式（后端）"
	@echo "  dev-frontend     - 开发模式（前端）"
	@echo "  prod             - 生产构建"
	@echo "  help             - 显示此帮助信息"

.PHONY: all build-all build-backend build-frontend deps deps-frontend clean test fmt lint dev dev-frontend prod check-env help
