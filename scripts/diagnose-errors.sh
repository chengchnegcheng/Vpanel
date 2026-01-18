#!/bin/bash

# 错误诊断脚本
# 用于诊断 IP 限制和财务报表错误

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================="
echo "V Panel 错误诊断工具"
echo "========================================="
echo ""

# 数据库配置
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:-vpanel}"
DB_USER="${DB_USER:-root}"
DB_PASS="${DB_PASS:-}"

# 检查数据库连接
echo -e "${BLUE}1. 检查数据库连接${NC}"
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -e "SELECT 1;" 2>/dev/null; then
    echo -e "${GREEN}✓ 数据库连接正常${NC}"
else
    echo -e "${RED}✗ 数据库连接失败${NC}"
    echo "请检查数据库配置："
    echo "  DB_HOST=$DB_HOST"
    echo "  DB_PORT=$DB_PORT"
    echo "  DB_USER=$DB_USER"
    exit 1
fi
echo ""

# 检查数据库是否存在
echo -e "${BLUE}2. 检查数据库${NC}"
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -e "USE $DB_NAME;" 2>/dev/null; then
    echo -e "${GREEN}✓ 数据库 '$DB_NAME' 存在${NC}"
else
    echo -e "${RED}✗ 数据库 '$DB_NAME' 不存在${NC}"
    echo "创建数据库："
    echo "  mysql -u root -p -e \"CREATE DATABASE $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;\""
    exit 1
fi
echo ""

# 检查 IP 相关表
echo -e "${BLUE}3. 检查 IP 限制相关表${NC}"
IP_TABLES=("ip_whitelist" "ip_blacklist" "active_ips" "ip_history" "subscription_ip_access" "geo_cache" "failed_attempts")

for table in "${IP_TABLES[@]}"; do
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SHOW TABLES LIKE '$table';" 2>/dev/null | grep -q "$table"; then
        echo -e "${GREEN}✓ 表 '$table' 存在${NC}"
    else
        echo -e "${RED}✗ 表 '$table' 不存在${NC}"
    fi
done
echo ""

# 检查订单相关表
echo -e "${BLUE}4. 检查财务报表相关表${NC}"
ORDER_TABLES=("orders" "commercial_plans" "balance_transactions")

for table in "${ORDER_TABLES[@]}"; do
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SHOW TABLES LIKE '$table';" 2>/dev/null | grep -q "$table"; then
        echo -e "${GREEN}✓ 表 '$table' 存在${NC}"
        # 显示记录数
        count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -se "SELECT COUNT(*) FROM $table;" 2>/dev/null)
        echo "  记录数: $count"
    else
        echo -e "${RED}✗ 表 '$table' 不存在${NC}"
    fi
done
echo ""

# 检查礼品卡表
echo -e "${BLUE}5. 检查礼品卡相关表${NC}"
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SHOW TABLES LIKE 'gift_cards';" 2>/dev/null | grep -q "gift_cards"; then
    echo -e "${GREEN}✓ 表 'gift_cards' 存在${NC}"
    count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -se "SELECT COUNT(*) FROM gift_cards;" 2>/dev/null)
    echo "  记录数: $count"
else
    echo -e "${RED}✗ 表 'gift_cards' 不存在${NC}"
fi
echo ""

# 检查应用日志
echo -e "${BLUE}6. 检查应用日志${NC}"
LOG_FILE="/var/log/vpanel/app.log"
if [ -f "$LOG_FILE" ]; then
    echo -e "${GREEN}✓ 日志文件存在: $LOG_FILE${NC}"
    echo ""
    echo "最近的错误日志："
    tail -n 20 "$LOG_FILE" | grep -i "error" || echo "  没有发现错误"
else
    echo -e "${YELLOW}⚠ 日志文件不存在: $LOG_FILE${NC}"
    echo "检查其他可能的日志位置："
    echo "  - ./logs/app.log"
    echo "  - /tmp/vpanel.log"
    echo "  - journalctl -u vpanel"
fi
echo ""

# 检查服务状态
echo -e "${BLUE}7. 检查服务状态${NC}"
if systemctl is-active --quiet vpanel 2>/dev/null; then
    echo -e "${GREEN}✓ V Panel 服务正在运行${NC}"
    systemctl status vpanel --no-pager | head -n 10
elif pgrep -f "agent" > /dev/null; then
    echo -e "${GREEN}✓ V Panel 进程正在运行${NC}"
    ps aux | grep agent | grep -v grep
else
    echo -e "${RED}✗ V Panel 服务未运行${NC}"
fi
echo ""

# 测试 API 端点
echo -e "${BLUE}8. 测试 API 端点${NC}"
API_URL="${API_URL:-http://localhost:8080}"

# 测试健康检查
if curl -s -f "$API_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 健康检查端点正常${NC}"
    curl -s "$API_URL/health" | jq '.' 2>/dev/null || curl -s "$API_URL/health"
else
    echo -e "${RED}✗ 健康检查端点失败${NC}"
    echo "  URL: $API_URL/health"
fi
echo ""

# 生成诊断报告
echo "========================================="
echo "诊断总结"
echo "========================================="
echo ""

# 检查是否需要运行迁移
MISSING_TABLES=0
for table in "${IP_TABLES[@]}" "${ORDER_TABLES[@]}" "gift_cards"; do
    if ! mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SHOW TABLES LIKE '$table';" 2>/dev/null | grep -q "$table"; then
        MISSING_TABLES=$((MISSING_TABLES + 1))
    fi
done

if [ $MISSING_TABLES -gt 0 ]; then
    echo -e "${RED}发现 $MISSING_TABLES 个缺失的表${NC}"
    echo ""
    echo "建议操作："
    echo "1. 停止服务："
    echo "   systemctl stop vpanel"
    echo ""
    echo "2. 运行数据库迁移："
    echo "   ./agent migrate"
    echo "   # 或者重启服务（会自动运行迁移）"
    echo ""
    echo "3. 启动服务："
    echo "   systemctl start vpanel"
else
    echo -e "${GREEN}所有必需的表都存在${NC}"
    echo ""
    echo "如果仍然有错误，请检查："
    echo "1. 应用日志: tail -f /var/log/vpanel/app.log"
    echo "2. 数据库权限: GRANT ALL ON $DB_NAME.* TO '$DB_USER'@'%';"
    echo "3. 服务配置: cat configs/config.yaml"
fi

echo ""
echo "========================================="
echo "诊断完成"
echo "========================================="
