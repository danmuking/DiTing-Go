@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo === DiTing-Go 用户服务测试 ===
echo.

REM 检查Go环境
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未找到Go环境，请先安装Go
    pause
    exit /b 1
)

REM 检查依赖
echo 检查依赖...
go mod tidy

REM 设置测试环境变量
set GIN_MODE=test

:menu
echo.
echo 请选择要运行的测试:
echo 1. 用户完整生命周期测试
echo 2. 用户注册流程测试
echo 3. 用户登录流程测试
echo 4. 用户注销流程测试
echo 5. 注册参数验证测试
echo 6. 登录参数验证测试
echo 7. 重复注册测试
echo 8. 错误凭据登录测试
echo 9. 错误验证码注销测试
echo 10. 有效数据注册测试
echo 11. 运行所有测试
echo 0. 退出
echo.
set /p choice=请输入选项 (0-11): 

if "%choice%"=="1" goto test_lifecycle
if "%choice%"=="2" goto test_register
if "%choice%"=="3" goto test_login
if "%choice%"=="4" goto test_cancel
if "%choice%"=="5" goto test_register_validation
if "%choice%"=="6" goto test_login_validation
if "%choice%"=="7" goto test_duplicate_register
if "%choice%"=="8" goto test_wrong_credentials
if "%choice%"=="9" goto test_wrong_captcha
if "%choice%"=="10" goto test_valid_data
if "%choice%"=="11" goto test_all
if "%choice%"=="0" goto exit
echo 无效选项，请重新选择
goto menu

:test_lifecycle
echo.
echo === 运行测试: 用户完整生命周期测试 ===
echo 命令: go test -v -run TestUserLifecycleFlow
echo.
go test -v -run TestUserLifecycleFlow ./service/
if %errorlevel% equ 0 (
    echo ✅ 用户完整生命周期测试通过
) else (
    echo ❌ 用户完整生命周期测试失败
)
goto continue

:test_register
echo.
echo === 运行测试: 用户注册流程测试 ===
echo 命令: go test -v -run TestUserRegistrationFlow
echo.
go test -v -run TestUserRegistrationFlow ./service/
if %errorlevel% equ 0 (
    echo ✅ 用户注册流程测试通过
) else (
    echo ❌ 用户注册流程测试失败
)
goto continue

:test_login
echo.
echo === 运行测试: 用户登录流程测试 ===
echo 命令: go test -v -run TestUserLoginFlow
echo.
go test -v -run TestUserLoginFlow ./service/
if %errorlevel% equ 0 (
    echo ✅ 用户登录流程测试通过
) else (
    echo ❌ 用户登录流程测试失败
)
goto continue

:test_cancel
echo.
echo === 运行测试: 用户注销流程测试 ===
echo 命令: go test -v -run TestUserCancelFlow
echo.
go test -v -run TestUserCancelFlow ./service/
if %errorlevel% equ 0 (
    echo ✅ 用户注销流程测试通过
) else (
    echo ❌ 用户注销流程测试失败
)
goto continue

:test_register_validation
echo.
echo === 运行测试: 注册参数验证测试 ===
echo 命令: go test -v -run TestRegisterValidation
echo.
go test -v -run TestRegisterValidation ./service/
if %errorlevel% equ 0 (
    echo ✅ 注册参数验证测试通过
) else (
    echo ❌ 注册参数验证测试失败
)
goto continue

:test_login_validation
echo.
echo === 运行测试: 登录参数验证测试 ===
echo 命令: go test -v -run TestLoginValidation
echo.
go test -v -run TestLoginValidation ./service/
if %errorlevel% equ 0 (
    echo ✅ 登录参数验证测试通过
) else (
    echo ❌ 登录参数验证测试失败
)
goto continue

:test_duplicate_register
echo.
echo === 运行测试: 重复注册测试 ===
echo 命令: go test -v -run TestDuplicateRegistration
echo.
go test -v -run TestDuplicateRegistration ./service/
if %errorlevel% equ 0 (
    echo ✅ 重复注册测试通过
) else (
    echo ❌ 重复注册测试失败
)
goto continue

:test_wrong_credentials
echo.
echo === 运行测试: 错误凭据登录测试 ===
echo 命令: go test -v -run TestLoginWithWrongCredentials
echo.
go test -v -run TestLoginWithWrongCredentials ./service/
if %errorlevel% equ 0 (
    echo ✅ 错误凭据登录测试通过
) else (
    echo ❌ 错误凭据登录测试失败
)
goto continue

:test_wrong_captcha
echo.
echo === 运行测试: 错误验证码注销测试 ===
echo 命令: go test -v -run TestCancelWithWrongCaptcha
echo.
go test -v -run TestCancelWithWrongCaptcha ./service/
if %errorlevel% equ 0 (
    echo ✅ 错误验证码注销测试通过
) else (
    echo ❌ 错误验证码注销测试失败
)
goto continue

:test_valid_data
echo.
echo === 运行测试: 有效数据注册测试 ===
echo 命令: go test -v -run TestUserRegistrationWithValidData
echo.
go test -v -run TestUserRegistrationWithValidData ./service/
if %errorlevel% equ 0 (
    echo ✅ 有效数据注册测试通过
) else (
    echo ❌ 有效数据注册测试失败
)
goto continue

:test_all
echo.
echo === 运行所有测试 ===
set failed_tests=

for %%t in (TestUserLifecycleFlow TestUserRegistrationFlow TestUserLoginFlow TestUserCancelFlow TestRegisterValidation TestLoginValidation TestDuplicateRegistration TestLoginWithWrongCredentials TestCancelWithWrongCaptcha TestUserRegistrationWithValidData) do (
    echo.
    echo 运行测试: %%t
    go test -v -run %%t ./service/
    if !errorlevel! neq 0 (
        set failed_tests=!failed_tests! %%t
    )
)

echo.
echo === 测试结果汇总 ===
if "%failed_tests%"=="" (
    echo ✅ 所有测试通过
) else (
    echo ❌ 以下测试失败:
    for %%t in (%failed_tests%) do (
        echo   - %%t
    )
)
goto continue

:continue
echo.
pause
goto menu

:exit
echo 退出测试
pause
exit /b 0 