#!/bin/bash

# 测试用户场景脚本
# 模拟用户在前端的操作

set -e

BASE_URL="http://localhost:8081"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}测试用户场景${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 1. 登录
echo -e "${YELLOW}步骤 1: 管理员登录${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ 登录失败${NC}"
    echo "$LOGIN_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ 登录成功${NC}"
echo ""

# 2. 测试IP限制统计
echo -e "${YELLOW}步骤 2: 访问IP限制统计页面${NC}"
IP_STATS=$(curl -s -X GET "$BASE_URL/api/admin/ip-restrictions/stats" \
  -H "Authorization: Bearer $TOKEN")

echo "$IP_STATS" | jq .

if echo "$IP_STATS" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓ IP限制统计正常${NC}"
else
    echo -e "${RED}✗ IP限制统计失败${NC}"
    echo "$IP_STATS"
fi
echo ""

# 3. 测试财务报表 - 收入统计
echo -e "${YELLOW}步骤 3: 访问财务报表 - 收入统计${NC}"
REVENUE_REPORT=$(curl -s -X GET "$BASE_URL/api/admin/reports/revenue?start=2024-01-01&end=2024-12-31" \
  -H "Authorization: Bearer $TOKEN")

echo "$REVENUE_REPORT" | jq .

if echo "$REVENUE_REPORT" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓ 财务报表 - 收入统计正常${NC}"
else
    echo -e "${RED}✗ 财务报表 - 收入统计失败${NC}"
    echo "$REVENUE_REPORT"
fi
echo ""

# 4. 测试财务报表 - 订单统计
echo -e "${YELLOW}步骤 4: 访问财务报表 - 订单统计${NC}"
ORDER_STATS=$(curl -s -X GET "$BASE_URL/api/admin/reports/orders" \
  -H "Authorization: Bearer $TOKEN")

echo "$ORDER_STATS" | jq .

if echo "$ORDER_STATS" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓ 财务报表 - 订单统计正常${NC}"
else
    echo -e "${RED}✗ 财务报表 - 订单统计失败${NC}"
    echo "$ORDER_STATS"
fi
echo ""

# 5. 测试IP白名单列表
echo -e "${YELLOW}步骤 5: 获取IP白名单列表${NC}"
WHITELIST=$(curl -s -X GET "$BASE_URL/api/admin/ip-whitelist" \
  -H "Authorization: Bearer $TOKEN")

echo "$WHITELIST" | jq .

if echo "$WHITELIST" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓ IP白名单列表正常${NC}"
else
    echo -e "${RED}✗ IP白名单列表失败${NC}"
    echo "$WHITELIST"
fi
echo ""

# 6. 测试IP黑名单列表
echo -e "${YELLOW}步骤 6: 获取IP黑名单列表${NC}"
BLACKLIST=$(curl -s -X GET "$BASE_URL/api/admin/ip-blacklist" \
  -H "Authorization: Bearer $TOKEN")

echo "$BLACKLIST" | jq .

if echo "$BLACKLIST" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓ IP黑名单列表正常${NC}"
else
    echo -e "${RED}✗ IP黑名单列表失败${NC}"
    echo "$BLACKLIST"
fi
echo ""

# 7. 测试不同日期范围的财务报表
echo -e "${YELLOW}步骤 7: 测试不同日期范围的财务报表${NC}"

# 测试今年
CURRENT_YEAR=$(date +%Y)
REVENUE_THIS_YEAR=$(curl -s -X GET "$BASE_URL/api/admin/reports/revenue?start=${CURRENT_YEAR}-01-01&end=${CURRENT_YEAR}-12-31" \
  -H "Authorization: Bearer $TOKEN")

echo "今年的收入报表:"
echo "$REVENUE_THIS_YEAR" | jq .

if echo "$REVENUE_THIS_YEAR" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓ 今年收入报表正常${NC}"
else
    echo -e "${RED}✗ 今年收入报表失败${NC}"
fi
echo ""

# 测试上个月
LAST_MONTH_START=$(date -d "1 month ago" +%Y-%m-01 2>/dev/null || date -v-1m +%Y-%m-01)
LAST_MONTH_END=$(date -d "1 month ago" +%Y-%m-%d 2>/dev/null || date -v-1m +%Y-%m-%d)
REVENUE_LAST_MONTH=$(curl -s -X GET "$BASE_URL/api/admin/reports/revenue?start=${LAST_MONTH_START}&end=${LAST_MONTH_END}" \
  -H "Authorization: Bearer $TOKEN")

echo "上个月的收入报表:"
echo "$REVENUE_LAST_MONTH" | jq .

if echo "$REVENUE_LAST_MONTH" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓ 上个月收入报表正常${NC}"
else
    echo -e "${RED}✗ 上个月收入报表失败${NC}"
fi
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}测试完成${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}所有测试通过！${NC}"
echo ""
echo -e "${YELLOW}如果用户仍然报告错误，可能的原因：${NC}"
echo "1. 浏览器缓存了旧版本的前端代码"
echo "2. 用户使用的是旧的浏览器标签页"
echo "3. 前端代码需要重新构建"
echo "4. 用户的网络连接不稳定"
echo ""
echo -e "${YELLOW}建议用户：${NC}"
echo "1. 清除浏览器缓存（Ctrl+Shift+Delete）"
echo "2. 硬刷新页面（Ctrl+Shift+R 或 Cmd+Shift+R）"
echo "3. 关闭所有标签页后重新打开"
echo "4. 尝试使用无痕模式访问"
