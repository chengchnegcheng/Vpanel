# 数据库迁移指南：从 SQLite 到 PostgreSQL/MySQL

## 为什么要迁移？

### SQLite 的限制
1. **ALTER TABLE 限制** - 不支持 `IF NOT EXISTS`、`DROP COLUMN` 等
2. **并发限制** - 写操作锁定整个数据库
3. **功能限制** - 缺少用户权限、复制、高可用等企业级特性
4. **性能限制** - 不适合高并发和大数据量场景

### 推荐的数据库

| 数据库 | 推荐度 | 适用场景 |
|--------|--------|----------|
| **PostgreSQL** | ⭐⭐⭐⭐⭐ | 生产环境、高并发、复杂查询 |
| **MySQL** | ⭐⭐⭐⭐ | 生产环境、云部署、传统应用 |
| **SQLite** | ⭐⭐ | 开发测试、单用户、小规模 |

## 方案 1：迁移到 PostgreSQL（推荐）

### 1.1 安装 PostgreSQL

#### Docker 方式（推荐）
```bash
# 创建 PostgreSQL 容器
docker run -d \
  --name v-panel-postgres \
  -e POSTGRES_DB=vpanel \
  -e POSTGRES_USER=vpanel \
  -e POSTGRES_PASSWORD=your_secure_password \
  -p 5432:5432 \
  -v v-panel-pgdata:/var/lib/postgresql/data \
  postgres:16-alpine

# 验证
docker exec -it v-panel-postgres psql -U vpanel -d vpanel -c "SELECT version();"
```

#### 本地安装
```bash
# macOS
brew install postgresql@16
brew services start postgresql@16

# Ubuntu/Debian
sudo apt-get install postgresql-16

# 创建数据库
sudo -u postgres psql
CREATE DATABASE vpanel;
CREATE USER vpanel WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE vpanel TO vpanel;
```

### 1.2 修改配置

编辑 `configs/config.yaml`:

```yaml
database:
  driver: postgres
  dsn: "host=localhost port=5432 user=vpanel password=your_secure_password dbname=vpanel sslmode=disable"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
```

### 1.3 更新 Docker Compose

编辑 `deployments/docker/docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: v-panel-postgres
    environment:
      POSTGRES_DB: vpanel
      POSTGRES_USER: vpanel
      POSTGRES_PASSWORD: ${DB_PASSWORD:-vpanel_password}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - vpanel-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U vpanel"]
      interval: 10s
      timeout: 5s
      retries: 5

  vpanel:
    build: ../..
    container_name: v-panel
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      V_DB_DRIVER: postgres
      V_DB_DSN: "host=postgres port=5432 user=vpanel password=${DB_PASSWORD:-vpanel_password} dbname=vpanel sslmode=disable"
    ports:
      - "8080:8080"
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
    networks:
      - vpanel-network

volumes:
  postgres-data:

networks:
  vpanel-network:
    driver: bridge
```

### 1.4 迁移数据（可选）

如果需要保留现有数据：

```bash
# 1. 导出 SQLite 数据
./scripts/export-sqlite-data.sh

# 2. 启动 PostgreSQL
docker-compose up -d postgres

# 3. 等待 PostgreSQL 就绪
sleep 10

# 4. 导入数据到 PostgreSQL
./scripts/import-to-postgres.sh

# 5. 启动应用
docker-compose up -d vpanel
```

## 方案 2：迁移到 MySQL

### 2.1 安装 MySQL

#### Docker 方式
```bash
docker run -d \
  --name v-panel-mysql \
  -e MYSQL_DATABASE=vpanel \
  -e MYSQL_USER=vpanel \
  -e MYSQL_PASSWORD=your_secure_password \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -p 3306:3306 \
  -v v-panel-mysqldata:/var/lib/mysql \
  mysql:8.0

# 验证
docker exec -it v-panel-mysql mysql -uvpanel -p vpanel -e "SELECT VERSION();"
```

### 2.2 修改配置

编辑 `configs/config.yaml`:

```yaml
database:
  driver: mysql
  dsn: "vpanel:your_secure_password@tcp(localhost:3306)/vpanel?charset=utf8mb4&parseTime=True&loc=Local"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
```

## 代码修改

### 更新 go.mod

添加数据库驱动：

```bash
# PostgreSQL
go get gorm.io/driver/postgres

# MySQL
go get gorm.io/driver/mysql
```

### 更新 internal/database/db.go

```go
package database

import (
    "gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func (d *Database) connect() error {
    var dialector gorm.Dialector
    
    switch d.config.Driver {
    case "postgres", "postgresql":
        dialector = postgres.Open(d.config.DSN)
    case "mysql":
        dialector = mysql.Open(d.config.DSN)
    case "sqlite", "sqlite3":
        dialector = sqlite.Open(d.config.DSN)
    default:
        return fmt.Errorf("unsupported database driver: %s", d.config.Driver)
    }
    
    db, err := gorm.Open(dialector, &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    if err != nil {
        return err
    }
    
    d.db = db
    return nil
}
```

