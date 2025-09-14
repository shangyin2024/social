#!/bin/bash

# 启动脚本
echo "🚀 启动 Social Media Platform API..."

# 检查Redis是否运行
if ! redis-cli ping >/dev/null 2>&1; then
  echo "❌ Redis 未运行，请先启动 Redis"
  echo "可以使用以下命令启动 Redis:"
  echo "  docker run -d -p 6379:6379 redis:alpine"
  echo "  或者"
  echo "  redis-server"
  exit 1
fi

echo "✅ Redis 连接正常"

# 检查配置文件
if [ ! -f "config.yaml" ]; then
  echo "❌ 配置文件 config.yaml 不存在"
  echo "请复制 config.yaml.example 并填入正确的配置"
  exit 1
fi

echo "✅ 配置文件存在"

# 启动应用
echo "🎯 启动应用..."
go run main.go
