#!/bin/bash

# DiTing-Go E2E测试运行脚本

echo "=== DiTing-Go E2E测试运行脚本 ==="
echo ""

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go环境，请先安装Go"
    exit 1
fi

# 获取项目根目录
PROJECT_ROOT=$(pwd)
echo "项目根目录: $PROJECT_ROOT"

# 设置环境变量
export PROJECT_ROOT="$PROJECT_ROOT"
export GIN_MODE=test
export TEST_ENV=test

echo "环境变量设置:"
echo "  PROJECT_ROOT=$PROJECT_ROOT"
echo "  GIN_MODE=$GIN_MODE"
echo "  TEST_ENV=$TEST_ENV"

# 检查配置文件
if [ -f "$PROJECT_ROOT/conf/config.yml" ]; then
    echo "✅ 配置文件存在: $PROJECT_ROOT/conf/config.yml"
else
    echo "❌ 配置文件不存在: $PROJECT_ROOT/conf/config.yml"
    exit 1
fi

# 检查依赖
echo "检查依赖..."
go mod tidy

echo ""
echo "=== 运行E2E测试 ==="

# 确保在项目根目录运行测试
cd "$PROJECT_ROOT"

# 运行特定的E2E测试
echo "运行用户完整工作流E2E测试..."
go test -v -run TestUserCompleteWorkflow ./tests/e2e/

echo ""
echo "运行多用户并发E2E测试..."
go test -v -run TestMultipleUserWorkflow ./tests/e2e/

echo ""
echo "运行用户错误场景E2E测试..."
go test -v -run TestUserErrorScenarios ./tests/e2e/

echo ""
echo "=== E2E测试完成 ===" 