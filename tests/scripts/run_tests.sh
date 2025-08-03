#!/bin/bash

# DiTing-Go 测试运行脚本

echo "=== DiTing-Go 测试运行脚本 ==="
echo ""

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go环境，请先安装Go"
    exit 1
fi

# 检查依赖
echo "检查依赖..."
go mod tidy

# 设置测试环境变量
export GIN_MODE=test

# 运行测试的函数
run_test() {
    local test_name=$1
    local test_pattern=$2
    local test_dir=$3
    
    echo ""
    echo "=== 运行测试: $test_name ==="
    echo "命令: go test -v -run $test_pattern $test_dir"
    echo ""
    
    go test -v -run "$test_pattern" "$test_dir"
    
    if [ $? -eq 0 ]; then
        echo "✅ $test_name 测试通过"
    else
        echo "❌ $test_name 测试失败"
        return 1
    fi
}

# 主菜单
show_menu() {
    echo ""
    echo "请选择要运行的测试:"
    echo "=== 单元测试 ==="
    echo "1. 用户服务单元测试"
    echo "2. 用户生命周期单元测试"
    echo "3. 注册验证单元测试"
    echo "4. 登录验证单元测试"
    echo ""
    echo "=== 集成测试 ==="
    echo "5. 用户注册集成测试"
    echo "6. 用户登录集成测试"
    echo "7. 用户注销集成测试"
    echo "8. 重复注册集成测试"
    echo "9. 错误凭据登录集成测试"
    echo "10. 错误验证码注销集成测试"
    echo "11. 有效数据注册集成测试"
    echo ""
    echo "=== 端到端测试 ==="
    echo "12. 用户完整工作流E2E测试"
    echo "13. 多用户并发E2E测试"
    echo "14. 用户错误场景E2E测试"
    echo ""
    echo "=== 性能测试 ==="
    echo "15. 用户注册性能测试"
    echo "16. 用户登录性能测试"
    echo "17. 并发用户注册测试"
    echo "18. 并发用户登录测试"
    echo "19. 负载测试"
    echo ""
    echo "=== 批量测试 ==="
    echo "20. 运行所有单元测试"
    echo "21. 运行所有集成测试"
    echo "22. 运行所有E2E测试"
    echo "23. 运行所有性能测试"
    echo "24. 运行所有测试"
    echo "0. 退出"
    echo ""
    read -p "请输入选项 (0-24): " choice
}

# 运行所有测试
run_all_tests() {
    echo ""
    echo "=== 运行所有测试 ==="
    
    test_dirs=(
        "./tests/unit/"
        "./tests/integration/"
        "./tests/e2e/"
        "./tests/performance/"
    )
    
    failed_tests=()
    
    for dir in "${test_dirs[@]}"; do
        echo ""
        echo "运行测试目录: $dir"
        go test -v "$dir"
        
        if [ $? -ne 0 ]; then
            failed_tests+=("$dir")
        fi
    done
    
    echo ""
    echo "=== 测试结果汇总 ==="
    if [ ${#failed_tests[@]} -eq 0 ]; then
        echo "✅ 所有测试通过"
    else
        echo "❌ 以下测试失败:"
        for test in "${failed_tests[@]}"; do
            echo "  - $test"
        done
    fi
}

# 主循环
while true; do
    show_menu
    
    case $choice in
        1)
            run_test "用户服务单元测试" "TestUserLifecycleFlow" "./tests/unit/"
            ;;
        2)
            run_test "用户生命周期单元测试" "TestUserLifecycleFlow" "./tests/unit/"
            ;;
        3)
            run_test "注册验证单元测试" "TestRegisterValidation" "./tests/unit/"
            ;;
        4)
            run_test "登录验证单元测试" "TestLoginValidation" "./tests/unit/"
            ;;
        5)
            run_test "用户注册集成测试" "TestUserRegistrationFlow" "./tests/integration/"
            ;;
        6)
            run_test "用户登录集成测试" "TestUserLoginFlow" "./tests/integration/"
            ;;
        7)
            run_test "用户注销集成测试" "TestUserCancelFlow" "./tests/integration/"
            ;;
        8)
            run_test "重复注册集成测试" "TestDuplicateRegistration" "./tests/integration/"
            ;;
        9)
            run_test "错误凭据登录集成测试" "TestLoginWithWrongCredentials" "./tests/integration/"
            ;;
        10)
            run_test "错误验证码注销集成测试" "TestCancelWithWrongCaptcha" "./tests/integration/"
            ;;
        11)
            run_test "有效数据注册集成测试" "TestUserRegistrationWithValidData" "./tests/integration/"
            ;;
        12)
            run_test "用户完整工作流E2E测试" "TestUserCompleteWorkflow" "./tests/e2e/"
            ;;
        13)
            run_test "多用户并发E2E测试" "TestMultipleUserWorkflow" "./tests/e2e/"
            ;;
        14)
            run_test "用户错误场景E2E测试" "TestUserErrorScenarios" "./tests/e2e/"
            ;;
        15)
            run_test "用户注册性能测试" "BenchmarkUserRegistration" "./tests/performance/"
            ;;
        16)
            run_test "用户登录性能测试" "BenchmarkUserLogin" "./tests/performance/"
            ;;
        17)
            run_test "并发用户注册测试" "TestConcurrentUserRegistration" "./tests/performance/"
            ;;
        18)
            run_test "并发用户登录测试" "TestConcurrentUserLogin" "./tests/performance/"
            ;;
        19)
            run_test "负载测试" "TestLoadTest" "./tests/performance/"
            ;;
        20)
            echo ""
            echo "=== 运行所有单元测试 ==="
            go test -v ./tests/unit/...
            ;;
        21)
            echo ""
            echo "=== 运行所有集成测试 ==="
            go test -v ./tests/integration/...
            ;;
        22)
            echo ""
            echo "=== 运行所有E2E测试 ==="
            go test -v ./tests/e2e/...
            ;;
        23)
            echo ""
            echo "=== 运行所有性能测试 ==="
            go test -v ./tests/performance/...
            ;;
        24)
            run_all_tests
            ;;
        0)
            echo "退出测试"
            exit 0
            ;;
        *)
            echo "无效选项，请重新选择"
            ;;
    esac
    
    echo ""
    read -p "按回车键继续..."
done 