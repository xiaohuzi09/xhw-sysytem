#!/bin/bash
# 从终端运行应用以保留控制台输出
# 用法: ./run-console.sh

set -e

cd "$(dirname "$0")"

APP="bin/app-image-handle"

if [ ! -f "$APP" ]; then
    echo "错误: 找不到应用二进制文件"
    echo "请先执行: task build"
    exit 1
fi

echo "启动 app-image-handle (控制台模式)..."
echo "日志将输出到当前终端"
echo "按 Ctrl+C 停止"
echo ""

"$APP"
