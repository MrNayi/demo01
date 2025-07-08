#!/bin/bash

# 部署脚本
set -e

echo "🚀 开始部署 demo01 项目..."

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 停止并删除现有容器
echo "🛑 停止现有容器..."
docker-compose down

# 清理旧镜像（可选）
if [ "$1" = "--clean" ]; then
    echo "🧹 清理旧镜像..."
    docker-compose down --rmi all --volumes --remove-orphans
fi

# 构建并启动服务
echo "🔨 构建并启动服务..."
docker-compose up --build -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 30

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose ps

# 检查健康状态
echo "🏥 检查健康状态..."
for i in {1..10}; do
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ 应用健康检查通过"
        break
    else
        echo "⏳ 等待应用启动... ($i/10)"
        sleep 10
    fi
done

# 显示服务信息
echo ""
echo "🎉 部署完成！"
echo "📊 服务信息："
echo "   - 应用地址: http://localhost:8080"
echo "   - 健康检查: http://localhost:8080/health"
echo "   - MySQL: localhost:3306"
echo "   - Redis: localhost:6379"
echo ""
echo "📝 常用命令："
echo "   - 查看日志: docker-compose logs -f app"
echo "   - 停止服务: docker-compose down"
echo "   - 重启服务: docker-compose restart"
echo "   - 清理数据: docker-compose down -v" 