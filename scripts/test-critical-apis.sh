#!/bin/bash

# 测试关键API端点
# 这个脚本会测试用户报告的404错误的API

BASE_URL="http://localhost:8080"
TOKEN=""

echo "=== 测试关键API端点 ==="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
  local method=$1
  local endpoint=$2
  local description=$3
  local auth_required=$4
  
  echo -n "测试 $description ($method $endpoint)... "
  
  if [ "$auth_required" = "true" ]; then
    if [ -z "$TOKEN" ]; then
      echo -e "${YELLOW}跳过 (需要认证)${NC}"
      return
    fi
    response=$(curl -s -o /dev/null -w "%{http_code}" -X $method "$BASE_URL$endpoint" -H "Authorization: Bearer $TOKEN")
  else
    response=$(curl -s -o /dev/null -w "%{http_code}" -X $method "$BASE_URL$endpoint")
  fi
  
  if [ "$response" = "404" ]; then
    echo -e "${RED}失败 (404 Not Found)${NC}"
    return 1
  elif [ "$response" = "401" ] || [ "$response" = "403" ]; then
    echo -e "${YELLOW}需要认证 ($response)${NC}"
    return 0
  elif [ "$response" -ge 200 ] && [ "$response" -lt 300 ]; then
    echo -e "${GREEN}成功 ($response)${NC}"
    return 0
  else
    echo -e "${YELLOW}其他状态 ($response)${NC}"
    return 0
  fi
}

echo "提示: 如果需要测试需要认证的API，请先登录并设置TOKEN环境变量"
echo "例如: export TOKEN='your-jwt-token'"
echo ""

# 测试用户报告的问题API
echo "--- 用户报告的问题API ---"
test_api "GET" "/api/admin/gift-cards/stats" "礼品卡统计" true
test_api "GET" "/api/admin/reports/pause-stats" "暂停统计" true
test_api "GET" "/api/portal/stats/usage" "用户门户使用统计" true
test_api "GET" "/api/portal/stats/traffic" "用户门户流量统计" true

echo ""
echo "--- 其他关键API ---"
test_api "GET" "/api/stats/dashboard" "仪表板统计" true
test_api "GET" "/api/stats/traffic" "流量统计" true
test_api "GET" "/api/stats/user" "用户统计" true
test_api "GET" "/api/system/info" "系统信息" true
test_api "GET" "/api/system/stats" "系统统计" true
test_api "GET" "/health" "健康检查" false
test_api "GET" "/ready" "就绪检查" false

echo ""
echo "=== 测试完成 ==="
echo ""
echo "如果看到404错误，说明该API端点在后端不存在"
echo "如果看到401/403错误，说明API存在但需要认证"
echo "如果看到200-299状态码，说明API正常工作"
