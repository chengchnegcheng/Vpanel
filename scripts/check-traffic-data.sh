#!/bin/bash
# 检查流量数据

echo "=== 检查流量数据 ==="
echo ""

# 读取数据库配置
if [ -f "configs/config.yaml" ]; then
    DB_HOST=$(grep "host:" configs/config.yaml | awk '{print $2}' | tr -d '"')
    DB_PORT=$(grep "port:" configs/config.yaml | awk '{print $2}')
    DB_USER=$(grep "user:" configs/config.yaml | awk '{print $2}' | tr -d '"')
    DB_NAME=$(grep "dbname:" configs/config.yaml | awk '{print $2}' | tr -d '"')
else
    echo "配置文件不存在，使用默认值"
    DB_HOST="localhost"
    DB_PORT="5432"
    DB_USER="vpanel"
    DB_NAME="vpanel"
fi

echo "数据库连接信息:"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  用户: $DB_USER"
echo "  数据库: $DB_NAME"
echo ""

# 检查流量表
echo "1. 检查流量表结构..."
PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\d traffic" 2>/dev/null

if [ $? -ne 0 ]; then
    echo "❌ 无法连接到数据库或流量表不存在"
    echo ""
    echo "请确保:"
    echo "1. PostgreSQL 正在运行"
    echo "2. 数据库配置正确"
    echo "3. 已运行数据库迁移"
    exit 1
fi

echo ""
echo "2. 检查流量数据..."
PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
-- 总记录数
SELECT COUNT(*) as total_records FROM traffic;

-- 按用户统计
SELECT 
    user_id,
    COUNT(*) as records,
    pg_size_pretty(SUM(upload)::bigint) as total_upload,
    pg_size_pretty(SUM(download)::bigint) as total_download,
    MIN(recorded_at) as earliest,
    MAX(recorded_at) as latest
FROM traffic
GROUP BY user_id
ORDER BY user_id
LIMIT 10;

-- 最近7天的数据
SELECT 
    DATE(recorded_at) as date,
    COUNT(*) as records,
    pg_size_pretty(SUM(upload)::bigint) as upload,
    pg_size_pretty(SUM(download)::bigint) as download
FROM traffic
WHERE recorded_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(recorded_at)
ORDER BY date DESC;
EOF

echo ""
echo "3. 如果没有数据，可以运行以下命令添加测试数据:"
echo "   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f scripts/add-test-traffic-data.sql"
echo ""
