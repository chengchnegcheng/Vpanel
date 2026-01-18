# 立即修复步骤

## 错误 ID: ERR-MKIMADZT-W501D2

这个错误表明数据库表可能缺失或服务初始化失败。

## 快速修复（5 分钟）

### 步骤 1: 运行诊断

```bash
# 设置数据库密码
export DB_PASS="your_database_password"

# 运行诊断脚本
chmod +x scripts/diagnose-errors.sh
./scripts/diagnose-errors.sh
```

### 步骤 2: 根据诊断结果修复

#### 情况 A: 表缺失

如果诊断显示表缺失：

```bash
# 停止服务
systemctl stop vpanel
# 或
./vpanel.sh stop

# 重新编译（包含最新修复）
go build -o agent cmd/agent/main.go

# 启动服务（会自动运行迁移）
systemctl start vpanel
# 或
./vpanel.sh start

# 等待 10 秒让迁移完成
sleep 10

# 检查日志
tail -f /var/log/vpanel/app.log
```

#### 情况 B: 服务未运行

如果服务未运行：

```bash
# 检查配置文件
cat configs/config.yaml

# 启动服务
systemctl start vpanel
# 或
./vpanel.sh start

# 检查状态
systemctl status vpanel
# 或
./vpanel.sh status
```

#### 情况 C: 数据库连接失败

如果数据库连接失败：

```bash
# 检查数据库服务
systemctl status mysql
# 或
systemctl status mariadb

# 测试连接
mysql -u root -p -e "SELECT 1;"

# 检查配置
cat configs/config.yaml | grep -A 10 database
```

### 步骤 3: 验证修复

```bash
# 设置管理员 token
export TOKEN="your_admin_token"

# 测试 IP 限制
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/ip-restrictions/stats

# 测试财务报表
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/reports/orders

# 测试礼品卡
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/admin/gift-cards?page=1&page_size=20"
```

## 如果快速修复失败

### 完整重新初始化

```bash
# 1. 备份数据
mysqldump -u root -p vpanel > backup-$(date +%Y%m%d).sql

# 2. 停止服务
systemctl stop vpanel

# 3. 删除旧的二进制文件
rm -f agent

# 4. 重新编译
go clean
go build -o agent cmd/agent/main.go

# 5. 重新创建数据库（可选，如果表结构有问题）
mysql -u root -p << EOF
DROP DATABASE IF EXISTS vpanel;
CREATE DATABASE vpanel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EOF

# 6. 启动服务
systemctl start vpanel

# 7. 监控日志
tail -f /var/log/vpanel/app.log
```

## 手动创建缺失的表

如果自动迁移失败，手动创建表：

```bash
# 下载 SQL 脚本
cat > /tmp/create-ip-tables.sql << 'EOF'
-- IP 白名单
CREATE TABLE IF NOT EXISTS `ip_whitelist` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ip` varchar(45) NOT NULL,
  `cidr` varchar(50) DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `created_by` bigint unsigned NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ip_whitelist_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- IP 黑名单
