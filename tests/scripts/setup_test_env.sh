#!/bin/bash

# DiTing-Go 测试环境设置脚本

echo "=== DiTing-Go 测试环境设置 ==="
echo ""

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go环境，请先安装Go"
    exit 1
fi

# 检查Docker环境（可选）
if command -v docker &> /dev/null; then
    echo "✅ 检测到Docker环境"
else
    echo "⚠️  未检测到Docker环境，将使用本地服务"
fi

# 设置测试环境变量
echo "设置测试环境变量..."
export GIN_MODE=test
export TEST_ENV=test

# 检查数据库连接
echo "检查数据库连接..."
# 这里可以添加数据库连接检查逻辑

# 检查Redis连接
echo "检查Redis连接..."
# 这里可以添加Redis连接检查逻辑

# 准备测试数据
echo "准备测试数据..."
# 这里可以添加测试数据准备逻辑

# 清理旧的测试数据
echo "清理旧的测试数据..."
# 这里可以添加清理逻辑

echo ""
echo "✅ 测试环境设置完成"
echo ""
echo "可以运行以下命令开始测试:"
echo "  ./tests/scripts/run_tests.sh"
echo "  go test -v ./tests/..."
echo "" 