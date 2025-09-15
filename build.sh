#!/bin/bash

# iperf-go 构建脚本

echo "================================"
echo "    iperf-go 构建脚本"
echo "================================"
echo ""

# 构建原始命令行工具
echo "构建原始 iperf-go..."
go build -o cmd/iperf-go cmd/main.go
if [ $? -eq 0 ]; then
    echo "✅ iperf-go 构建成功"
else
    echo "❌ iperf-go 构建失败"
    exit 1
fi

# 构建持续运行服务器
echo ""
echo "构建持续运行服务器..."
go build -o cmd/iperf-server-loop/iperf-server-loop cmd/iperf-server-loop/main.go
if [ $? -eq 0 ]; then
    echo "✅ iperf-server-loop 构建成功"
else
    echo "❌ iperf-server-loop 构建失败"
    exit 1
fi

echo ""
echo "================================"
echo "构建完成！"
echo ""
echo "可执行文件："
echo "  - cmd/iperf-go           （原始客户端/服务器）"
echo "  - cmd/iperf-server-loop  （持续运行服务器）"
echo ""
echo "使用方法："
echo "  客户端: ./cmd/iperf-go -c <server_ip>"
echo "  单次服务器: ./cmd/iperf-go -s"
echo "  持续服务器: ./cmd/iperf-server-loop/iperf-server-loop -p 5201"
echo "================================"
