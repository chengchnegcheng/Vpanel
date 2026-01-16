# V Panel 脚本说明

本目录包含 V Panel 的构建、启动和部署脚本。

## 脚本列表

| 脚本 | 说明 |
|------|------|
| `build.sh` | 构建脚本，编译前后端 |
| `start.sh` | V Panel 主程序启动脚本 |
| `start-agent.sh` | Node Agent 启动脚本 |
| `docker-build.sh` | Docker 镜像构建脚本 |

## 快速开始

### 1. 构建项目

```bash
# 完整构建（前端 + 后端）
./scripts/build.sh all

# 仅构建后端
./scripts/build.sh backend

# 仅构建前端
./scripts/build.sh frontend

# 多平台构建
./scripts/build.sh platforms

# 运行测试
./scripts/build.sh test

# 清理构建目录
./scripts/build.sh clean
```

### 2. 启动 V Panel

```bash
# 后台启动
./scripts/start.sh start

# 查看状态
./scripts/start.sh status

# 查看日志
./scripts/start.sh logs

# 停止服务
./scripts/start.sh stop

# 重启服务
./scripts/start.sh restart

# 前台运行（开发调试）
./scripts/start.sh run
```

### 3. 启动 Node Agent

```bash
# 构建 Agent
./scripts/start-agent.sh build

# 后台启动
./scripts/start-agent.sh start

# 查看状态
./scripts/start-agent.sh status

# 查看日志
./scripts/start-agent.sh logs

# 停止
./scripts/start-agent.sh stop

# 前台运行
./scripts/start-agent.sh run
```

## Docker 部署

### 使用 docker-build.sh

```bash
# 构建 Docker 镜像
./scripts/docker-build.sh build

# 构建多平台镜像（需要 buildx）
./scripts/docker-build.sh multiplatform

# 使用 docker-compose 启动
./scripts/docker-build.sh run

# 停止
./scripts/docker-build.sh stop

# 查看日志
./scripts/docker-build.sh logs

# 清理
./scripts/docker-build.sh clean
```

### 使用 docker-compose 直接部署

```bash
# 进入 Docker 部署目录
cd deployments/docker

# 复制环境变量配置
cp .env.example .env

# 编辑配置（修改密码等）
vim .env

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 使用 Docker 命令直接运行

```bash
# 构建镜像
docker build -t v-panel:latest -f deployments/docker/Dockerfile .

# 运行容器
docker run -d \
  --name v-panel \
  -p 8080:8080 \
  -v v-panel-data:/app/data \
  -e V_JWT_SECRET=your-secret \
  -e V_ADMIN_PASS=your-password \
  v-panel:latest
```

## 环境变量

### V Panel 主程序

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `CONFIG_FILE` | 配置文件路径 | `configs/config.yaml` |
| `V_SERVER_PORT` | 服务端口 | `8080` |
| `V_SERVER_MODE` | 运行模式 | `release` |
| `V_JWT_SECRET` | JWT 密钥 | - |
| `V_ADMIN_USER` | 管理员用户名 | `admin` |
| `V_ADMIN_PASS` | 管理员密码 | `admin123` |
| `V_LOG_LEVEL` | 日志级别 | `info` |
| `V_DB_PATH` | 数据库路径 | `data/v.db` |

### Node Agent

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `CONFIG_FILE` | 配置文件路径 | `configs/agent.yaml` |

## 文件说明

### 构建输出

- `build/` - 编译后的二进制文件
- `dist/` - 打包后的发布文件
- `web/dist/` - 前端构建输出

### 运行时文件

- `data/v-panel.pid` - V Panel 进程 ID 文件
- `data/v-panel.log` - V Panel 日志文件
- `data/v-agent.pid` - Agent 进程 ID 文件
- `data/v-agent.log` - Agent 日志文件
- `data/v.db` - SQLite 数据库文件

## 开发模式

开发时推荐使用前台运行模式，方便查看实时日志：

```bash
# 后端开发
./scripts/start.sh run

# 或直接使用 go run
go run ./cmd/v/main.go -config configs/config.yaml

# 前端开发（另开终端）
cd web
npm run dev
```

## 生产部署建议

1. **使用 Docker 部署** - 推荐使用 docker-compose 进行部署
2. **修改默认密码** - 首次部署后立即修改管理员密码
3. **配置 JWT 密钥** - 使用强随机字符串作为 JWT 密钥
4. **配置反向代理** - 使用 Nginx 或 Caddy 配置 HTTPS
5. **定期备份** - 定期备份 `data/v.db` 数据库文件

## 常见问题

### 端口被占用

```bash
# 查看端口占用
lsof -i :8080

# 修改端口
export V_SERVER_PORT=8081
./scripts/start.sh start
```

### 权限问题

```bash
# 添加执行权限
chmod +x scripts/*.sh
```

### 找不到配置文件

```bash
# 从示例创建配置
cp configs/config.yaml.example configs/config.yaml
cp configs/agent.yaml.example configs/agent.yaml
```
