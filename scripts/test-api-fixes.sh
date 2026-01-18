#!/bin/bash

# API 错误修复测试脚本
# 用于测试三个已修复的 API 端点

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
API_BASE_URL="${API_BASE_URL:-http://localhost:8080/api}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"

# 检查是否提供了 token
if [ -z "$ADMIN_TOKEN" ]; then
    echo -e "${RED}错误: 请设置 ADMIN_TOKEN 环境变量${NC}"
    echo "用法: ADMIN_TOKEN=your_token $0"
    exit 1
fi

echo "========================================="
echo "API 错误修复测试"
echo "========================================="
echo ""

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local expected_code=$4
    
    echo -e "${YELLOW}测试: $name${NC}"
    echo "请求: $method $endpoint"
    
    response=$(curl -s -w "\n%{http_code}" -X "$method" \
        "$API_BASE_URL$endpoint" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq "$expected_code" ]; then
        echo -e "${GREEN}✓ 成功 (HTTP $http_code)${NC}"
        echo "响应: $body" | jq '.' 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ 失败 (期望 HTTP $expected_code, 实际 HTTP $http_code)${NC}"
        echo "响应: $body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi
    echo ""
}

# 测试 1: IP 限制管理统计
echo "========================================="
echo "测试 1: IP 限制管理统计"
echo "========================================="
test_api "获取 IP 限制统计" "GET" "/admin/ip-restrictions/stats" 200 || true

# 测试 2: 财务报表 - 收入统计
echo "========================================="
echo "测试 2: 财务报表 - 收入统计"
echo "========================================="
START_DATE=$(date -d "30 days ago" +%Y-%m-%d 2>/dev/null || date -v-30d +%Y-%m-%d)
END_DATE=$(date +%Y-%m-%d)
test_api "获取收入报表" "GET" "/admin/reports/revenue?start=$START_DATE&end=$END_DATE" 200 || true

# 测试 3: 财务报表 - 订单统计
echo "========================================="
echo "测试 3: 财务报表 - 订单统计"
echo "========================================="
test_api "获取订单统计" "GET" "/admin/reports/orders" 200 || true

# 测试 4: 礼品卡列表
echo "========================================="
echo "测试 4: 礼品卡列表"
echo "========================================="
test_api "获取礼品卡列表" "GET" "/admin/gift-cards?page=1&page_size=20" 200 || true

# 测试 5: 礼品卡统计
echo "========================================="
echo "测试 5: 礼品卡统计"
echo "========================================="
test_api "获取礼品卡统计" "GET" "/admin/gift-cards/stats" 200 || true

# 测试 6: IP 白名单
echo "========================================="
echo "测试 6: IP 白名单"
echo "========================================="
test_api "获取 IP 白名单" "GET" "/admin/ip-whitelist" 200 || true

# 测试 7: IP 黑名单
echo "========================================="
echo "测试 7: IP 黑名单"
echo "========================================="
test_api "获取 IP 黑名单" "GET" "/admin/ip-blacklist" 200 || true

# 测试错误处理
echo "========================================="
echo "测试错误处理"
echo "========================================="

# 测试无效日期格式
echo -e "${YELLOW}测试: 无效日期格式${NC}"
response=$(curl -s -w "\n%{http_code}" -X GET \
    "$API_BASE_URL/admin/reports/revenue?start=invalid-date" \
    -H "Authorization: Bearer $ADMIN_TOKEN")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 400 ]; then
    echo -e "${GREEN}✓ 正确返回 400 错误${NC}"
    echo "响应: $body" | jq '.' 2>/dev/null || echo "$body"
else
    echo -e "${RED}✗ 期望 400 错误，实际 HTTP $http_code${NC}"
fi
echo ""

# 测试无效分页参数
echo -e "${YELLOW}测试: 无效分页参数（自动修正）${NC}"
response=$(curl -s -w "\n%{http_code}" -X GET \
    "$API_BASE_URL/admin/gift-cards?page=-1&page_size=1000" \
    -H "Authorization: Bearer $ADMIN_TOKEN")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 200 ]; then
    echo -e "${GREEN}✓ 成功处理无效参数并返回结果${NC}"
    echo "响应: $body" | jq '.data.page, .data.page_size' 2>/dev/null || echo "$body"
else
    echo -e "${RED}✗ 失败 HTTP $http_code${NC}"
fi
echo ""

echo "========================================="
echo "测试完成"
echo "========================================="
