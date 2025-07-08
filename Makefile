.PHONY: help build run test clean docker-build docker-run docker-stop deploy

# 默认目标
help:
	@echo "📋 可用命令："
	@echo "  build        - 编译项目"
	@echo "  run          - 本地运行项目"
	@echo "  test         - 运行测试"
	@echo "  clean        - 清理构建文件"
	@echo "  docker-build - 构建Docker镜像"
	@echo "  docker-run   - 运行Docker容器"
	@echo "  docker-stop  - 停止Docker容器"
	@echo "  deploy       - 一键部署"

# 编译项目
build:
	@echo "🔨 编译项目..."
	go build -o bin/demo01 ./cmd/main.go
	@echo "✅ 编译完成: bin/demo01"

# 本地运行
run:
	@echo "🚀 启动项目..."
	go run ./cmd/main.go

# 运行测试
test:
	@echo "🧪 运行测试..."
	go test -v ./...

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -rf bin/
	go clean

# 构建Docker镜像
docker-build:
	@echo "🐳 构建Docker镜像..."
	docker build -t demo01:latest .

# 运行Docker容器
docker-run:
	@echo "🐳 运行Docker容器..."
	docker-compose up -d

# 停止Docker容器
docker-stop:
	@echo "🛑 停止Docker容器..."
	docker-compose down

# 一键部署
deploy:
	@echo "🚀 一键部署..."
	./scripts/deploy.sh

# 查看日志
logs:
	@echo "📋 查看应用日志..."
	docker-compose logs -f app

# 健康检查
health:
	@echo "🏥 健康检查..."
	curl -f http://localhost:8080/health || echo "❌ 健康检查失败"

# 格式化代码
fmt:
	@echo "🎨 格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "🔍 代码检查..."
	golangci-lint run

# 依赖更新
deps:
	@echo "📦 更新依赖..."
	go mod tidy
	go mod download 