# 故障排查指南

## 错误 ID: ERR-MKIMADZT-W501D2

这个错误表明 IP 限制和财务报表功能仍然无法正常工作。

## 快速诊断

运行诊断脚本：

```bash
# 设置数据库配置
export DB_HOST=localhost
export DB_PORT=3306
export DB_NAME=vpanel
export DB_USER=root
export DB_PASS=your_password

# 运行诊断
./scripts/diagnose-errors.sh
```

## 常见问题和解决方案

### 问题 1: 数据库表不存在

**症状**:
- API 返回 500 错误
- 日志显示 "Table doesn't exist" 错误

**解决方案**:

```bash
# 方法 1: 重启服务（自动运行迁移）
systemctl restart vpanel

# 方法 2: 手动运行迁移
./agent migrate

# 方法 3: 使用 SQL 手动创建表
mysql -u root -p vpanel < scripts/create-tables.sql
```

### 问题 2: IP 服务初始化失败

**症状**:
- IP 限制页面显示 "IP restriction service is not available"
- 日志显示 "Failed to create IP service"

**解决方案**:

1. 检查数据库连接：
```bash
mysql -u root -p -e "USE vpanel; SHOW TABLES;"
```

2. 验证 IP 相关表存在：
```sql
SHOW TABLES LIKE 'ip_%';
SHOW TABLES LIKE 'active_ips';
SHOW TABLES LIKE 'geo_cache';
SHOW TABLES LIKE 'failed_attempts';
```

3. 如果表不存在，运行迁移：
```bash
systemctl stop vpanel
./agent migrate
systemctl start vpanel
```

### 问题 3: 财务报表查询失败

**症状**:
- 财务报表页面显示错误
- 日志显示 "Failed to get revenue" 或 "Failed to get order count"

**解决方案**:

1. 检查订单表：
```sql
USE vpanel;
SHOW TABLES LIKE 'orders';
DESC orders;
SELECT COUNT(*) FROM orders;
```

2. 检查数据库权限：
```sql
SHOW GRANTS FOR 'vpanel_user'@'localhost';
```

3. 测试查询：
```sql
SELECT 
    SUM(amount) as revenue,
    COUNT(*) as order_count
FROM orders
WHERE status = 'paid'
AND created_at >= '2024-01-01'
AND created_at <= '2024-12-31';
```

### 问题 4: 礼品卡列表查询失败

**症状**:
- 礼品卡页面无法加载
- 日志显示 "Failed to list gift cards"

**解决方案**:

1. 检查礼品卡表：
```sql
USE vpanel;
SHOW TABLES LIKE 'gift_cards';
DESC gift_cards;
SELECT COUNT(*) FROM gift_cards;
```

2. 测试查询：
```sql
SELECT * FROM gift_cards LIMIT 10;
```

## 手动创建缺失的表

如果自动迁移失败，可以手动创建表：

### IP 限制相关表

```sql
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
```

## 检查日志

### 查看应用日志

```bash
# 实时查看日志
tail -f /var/log/vpanel/app.log

# 查看错误日志
tail -f /var/log/vpanel/app.log | grep -i error

# 查看最近 100 行
tail -n 100 /var/log/vpanel/app.log

# 使用 journalctl（如果使用 systemd）
journalctl -u vpanel -f
journalctl -u vpanel -n 100
```

### 查看数据库日志

```bash
# MySQL 错误日志
tail -f /var/log/mysql/error.log

# MySQL 慢查询日志
tail -f /var/log/mysql/slow-query.log
```

## 测试修复

### 1. 测试 IP 限制

```bash
export TOKEN="your_admin_token"

# 测试 IP 统计
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/ip-restrictions/stats

# 预期响应
{
  "code": 200,
  "message": "success",
  "data": {
    "total_active_ips": 0,
    "total_blacklisted": 0,
    "total_whitelisted": 0,
    "settings": {...}
  }
}
```

### 2. 测试财务报表

