# DiTing-Go 测试文件夹结构

## 概述

本文件夹包含了DiTing-Go项目的所有测试文件，按照测试类型和功能模块进行了分类组织。

## 文件夹结构

```
tests/
├── unit/                    # 单元测试
├── integration/            # 集成测试
├── e2e/                   # 端到端测试
├── performance/           # 性能测试
├── scripts/               # 测试脚本
├── config/                # 测试配置文件
└── README.md              # 本文件
```

## 测试类型说明

### 1. 单元测试 (unit/)
- **目的**: 测试单个函数或方法的正确性
- **特点**: 快速、独立、可重复
- **覆盖范围**: 业务逻辑、工具函数、数据验证等

### 2. 集成测试 (integration/)
- **目的**: 测试多个组件之间的协作
- **特点**: 涉及数据库、缓存、外部服务
- **覆盖范围**: 服务层、数据访问层、缓存层

### 3. 端到端测试 (e2e/)
- **目的**: 测试完整的用户业务流程
- **特点**: 模拟真实用户操作
- **覆盖范围**: API接口、WebSocket、完整业务流程

### 4. 性能测试 (performance/)
- **目的**: 测试系统性能和稳定性
- **特点**: 高并发、大数据量、长时间运行
- **覆盖范围**: 负载测试、压力测试、基准测试

## 运行测试

### 运行所有测试
```bash
go test ./tests/...
```

### 运行特定类型测试
```bash
# 运行单元测试
go test ./tests/unit/...

# 运行集成测试
go test ./tests/integration/...

# 运行端到端测试
go test ./tests/e2e/...

# 运行性能测试
go test ./tests/performance/...
```

### 使用测试脚本
```bash
# Linux/Mac
./tests/scripts/run_tests.sh

# Windows
tests\scripts\run_tests.bat
```

## 测试配置

### 环境变量
```bash
# 测试环境
export GIN_MODE=test
export TEST_ENV=test

# 数据库配置
export TEST_DB_HOST=localhost
export TEST_DB_PORT=3306
export TEST_DB_NAME=diting_test

# Redis配置
export TEST_REDIS_HOST=localhost
export TEST_REDIS_PORT=6379
export TEST_REDIS_DB=1
```

## 最佳实践

### 1. 测试编写
- 每个测试函数只测试一个功能点
- 使用描述性的测试函数名
- 包含正向和异常测试用例

### 2. 测试数据
- 使用固定的测试数据确保可重复性
- 避免测试之间的数据依赖
- 及时清理测试数据

### 3. 测试维护
- 定期更新测试用例
- 保持测试代码的可读性
- 及时修复失败的测试 