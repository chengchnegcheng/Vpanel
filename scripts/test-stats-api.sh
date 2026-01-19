#!/bin/bash

# 测试用户门户统计数据 API
# 使用方法: ./scripts/test-stats-api.sh <token>

set -e

# 检查参数
if [ -z "$1" ]; then
    echo "使用方法: $0 <auth_token>"
    echo "示例: $0 eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    exit 1
fi

TOKEN="$1"
BASE_URL="${BASE_URL:-http://localhost:8080}"
API_URL="$BASE_URL/api"

echo "=========================================="
echo "测试用户门户统计数据 API"
echo "=========================================="
echo ""

# 测试1: 获取流量统计（周）
echo "1. 测试获取流量统计（周）..."
curl -s -X GET \
    -H "Authorization: Bearer $TOKEN" \
    "$API_URL/portal/stats/traffic?period=week" | jq '.'
echo ""

# 测试2: 获取流量统计（月）
echo "2. 测试获取流量统计（月）..."
curl -s -X GET \
    -H "Authorization: Bearer $TOKEN" \
    "$API_URL/portal/stats/traffic?period=month" | jq '.'
echo ""

# 测试3: 获取每日流量
echo "3. 测试获取每日流量..."
curl -s -X GET \
    -H "Authorization: Bearer $TOKEN" \
    "$API_URL/portal/stats/daily?days=7" | jq '.'
echo ""

# 测试4: 获取使用统计
echo "4. 测试获取使用统计..."
curl -s -X GET \
    -H "Authorization: Bearer $TOKEN" \
    "$API_URL/portal/stats/usage" | jq '.'
echo ""

# 测试5: 获取仪表板数据
echo "5. 测试获取仪表板数据..."
curl -s -X GET \
    -H "Authorization: Bearer $TOKEN" \
    "$API_URL/portal/dashboard" | jq '.'
echo ""

# 测试6: 获取流量摘要
echo "6. 测试获取流量摘要..."
curl -s -X GET \
    -H "Authorization: Bearer $TOKEN" \
    "$API_URL/portal/dashboard/traffic" | jq '.'
echo ""

echo "=========================================="
echo "测试完成"
echo "=========================================="