CREATE TABLE IF NOT EXISTS `ip_blacklist` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ip` varchar(45) NOT NULL,
  `cidr` varchar(50) DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL,
  `reason` varchar(255) DEFAULT NULL,
  `expires_at` datetime DEFAULT NULL,
  `is_automatic` tinyint(1) DEFAULT '0',
  `created_by` bigint unsigned DEFAULT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ip_blacklist_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 活跃 IP
CREATE TABLE IF NOT EXISTS `active_ips` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `ip` varchar(45) NOT NULL,
  `user_agent` varchar(500) DEFAULT NULL,
  `device_type` varchar(50) DEFAULT NULL,
  `country` varchar(100) DEFAULT NULL,
  `city` varchar(100) DEFAULT NULL,
  `last_active` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_active_ip_user_ip` (`user_id`,`ip`),
  KEY `idx_active_ips_last_active` (`last_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- IP 历史
CREATE TABLE IF NOT EXISTS `ip_history` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `ip` varchar(45) NOT NULL,
  `user_agent` varchar(500) DEFAULT NULL,
  `access_type` varchar(20) DEFAULT NULL,
  `country` varchar(100) DEFAULT NULL,
  `city` varchar(100) DEFAULT NULL,
  `is_suspicious` tinyint(1) DEFAULT '0',
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ip_history_user_time` (`user_id`,`created_at`),
  KEY `idx_ip_history_ip` (`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 订阅 IP 访问
CREATE TABLE IF NOT EXISTS `subscription_ip_access` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `subscription_id` bigint unsigned NOT NULL,
  `ip` varchar(45) NOT NULL,
  `user_agent` varchar(500) DEFAULT NULL,
  `country` varchar(100) DEFAULT NULL,
  `access_count` int NOT NULL DEFAULT '0',
  `first_access` datetime NOT NULL,
  `last_access` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_sub_ip_access` (`subscription_id`,`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 地理位置缓存
CREATE TABLE IF NOT EXISTS `geo_cache` (
  `ip` varchar(45) NOT NULL,
  `country` varchar(100) DEFAULT NULL,
  `country_code` varchar(2) DEFAULT NULL,
  `region` varchar(100) DEFAULT NULL,
  `city` varchar(100) DEFAULT NULL,
  `latitude` double DEFAULT NULL,
  `longitude` double DEFAULT NULL,
  `isp` varchar(200) DEFAULT NULL,
  `cached_at` datetime NOT NULL,
  PRIMARY KEY (`ip`),
  KEY `idx_geo_cache_cached_at` (`cached_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 失败尝试
CREATE TABLE IF NOT EXISTS `failed_attempts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ip` varchar(45) NOT NULL,
  `reason` varchar(255) DEFAULT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_failed_attempts_ip` (`ip`),
  KEY `idx_failed_attempts_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
EOF

# 执行 SQL
mysql -u root -p vpanel < /tmp/create-ip-tables.sql

# 验证表创建
mysql -u root -p vpanel -e "SHOW TABLES LIKE 'ip_%'; SHOW TABLES LIKE 'active_ips'; SHOW TABLES LIKE 'geo_cache'; SHOW TABLES LIKE 'failed_attempts';"
```

## 获取管理员 Token

如果没有管理员 token：

```bash
# 方法 1: 从浏览器开发者工具获取
# 1. 打开浏览器开发者工具 (F12)
# 2. 切换到 Application/Storage 标签
# 3. 查看 Local Storage
# 4. 找到 'token' 或 'userToken' 键

# 方法 2: 使用 API 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your_password"
  }'

# 从响应中提取 token
```

## 常见错误和解决方案

### 错误: "IP restriction service is not available"

**原因**: IP 服务初始化失败

**解决**:
```bash
# 检查数据库表
mysql -u root -p vpanel -e "SHOW TABLES LIKE 'ip_%';"

# 如果表不存在，运行迁移
systemctl restart vpanel
```

### 错误: "Failed to get revenue"

**原因**: 订单表不存在或查询失败

**解决**:
```bash
# 检查订单表
mysql -u root -p vpanel -e "SHOW TABLES LIKE 'orders';"

# 测试查询
mysql -u root -p vpanel -e "SELECT COUNT(*) FROM orders;"
```

### 错误: "Failed to list gift cards"

**原因**: 礼品卡表不存在或查询失败

**解决**:
```bash
# 检查礼品卡表
mysql -u root -p vpanel -e "SHOW TABLES LIKE 'gift_cards';"

# 测试查询
mysql -u root -p vpanel -e "SELECT COUNT(*) FROM gift_cards;"
```

## 紧急联系

如果以上步骤都无法解决问题，请提供：

1. 诊断报告输出
2. 应用日志最后 100 行
3. 数据库表列表
4. 错误截图

```bash
# 生成完整诊断报告
./scripts/diagnose-errors.sh > diagnosis-$(date +%Y%m%d_%H%M%S).txt 2>&1
tail -n 100 /var/log/vpanel/app.log > app-log-$(date +%Y%m%d_%H%M%S).txt
mysql -u root -p vpanel -e "SHOW TABLES;" > tables-$(date +%Y%m%d_%H%M%S).txt
```

将这些文件发送给技术支持。