```bash
# 测试收入报表
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/admin/reports/revenue?start=2024-01-01&end=2024-12-31"

# 预期响应
{
  "code": 200,
  "message": "success",
  "data": {
    "revenue": 0,
    "order_count": 0,
    "start": "2024-01-01",
    "end": "2024-12-31"
  }
}

# 测试订单统计
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/reports/orders

# 预期响应
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 0,
    "pending": 0,
    "paid": 0,
    "completed": 0,
    "cancelled": 0,
    "refunded": 0
  }
}
```

### 3. 测试礼品卡

```bash
# 测试礼品卡列表
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/admin/gift-cards?page=1&page_size=20"

# 预期响应
{
  "code": 200,
  "message": "success",
  "data": {
    "gift_cards": [],
    "total": 0,
    "page": 1,
    "page_size": 20
  }
}
```

## 如果问题仍然存在

### 1. 收集诊断信息

```bash
# 创建诊断报告
./scripts/diagnose-errors.sh > diagnosis-report.txt 2>&1

# 收集日志
tail -n 500 /var/log/vpanel/app.log > app-log.txt

# 导出数据库结构
mysqldump -u root -p --no-data vpanel > schema.sql
```

### 2. 重新初始化

如果所有方法都失败，可以尝试重新初始化：

```bash
# 备份数据
mysqldump -u root -p vpanel > backup-$(date +%Y%m%d).sql

# 停止服务
systemctl stop vpanel

# 删除并重新创建数据库
mysql -u root -p -e "DROP DATABASE IF EXISTS vpanel;"
mysql -u root -p -e "CREATE DATABASE vpanel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 启动服务（会自动运行迁移）
systemctl start vpanel

# 检查日志
tail -f /var/log/vpanel/app.log
```

### 3. 联系支持

提供以下信息：
1. 诊断报告 (`diagnosis-report.txt`)
2. 应用日志 (`app-log.txt`)
3. 数据库结构 (`schema.sql`)
4. 错误 ID (ERR-MKIMADZT-W501D2)
5. 操作系统和数据库版本

## 预防措施

### 1. 定期备份

```bash
# 创建备份脚本
cat > /usr/local/bin/vpanel-backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/var/backups/vpanel"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p "$BACKUP_DIR"
mysqldump -u root -p vpanel | gzip > "$BACKUP_DIR/vpanel-$DATE.sql.gz"
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +7 -delete
EOF

chmod +x /usr/local/bin/vpanel-backup.sh

# 添加到 crontab
echo "0 2 * * * /usr/local/bin/vpanel-backup.sh" | crontab -
```

### 2. 监控日志

```bash
# 设置日志告警
cat > /usr/local/bin/vpanel-monitor.sh << 'EOF'
#!/bin/bash
LOG_FILE="/var/log/vpanel/app.log"
ERROR_COUNT=$(tail -n 100 "$LOG_FILE" | grep -c "ERROR")
if [ "$ERROR_COUNT" -gt 10 ]; then
    echo "警告: 发现 $ERROR_COUNT 个错误" | mail -s "V Panel 错误告警" admin@example.com
fi
EOF

chmod +x /usr/local/bin/vpanel-monitor.sh
echo "*/5 * * * * /usr/local/bin/vpanel-monitor.sh" | crontab -
```

### 3. 健康检查

```bash
# 创建健康检查脚本
cat > /usr/local/bin/vpanel-health.sh << 'EOF'
#!/bin/bash
if ! curl -sf http://localhost:8080/health > /dev/null; then
    echo "V Panel 健康检查失败，尝试重启..."
    systemctl restart vpanel
    sleep 5
    if ! curl -sf http://localhost:8080/health > /dev/null; then
        echo "重启后仍然失败" | mail -s "V Panel 服务异常" admin@example.com
    fi
fi
EOF

chmod +x /usr/local/bin/vpanel-health.sh
echo "*/5 * * * * /usr/local/bin/vpanel-health.sh" | crontab -
```
