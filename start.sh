#!/bin/bash

# 主机管理系统快速启动脚本

echo "🚀 主机管理系统启动脚本"
echo "========================"

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    echo "   macOS: brew install --cask docker"
    echo "   Ubuntu: curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh"
    exit 1
fi

# 检查Docker Compose是否可用
if ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose 不可用，请确保 Docker Desktop 正在运行"
    exit 1
fi

# 创建数据目录
echo "📁 创建数据目录..."
mkdir -p data

# 选择数据库模式
echo ""
echo "请选择数据库模式："
echo "1) SQLite (默认，适合开发和小规模部署)"
echo "2) MySQL (推荐生产环境)"
echo ""
read -p "请输入选择 (1 或 2，默认为 1): " choice

case $choice in
    2)
        echo "🐬 启动 MySQL 模式..."
        echo "🧹 正在清理旧容器和数据卷..."
        docker compose down -v
        docker rmi host-manager-frontend host-manager-backend
        docker compose build --no-cache
        docker compose --profile mysql up -d
        ;;
    *)
        echo "🗃️ 启动 SQLite 模式..."
        echo "🧹 正在清理旧容器和数据卷..."
        docker compose down -v
        docker rmi host-manager-frontend host-manager-backend
        docker compose build --no-cache
        docker compose up -d
        ;;
esac

echo ""
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo ""
echo "📊 服务状态："
docker compose ps

echo ""
echo "✅ 启动完成！"
echo ""
echo "🌐 访问地址："
echo "   前端: http://localhost:3000"
echo "   后端API: http://localhost:8080"
echo ""
echo "👤 默认登录账户："
echo "   用户名: admin"
echo "   密码: admin123"
echo ""
echo "📝 常用命令："
echo "   查看日志: docker compose logs -f"
echo "   停止服务: docker compose down"
echo "   重新构建: docker compose up -d --build" 