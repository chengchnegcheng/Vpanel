#!/bin/bash

# 最终验证脚本 - 验证所有修复

set -e

BASE_URL="http://localhost:8081"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================"
echo -e "最终验证 - 所有功能测试"
echo -e "========================================${NC}"
echo ""

# 登录
echo -e "${YELLOW}登录...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ 登录失败${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 登录成功${NC}"
echo ""

# 测试所有API
TESTS=(
  "GET|/api/admin/ip-restrictions/stats|IP限制统计"
  "GET|/api/admin/ip-restrictions/online|在线IP列表"
  "GET|/api/admin/ip-whitelist|IP白名单"
  "GET|/api/admin/ip-blacklist|IP黑名单"
  "GET|/api/admin/reports/revenue?start=2024-01-01&end=2024-12-31|财务报表-收入"
  "GET|/api/admin/reports/orders|财务报表-订单"
  "GET|/api/admin/reports/failed-payments|失败支付统计"
  "GET|/api/admin/reports/pause-stats|暂停统计"
)

PASSED=0
FAILED=0

for test in "${TESTS[@]}"; do
  IFS='|' read -r method endpoint name <<< "$test"
  
  echo -e "${YELLOW}测试: $name${NC}"
  
  RESPONSE=$(curl -s -X $method "$BASE_URL$endpoint" \
    -H "Authorization: Bearer $TOKEN" \
    -w "\n%{http_code}")
  
  HTTP_CODE=$(echo "$RESPONSE" | tail -1)
  BODY=$(echo "$RESPONSE" | sed '$d')
  
  if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ $name - 成功 (200)${NC}"
    PASSED=$((PASSED + 1))
  else
    echo -e "${RED}✗ $name - 失败 ($HTTP_CODE)${NC}"
    echo "  响应: $(echo "$BODY" | head -c 100)"
    FAILED=$((FAILED + 1))
  fi
  echo ""
done

echo -e "${BLUE}========================================"
echo -e "测试结果"
echo -e "========================================${NC}"
echo ""
echo -e "${GREEN}通过: $PASSED${NC}"
echo -e "${RED}失败: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
  echo -e "${GREEN}✅ 所有测试通过！${NC}"
  echo ""
  echo -e "${BLUE}修复内容:${NC}"
  echo "1. ✓ 添加了 RetryService 初始化"
  echo "2. ✓ 修复了 PaymentHandler 使用 RetryService"
  echo "3. ✓ 添加了 /api/admin/ip-restrictions/online 端点"
  echo "4. ✓ 增强了所有handler的错误处理"
  echo "5. ✓ 添加了nil检查防止panic"
  echo ""
  echo -e "${YELLOW}用户现在可以正常使用:${NC}"
  echo "- IP限制管理功能"
  echo "- 财务报表功能"
  echo "- 失败支付统计"
  echo "- 订阅暂停统计"
  exit 0
else
  echo -e "${RED}❌ 有测试失败，请检查${NC}"
  exit 1
fi
