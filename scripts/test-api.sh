#!/bin/bash

# API 测试脚本

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 配置
API_URL=${API_URL:-"http://localhost:8080"}
ADMIN_TOKEN=${ADMIN_TOKEN:-""}

if [ -z "$ADMIN_TOKEN" ]; then
    echo -e "${RED}错误: 需要设置 ADMIN_TOKEN 环境变量${NC}"
    echo "export ADMIN_TOKEN=your-token"
    exit 1
fi

echo -e "${GREEN}测试 V Panel API${NC}"
echo "API URL: $API_URL"
echo ""

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    
    echo -e "${YELLOW}测试: $name${NC}"
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -H "Content-Type: application/json" \
            "$API_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ 成功 (HTTP $http_code)${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
        echo "$body"
    fi
    
    echo ""
}

# 1. 测试健康检查
test_api "健康检查" "GET" "/health"

# 2. 测试节点列表
test_api "获取节点列表" "GET" "/api/admin/nodes"

# 3. 创建节点
NODE_DATA='{
  "name": "Test-Node",
  "address": "test.example.com",
  "port": 443,
  "enabled": true
}'
test_api "创建节点" "POST" "/api/admin/nodes" "$NODE_DATA"

# 4. 测试代理列表
test_api "获取代理列表" "GET" "/api/proxies"

# 5. 创建代理（带节点）
PROXY_DATA='{
  "name": "Test-VLESS",
  "protocol": "vless",
  "node_id": 1,
  "port": 10443,
  "settings": {
    "uuid": "test-uuid-12345"
  },
  "enabled": true
}'
test_api "创建代理（带节点）" "POST" "/api/proxies" "$PROXY_DATA"

# 6. 测试配置预览
test_api "预览节点配置" "GET" "/api/admin/nodes/1/config/preview"

# 7. 测试 SSH 连接（会失败，仅测试 API）
SSH_DATA='{
  "host": "test.example.com",
  "port": 22,
  "username": "root",
  "password": "test"
}'
echo -e "${YELLOW}测试: SSH 连接测试（预期失败）${NC}"
curl -s -X POST \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$SSH_DATA" \
    "$API_URL/api/admin/nodes/test-connection" | jq '.'
echo ""

echo -e "${GREEN}API 测试完成！${NC}"
