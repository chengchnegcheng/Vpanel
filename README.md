# V Panel - 高性能代理服务器

<div align="center">
  <p>
    <a href="#功能特点">功能特点</a> •
    <a href="#快速开始">快速开始</a> •
    <a href="#docker-部署">Docker 部署</a> •
    <a href="#开发指南">开发指南</a> •
    <a href="#项目结构">项目结构</a>
  </p>
</div>

V Panel 是一个用 Go 语言编写的高性能代理服务器管理面板，基于 Xray-core，支持多种代理协议。提供完整的用户管理、流量统计、证书管理等功能，以及直观的 Web 管理界面。

## 功能特点

### 核心功能
- **多协议支持**: VMess, VLESS, Trojan, Shadowsocks
- **用户管理**: 认证授权、流量限制、状态监控、多级权限
- **流量管理**: 实时统计、每日统计、流量限制和警告
- **证书管理**: 自动 SSL 证书申请和更新、多域名支持
- **Xray 管理**: 版本切换、远程更新、运行状态监控
- **系统管理**: 完整日志系统、系统状态监控、配置管理

## 快速开始

### 系统要求
- Go 1.23+
- Node.js 20+ (前端开发)
- Docker 20+ (容器部署)

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/chengchnegcheng/V.git
cd V

# 构建
./scripts/build.sh all

# 运行
./build/v-panel -config configs/config.yaml
```

### 首次运行
1. 访问 `http://localhost:8080`
2. 默认账号: `admin` / `admin123`
3. 首次登录后请立即修改密码

## Docker 部署

### 使用 Docker Compose (推荐)

```bash
# 进入部署目录
cd deployments/docker

# 复制环境变量配置
cp .env.example .env

# 编辑配置 (修改密码等)
vim .env

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 使用 Docker 命令

```bash
# 构建镜像
./scripts/docker-build.sh build

# 运行容器
docker run -d \
  --name v-panel \
  -p 8080:8080 \
  -v v-panel-data:/app/data \
  -e V_JWT_SECRET=your-secret \
  -e V_ADMIN_PASS=your-password \
  v-panel:latest
```

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `V_SERVER_PORT` | 服务端口 | 8080 |
| `V_SERVER_MODE` | 运行模式 (debug/release) | release |
| `V_JWT_SECRET` | JWT 密钥 | - |
| `V_ADMIN_USER` | 管理员用户名 | admin |
| `V_ADMIN_PASS` | 管理员密码 | admin123 |
| `V_LOG_LEVEL` | 日志级别 | info |
| `V_DB_PATH` | 数据库路径 | /app/data/v.db |

## 开发指南

### 开发环境

```bash
# 后端开发
go run ./cmd/v/main.go -config configs/config.yaml

# 前端开发
cd web
npm install
npm run dev
```

### 构建命令

```bash
# 完整构建
./scripts/build.sh all

# 仅构建后端
./scripts/build.sh backend

# 仅构建前端
./scripts/build.sh frontend

# 多平台构建
./scripts/build.sh platforms

# 运行测试
./scripts/build.sh test
```

## 项目结构

```
v/
├── cmd/v/                  # 应用入口
│   └── main.go
├── internal/               # 私有包
│   ├── api/               # API 层
│   │   ├── handlers/      # HTTP 处理器
│   │   ├── middleware/    # 中间件
│   │   └── routes.go      # 路由定义
│   ├── auth/              # 认证模块
│   ├── config/            # 配置管理
│   ├── database/          # 数据库层
│   │   ├── migrations/    # 数据库迁移
│   │   └── repository/    # 数据访问
│   ├── logger/            # 日志模块
│   ├── proxy/             # 代理协议
│   │   └── protocols/     # 协议实现
│   │       ├── vmess/
│   │       ├── vless/
│   │       ├── trojan/
│   │       └── shadowsocks/
│   └── server/            # HTTP 服务器
├── pkg/                    # 公共包
│   └── errors/            # 错误处理
├── configs/               # 配置模板
├── deployments/           # 部署文件
│   ├── docker/            # Docker 配置
│   └── scripts/           # 部署脚本
├── scripts/               # 构建脚本
├── web/                   # 前端代码
│   └── src/
└── data/                  # 数据目录
```

## API 文档

### 认证
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/refresh` - 刷新令牌
- `POST /api/auth/logout` - 用户登出
- `GET /api/auth/me` - 获取当前用户

### 代理管理
- `GET /api/proxies` - 获取代理列表
- `POST /api/proxies` - 创建代理
- `GET /api/proxies/:id` - 获取代理详情
- `PUT /api/proxies/:id` - 更新代理
- `DELETE /api/proxies/:id` - 删除代理
- `GET /api/proxies/:id/link` - 获取分享链接

### 系统
- `GET /api/system/info` - 系统信息
- `GET /api/system/status` - 系统状态
- `GET /health` - 健康检查
- `GET /ready` - 就绪检查

## 特别鸣谢

- [Xray-core](https://github.com/XTLS/Xray-core) - 核心代理引擎
- [Vue.js](https://vuejs.org/) - 前端框架
- [Gin](https://gin-gonic.com/) - Web 框架
- [GORM](https://gorm.io/) - ORM 框架

## License

MIT License
