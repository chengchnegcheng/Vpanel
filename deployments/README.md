# V Panel 部署说明

本目录包含 V Panel 的部署配置文件。

## 目录结构

```
deployments/
├── docker/                 # Docker 部署文件
│   ├── Dockerfile         # Docker 镜像构建文件
│   ├── docker-compose.yml # Docker Compose 配置
│   ├── .env.example       # 环境变量示例
│   └── .env               # 环境变量配置（需自行创建）
└── scripts/               # 部署脚本
    ├── start.sh           # 一键部署脚本
    ├── dev.sh             # 本地开发脚本
    └── entrypoint.sh      # Docker 容器入口脚本
```

## 🚀 一键部署（推荐）

最简单的部署方式，只需一条命令：

```bash
# 一键启动
./deployments/scripts/start.sh start

# 停止服务
./deployments/scripts/start.sh stop

# 重启服务
./deployments/scripts/start.sh restart

# 查看日志
./deployments/scripts/start.sh logs

# 查看状态
./deployments/scripts/start.sh status

# 清理所有数据（谨慎使用）
./deployments/scripts/start.sh clean
```

启动后访问 `http://localhost:8080`，默认账号 `admin`，密码查看 `.env` 文件。

## 🛠️ 本地开发

```bash
# 安装依赖
./deployments/scripts/dev.sh install

# 编译并启动
./deployments/scripts/dev.sh start

# 仅编译
./deployments/scripts/dev.sh build

# 直接运行（go run）
./deployments/scripts/dev.sh run

# 运行测试
./deployments/scripts/dev.sh test

# 启动前端开发服务器
./deployments/scripts/dev.sh frontend
```

## Docker 部署（手动方式）

### 快速开始

```bash
# 1. 进入 Docker 目录
cd deployments/docker

# 2. 创建环境变量配置
cp .env.example .env

# 3. 编辑配置（重要：修改密码和密钥）
vim .env

# 4. 启动服务
docker-compose up -d

# 5. 查看日志
docker-compose logs -f

# 6. 访问面板
# http://localhost:8080
```

### 环境变量配置

编辑 `.env` 文件：

```bash
# 应用版本
VERSION=latest

# 服务端口
V_SERVER_PORT=8080
V_SERVER_MODE=release

# 认证配置（重要：请修改）
V_JWT_SECRET=your-secure-jwt-secret-change-me
V_ADMIN_USER=admin
V_ADMIN_PASS=your-secure-admin-password

# 日志配置
V_LOG_LEVEL=info
V_LOG_FORMAT=json

# 时区
TZ=Asia/Shanghai
```

### 常用命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f

# 查看状态
docker-compose ps

# 进入容器
docker-compose exec v-panel sh

# 重新构建并启动
docker-compose up -d --build
```

### 数据持久化

Docker Compose 配置了以下数据卷：

| 卷名 | 容器路径 | 说明 |
|------|----------|------|
| `v-panel-data` | `/app/data` | 数据库和配置 |
| `v-panel-logs` | `/app/logs` | 日志文件 |
| `v-panel-xray` | `/app/xray` | Xray 相关文件 |

### 备份数据

```bash
# 备份数据库
docker cp v-panel:/app/data/v.db ./backup/v.db

# 或使用 docker-compose
docker-compose exec v-panel cat /app/data/v.db > ./backup/v.db
```

### 恢复数据

```bash
# 停止服务
docker-compose down

# 恢复数据库
docker cp ./backup/v.db v-panel:/app/data/v.db

# 启动服务
docker-compose up -d
```

## 手动 Docker 部署

如果不使用 docker-compose，可以手动运行：

```bash
# 构建镜像
docker build -t v-panel:latest -f deployments/docker/Dockerfile .

# 创建数据卷
docker volume create v-panel-data
docker volume create v-panel-logs

# 运行容器
docker run -d \
  --name v-panel \
  --restart unless-stopped \
  -p 8080:8080 \
  -v v-panel-data:/app/data \
  -v v-panel-logs:/app/logs \
  -e V_JWT_SECRET=your-secret \
  -e V_ADMIN_USER=admin \
  -e V_ADMIN_PASS=your-password \
  -e TZ=Asia/Shanghai \
  v-panel:latest

# 查看日志
docker logs -f v-panel
```

## 生产环境配置

### 使用 Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### 使用 Caddy 反向代理

```caddyfile
your-domain.com {
    reverse_proxy localhost:8080
}
```

### 安全建议

1. **修改默认密码** - 首次登录后立即修改管理员密码
2. **使用强 JWT 密钥** - 生成随机字符串：`openssl rand -hex 32`
3. **启用 HTTPS** - 使用反向代理配置 SSL 证书
4. **限制访问** - 配置防火墙只允许必要端口
5. **定期备份** - 设置定时任务备份数据库
6. **监控日志** - 定期检查日志发现异常

## 故障排除

### 容器无法启动

```bash
# 查看容器日志
docker-compose logs v-panel

# 检查容器状态
docker-compose ps

# 进入容器调试
docker-compose run --rm v-panel sh
```

### 端口冲突

```bash
# 检查端口占用
lsof -i :8080

# 修改端口（编辑 .env）
V_SERVER_PORT=8081
```

### 数据库问题

```bash
# 进入容器
docker-compose exec v-panel sh

# 检查数据库文件
ls -la /app/data/

# 检查数据库权限
sqlite3 /app/data/v.db ".tables"
```

### 健康检查失败

```bash
# 手动检查健康端点
curl http://localhost:8080/health

# 查看健康检查日志
docker inspect v-panel | grep -A 10 Health
```
