#!/bin/bash
# 数据库诊断脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 数据库路径
DB_PATH="${1:-data/v.db}"

if [ ! -f "$DB_PATH" ]; then
    echo -e "${RED}错误: 数据库文件不存在: $DB_PATH${NC}"
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}     V Panel 数据库诊断${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}数据库路径:${NC} $DB_PATH"
echo ""

# 检查 SQLite 是否安装
if ! command -v sqlite3 &> /dev/null; then
    echo -e "${RED}错误: sqlite3 未安装${NC}"
    echo "请安装 sqlite3: brew install sqlite (macOS) 或 apt-get install sqlite3 (Linux)"
    exit 1
fi

# 检查所有表
echo -e "${YELLOW}=== 数据库表列表 ===${NC}"
sqlite3 "$DB_PATH" ".tables"
echo ""

# 检查 IP 限制相关的表
echo -e "${YELLOW}=== IP 限制表检查 ===${NC}"

IP_TABLES=("ip_whitelist" "ip_blacklist" "active_ips" "ip_history" "subscription_ip_access" "geo_cache" "failed_attempts")

for table in "${IP_TABLES[@]}"; do
    if sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='$table';" | grep -q "$table"; then
        count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM $table;")
        echo -e "${GREEN}✓${NC} $table (记录数: $count)"
    else
        echo -e "${RED}✗${NC} $table (不存在)"
    fi
done
echo ""

# 检查迁移状态
echo -e "${YELLOW}=== 迁移状态 ===${NC}"
if sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='migrations';" | grep -q "migrations"; then
    echo -e "${GREEN}迁移表存在${NC}"
    echo ""
    echo "已应用的迁移:"
    sqlite3 "$DB_PATH" "SELECT version, name, applied_at FROM migrations ORDER BY version;" -header -column
else
    echo -e "${RED}迁移表不存在${NC}"
fi
echo ""

# 检查 IP 白名单表结构
echo -e "${YELLOW}=== IP 白名单表结构 ===${NC}"
if sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='ip_whitelist';" | grep -q "ip_whitelist"; then
    sqlite3 "$DB_PATH" "PRAGMA table_info(ip_whitelist);" -header -column
else
    echo -e "${RED}表不存在${NC}"
fi
echo ""

# 检查 IP 黑名单表结构
echo -e "${YELLOW}=== IP 黑名单表结构 ===${NC}"
if sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='ip_blacklist';" | grep -q "ip_blacklist"; then
    sqlite3 "$DB_PATH" "PRAGMA table_info(ip_blacklist);" -header -column
else
    echo -e "${RED}表不存在${NC}"
fi
echo ""

# 检查索引
echo -e "${YELLOW}=== IP 限制相关索引 ===${NC}"
sqlite3 "$DB_PATH" "SELECT name, tbl_name FROM sqlite_master WHERE type='index' AND (tbl_name LIKE 'ip_%' OR tbl_name LIKE '%_ip%') ORDER BY tbl_name, name;" -header -column
echo ""

# 数据库完整性检查
echo -e "${YELLOW}=== 数据库完整性检查 ===${NC}"
INTEGRITY=$(sqlite3 "$DB_PATH" "PRAGMA integrity_check;")
if [ "$INTEGRITY" = "ok" ]; then
    echo -e "${GREEN}✓ 数据库完整性正常${NC}"
else
    echo -e "${RED}✗ 数据库完整性问题:${NC}"
    echo "$INTEGRITY"
fi
echo ""

# 数据库大小
echo -e "${YELLOW}=== 数据库信息 ===${NC}"
DB_SIZE=$(du -h "$DB_PATH" | cut -f1)
echo -e "数据库大小: ${GREEN}$DB_SIZE${NC}"

# 页面大小和页面数
PAGE_SIZE=$(sqlite3 "$DB_PATH" "PRAGMA page_size;")
PAGE_COUNT=$(sqlite3 "$DB_PATH" "PRAGMA page_count;")
echo -e "页面大小: ${GREEN}$PAGE_SIZE bytes${NC}"
echo -e "页面数量: ${GREEN}$PAGE_COUNT${NC}"
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}     诊断完成${NC}"
echo -e "${BLUE}========================================${NC}"
