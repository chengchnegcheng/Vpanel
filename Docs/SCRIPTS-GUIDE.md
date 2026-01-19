# 脚本使用指南

## 概述

V Panel 提供了一系列脚本来简化开发、部署和维护工作。

## 脚本列表

### 开发相关

#### 1. Makefile - 开发命令集合

**用途**: 提供常用开发命令的快捷方式

**使用**:
```bash
# 查看所有可用命令
make help

# 编译 Panel
make build

# 编译 Agent
make agent

# 运行 Panel
make run

# 开发模式（热重载）
make dev

# 运行测试
make test

# 代码检查
make lint

# 清理构建文件
make clean

# 完整构建
make all
```

**常用命令**:
- `make build` - 编译 Panel
- `make agent` - 编译 Agent
- `make dev` - 开发模式（需要 air）
- `make test` - 运行测试
- `make deploy-panel` - 部署 Panel
- `make api-test ADMIN_TOKEN=xxx` - 测试 API

---

#### 2. dev-setup.sh - 开发环境设置

**用途**: 一键设置开发环境

**使用**:
```bash
./scripts/dev-setup.sh
```

**功能**:
- ✅ 检查 Go、Node.js、PostgreSQL
- ✅ 安装 Go 依赖
- ✅ 安装开发工具（air, golangci-lint）
- ✅ 创建目录结构
- ✅ 创建配置文件
- ✅ 创建 Makefile

**输出**:
- `.air.toml` - air 热重载配置
- `Makefile` - 开发命令
- `configs/config.yaml` - 配置文件

---

#### 3. build-agent.sh - 编译 Agent

**用途**: 编译多平台 Agent 二进制

**使用**:
```bash
# 编译所有平台
./scripts/build-agent.sh

# 指定版本
VERSION=1.0.1 ./scripts/build-agent.sh
```

**支持平台**:
- Linux: amd64, arm64, arm
- macOS: amd64, arm64
- Windows: amd64

**输出**:
```
build/
├── vpanel-agent-linux-amd64
├── vpanel-agent-linux-amd64.tar.gz
├── vpanel-agent-linux-arm64
├── vpanel-agent-linux-arm64.tar.gz
├── vpanel-agent-darwin-amd64
├── vpanel-agent-darwin-amd64.tar.gz
└── ...
```

---

### 部署相关

#### 4. quick-deploy.sh - 快速部署

**用途**: 快速部署 Panel 或 Agent

**使用**:

**部署 Panel**:
```bash
./scripts/quick-deploy.sh panel
```

**部署 Agent**:
```bash
./scripts/quick-deploy.sh agent https://panel.example.com node-token-here
```

**功能**:
- Panel 部署:
  - 编译 Panel
  - 创建目录
  - 运行迁移
  - 启动服务

- Agent 部署:
  - 编译 Agent
  - 安装 Xray
  - 创建配置
  - 创建 systemd 服务
  - 启动服务

---

#### 5. install-xray.sh - 安装 Xray

**用途**: 自动安装和配置 Xray

**使用**:
```bash
sudo ./scripts/install-xray.sh
```

**功能**:
- ✅ 检测操作系统
- ✅ 安装依赖
- ✅ 下载安装 Xray
- ✅ 创建配置文件
- ✅ 配置 systemd 服务
- ✅ 验证安装

**输出**:
- `/usr/local/bin/xray` - Xray 二进制
- `/etc/xray/config.json` - 配置文件
- `/var/log/xray/` - 日志目录

---

### 测试相关

#### 6. test-api.sh - API 测试

**用途**: 测试 Panel API 端点

**使用**:
```bash
# 设置 token
export ADMIN_TOKEN=your-admin-token

# 运行测试
./scripts/test-api.sh

# 自定义 API URL
API_URL=https://panel.example.com ./scripts/test-api.sh
```

**测试内容**:
- ✅ 健康检查
- ✅ 节点管理
- ✅ 代理管理
- ✅ 配置生成
- ✅ SSH 连接测试

**示例输出**:
```
测试 V Panel API
API URL: http://localhost:8080

测试: 健康检查
✓ 成功 (HTTP 200)
{
  "status": "ok"
}

测试: 获取节点列表
✓ 成功 (HTTP 200)
{
  "data": [...]
}
```

---

#### 7. verify-migration.sh - 验证数据库迁移

**用途**: 验证数据库迁移是否正确执行

**使用**:
```bash
# 使用默认配置
./scripts/verify-migration.sh

# 自定义数据库配置
DB_HOST=localhost \
DB_PORT=5432 \
DB_NAME=vpanel \
DB_USER=vpanel \
DB_PASSWORD=password \
./scripts/verify-migration.sh
```