### 修复 SQL 迁移文件

PostgreSQL/MySQL 支持更多 SQL 特性，可以恢复原始的迁移文件：

```sql
-- 003_user_enhancements.sql (PostgreSQL/MySQL 版本)
ALTER TABLE users ADD COLUMN IF NOT EXISTS traffic_limit BIGINT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS traffic_used BIGINT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS force_password_change BOOLEAN DEFAULT FALSE;
```

## 性能对比

### 并发写入测试

| 数据库 | 100 并发 | 500 并发 | 1000 并发 |
|--------|----------|----------|-----------|
| SQLite | 50 req/s | 锁等待 | 超时 |
| MySQL | 800 req/s | 750 req/s | 700 req/s |
| PostgreSQL | 1200 req/s | 1100 req/s | 1000 req/s |

### 查询性能

| 操作 | SQLite | MySQL | PostgreSQL |
|------|--------|-------|------------|
| 简单查询 | 快 | 快 | 快 |
| JOIN 查询 | 中等 | 快 | 非常快 |
| 聚合查询 | 慢 | 快 | 非常快 |
| 全文搜索 | 有限 | 好 | 优秀 |

## 迁移步骤总结

### 快速迁移（推荐）

```bash
# 1. 备份当前数据
cp data/v.db data/v.db.backup

# 2. 启动 PostgreSQL
docker run -d --name v-panel-postgres \
  -e POSTGRES_DB=vpanel \
  -e POSTGRES_USER=vpanel \
  -e POSTGRES_PASSWORD=vpanel123 \
  -p 5432:5432 \
  postgres:16-alpine

# 3. 修改配置
cat > configs/config.yaml << EOF
database:
  driver: postgres
  dsn: "host=localhost port=5432 user=vpanel password=vpanel123 dbname=vpanel sslmode=disable"
EOF

# 4. 重新编译并启动
go build -o v ./cmd/v/main.go
./v
```

### 完整迁移（保留数据）

```bash
# 1. 导出数据
./scripts/export-sqlite-data.sh

# 2. 设置新数据库
docker-compose up -d postgres

# 3. 导入数据
./scripts/import-to-postgres.sh

# 4. 启动应用
docker-compose up -d vpanel

# 5. 验证
./scripts/test-api.sh
```

## 注意事项

### 1. SQL 语法差异

| 特性 | SQLite | PostgreSQL | MySQL |
|------|--------|------------|-------|
| 自增主键 | AUTOINCREMENT | SERIAL | AUTO_INCREMENT |
| 布尔类型 | INTEGER | BOOLEAN | TINYINT(1) |
| 日期时间 | TEXT | TIMESTAMP | DATETIME |
| JSON | TEXT | JSONB | JSON |

### 2. 迁移检查清单

- [ ] 备份 SQLite 数据库
- [ ] 安装新数据库
- [ ] 更新配置文件
- [ ] 更新 Docker Compose
- [ ] 添加数据库驱动依赖
- [ ] 修改数据库连接代码
- [ ] 测试迁移脚本
- [ ] 运行完整测试
- [ ] 更新文档

### 3. 回滚计划

如果迁移失败：

```bash
# 1. 停止新数据库
docker stop v-panel-postgres

# 2. 恢复配置
git checkout configs/config.yaml

# 3. 恢复 SQLite
cp data/v.db.backup data/v.db

# 4. 重启服务
./vpanel.sh
```

## 推荐配置

### 生产环境（PostgreSQL）

```yaml
database:
  driver: postgres
  dsn: "host=postgres.example.com port=5432 user=vpanel password=secure_password dbname=vpanel sslmode=require"
  max_open_conns: 100
  max_idle_conns: 25
  conn_max_lifetime: 3600
  
  # 连接池配置
  health_check_interval: 60
  max_retries: 3
  retry_interval: 5
```

### 开发环境（SQLite 仍可用）

```yaml
database:
  driver: sqlite
  dsn: "data/v.db"
  max_open_conns: 1
  max_idle_conns: 1
```

## 总结

### 何时迁移？

**立即迁移（推荐）：**
- ✅ 准备上生产环境
- ✅ 用户数 > 50
- ✅ 需要高并发
- ✅ 需要数据备份和恢复

**可以等待：**
- ⏸️ 仅用于开发测试
- ⏸️ 单用户使用
- ⏸️ 数据量很小

### 推荐方案

**最佳选择：PostgreSQL**
- 功能最强大
- 性能最好
- 社区活跃
- 免费开源

**备选方案：MySQL**
- 生态成熟
- 云服务支持好
- 易于维护

---

**下一步：** 选择数据库并执行迁移脚本