**检查项**:
- ✅ 数据库连接
- ✅ proxies 表存在
- ✅ node_id 字段存在
- ✅ 索引存在
- ✅ 外键约束存在
- ✅ nodes 表存在

**示例输出**:
```
验证数据库迁移
数据库: localhost:5432/vpanel

测试数据库连接...
✓ 数据库连接成功

检查 proxies 表...
✓ proxies 表存在

检查 node_id 字段...
✓ node_id 字段存在

检查索引...
✓ idx_proxies_node_id 索引存在

检查外键约束...
✓ fk_proxies_node 外键存在

验证完成！
```

---

### 运维相关

#### 8. health-check.sh - 健康检查

**用途**: 检查 Panel 和 Agent 的运行状态

**使用**:
```bash
# 检查 Panel 状态
./scripts/health-check.sh panel

# 检查 Agent 状态
./scripts/health-check.sh agent

# 检查所有组件
./scripts/health-check.sh all

# 检查系统资源
./scripts/health-check.sh system
```

**检查项**:
- ✅ 进程状态
- ✅ HTTP 端点响应
- ✅ 服务健康检查
- ✅ Xray 安装和运行
- ✅ 配置文件存在
- ✅ 日志错误统计
- ✅ 系统资源使用

**示例输出**:
```
检查 Panel 状态...

✓ Panel 进程运行中
✓ Panel HTTP 响应正常
✓ Panel 健康检查通过
✓ 日志状态正常

Panel 状态: 正常
```

**环境变量**:
- `PANEL_URL`: Panel 地址 (默认: http://localhost:8080)
- `AGENT_PORT`: Agent 端口 (默认: 8081)

---

#### 9. backup.sh - 备份和恢复

**用途**: 备份数据库和配置文件

**使用**:
```bash
# 备份数据库
./scripts/backup.sh database

# 备份配置文件
./scripts/backup.sh config

# 备份 Agent 配置
./scripts/backup.sh agent

# 备份所有内容
./scripts/backup.sh all

# 列出备份
./scripts/backup.sh list

# 恢复数据库
./scripts/backup.sh restore backups/database/vpanel_db_20260119_120000.sql.gz

# 恢复配置
./scripts/backup.sh restore backups/config/vpanel_config_20260119_120000.tar.gz

# 清理旧备份
./scripts/backup.sh clean
```

**功能**:
- ✅ 数据库备份（PostgreSQL）
- ✅ 配置文件备份
- ✅ Agent 配置备份
- ✅ 自动压缩
- ✅ 备份恢复
- ✅ 自动清理旧备份

**环境变量**:
- `BACKUP_DIR`: 备份目录 (默认: backups)
- `DB_HOST`: 数据库主机 (默认: localhost)
- `DB_PORT`: 数据库端口 (默认: 5432)
- `DB_NAME`: 数据库名称 (默认: vpanel)
- `DB_USER`: 数据库用户 (默认: vpanel)
- `DB_PASSWORD`: 数据库密码
- `KEEP_DAYS`: 保留天数 (默认: 7)

**示例**:
```bash
# 设置数据库密码并备份
DB_PASSWORD=mypassword ./scripts/backup.sh database

# 保留 30 天的备份
KEEP_DAYS=30 ./scripts/backup.sh clean
```

---

#### 10. log-rotate.sh - 日志轮转

**用途**: 管理和轮转日志文件

**使用**:
```bash
# 轮转日志
./scripts/log-rotate.sh rotate

# 清理旧日志
./scripts/log-rotate.sh clean

# 分析日志
./scripts/log-rotate.sh analyze

# 设置自动轮转
./scripts/log-rotate.sh setup

# 查看状态
./scripts/log-rotate.sh status
```

**功能**:
- ✅ 自动轮转大文件
- ✅ 压缩旧日志
- ✅ 清理过期日志
- ✅ 日志分析
- ✅ 错误统计
- ✅ 集成 logrotate
- ✅ 定时任务设置

**环境变量**:
- `LOG_DIR`: 日志目录 (默认: logs)
- `MAX_SIZE`: 最大文件大小 (默认: 100M)
- `KEEP_DAYS`: 保留天数 (默认: 30)
- `COMPRESS`: 是否压缩 (默认: true)

**示例输出**:
```
分析日志...

日志文件统计:
  文件数: 5
  总大小: 250MB

错误统计 (最近 1000 行):
  vpanel.log:
    ERROR: 3
    WARN: 12

最近的错误 (最多 5 条):
  2026-01-19 12:00:00 ERROR Failed to connect to database
```

---

## 使用场景

### 场景 1: 新开发者入门

```bash
# 1. 克隆仓库
git clone <repo-url>
cd V

# 2. 设置开发环境
./scripts/dev-setup.sh

# 3. 配置数据库
vim configs/config.yaml

# 4. 启动开发
make dev
```

### 场景 2: 生产部署

```bash
# 在 Panel 服务器
./scripts/quick-deploy.sh panel

# 在节点服务器
./scripts/quick-deploy.sh agent https://panel.example.com <token>
```

### 场景 3: 编译发布版本

```bash
# 编译 Panel
go build -o vpanel ./cmd/v/main.go

# 编译所有平台的 Agent
./scripts/build-agent.sh

# 打包
tar -czf vpanel-release.tar.gz vpanel build/
```

### 场景 4: 测试验证

```bash
# 验证数据库
./scripts/verify-migration.sh

# 测试 API
export ADMIN_TOKEN=your-token
./scripts/test-api.sh

# 运行单元测试
make test
```

### 场景 5: 运维监控

```bash
# 健康检查
./scripts/health-check.sh all

# 备份数据
./scripts/backup.sh all

# 分析日志
./scripts/log-rotate.sh analyze

# 设置自动化
./scripts/log-rotate.sh setup
./scripts/backup.sh clean
```

### 场景 6: 故障排查

```bash
# 检查服务状态
./scripts/health-check.sh all

# 查看最近错误
./scripts/log-rotate.sh analyze

# 恢复备份
./scripts/backup.sh list
./scripts/backup.sh restore backups/database/xxx.sql.gz

# 重启服务
systemctl restart vpanel-agent
```

---

## Makefile 命令

开发环境设置后，可以使用 Makefile 命令：

```bash
# 查看帮助
make help

# 编译 Panel
make build

# 编译 Agent
make agent

# 运行 Panel
make run

# 开发模式（热重载）
make dev

# 运行测试
make test

# 代码检查
make lint

# 清理构建文件
make clean
```

---

## 环境变量

### dev-setup.sh

无需环境变量

### build-agent.sh

- `VERSION`: 版本号（默认: 1.0.0）

### quick-deploy.sh

无需环境变量（通过参数传递）

### test-api.sh

- `API_URL`: API 地址（默认: http://localhost:8080）
- `ADMIN_TOKEN`: 管理员 Token（必需）

### verify-migration.sh

- `DB_HOST`: 数据库主机（默认: localhost）
- `DB_PORT`: 数据库端口（默认: 5432）
- `DB_NAME`: 数据库名称（默认: vpanel）
- `DB_USER`: 数据库用户（默认: vpanel）
- `DB_PASSWORD`: 数据库密码（必需）

### health-check.sh

- `PANEL_URL`: Panel 地址（默认: http://localhost:8080）
- `AGENT_PORT`: Agent 端口（默认: 8081）

### backup.sh

- `BACKUP_DIR`: 备份目录（默认: backups）
- `DB_HOST`: 数据库主机（默认: localhost）
- `DB_PORT`: 数据库端口（默认: 5432）
- `DB_NAME`: 数据库名称（默认: vpanel）
- `DB_USER`: 数据库用户（默认: vpanel）
- `DB_PASSWORD`: 数据库密码（必需）
- `KEEP_DAYS`: 保留天数（默认: 7）

### log-rotate.sh

- `LOG_DIR`: 日志目录（默认: logs）
- `MAX_SIZE`: 最大文件大小（默认: 100M）
- `KEEP_DAYS`: 保留天数（默认: 30）
- `COMPRESS`: 是否压缩（默认: true）

---

## 故障排查

### 脚本权限问题

```bash
# 添加执行权限
chmod +x scripts/*.sh
```

### 数据库连接失败

```bash
# 检查 PostgreSQL 是否运行
systemctl status postgresql

# 测试连接
psql -h localhost -U vpanel -d vpanel
```

### Go 依赖问题

```bash
# 清理并重新下载
go clean -modcache
go mod download
```

### 编译失败

```bash
# 检查 Go 版本
go version

# 更新依赖
go mod tidy
```

---

## 最佳实践

1. **开发环境**
   - 使用 `make dev` 进行开发
   - 定期运行 `make lint` 检查代码
   - 提交前运行 `make test`

2. **部署**
   - 使用脚本自动化部署
   - 部署前验证数据库迁移
   - 部署后运行 API 测试

3. **维护**
   - 定期更新依赖
   - 定期备份数据库
   - 监控日志文件

---

## 相关文档

- [快速开始](./quick-start-xray.md)
- [远程部署指南](./remote-deploy-guide.md)
- [已知问题](./KNOWN-ISSUES.md)
- [功能完成清单](./FEATURES-COMPLETED.md)
