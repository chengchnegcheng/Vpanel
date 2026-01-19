#!/bin/bash

# 创建示例节点的快速脚本

API_URL=${API_URL:-"http://localhost:8080"}
ADMIN_TOKEN=${ADMIN_TOKEN:-""}

if [ -z "$ADMIN_TOKEN" ]; then
    echo "错误: 需要设置 ADMIN_TOKEN"
    echo "用法: ADMIN_TOKEN=xxx ./scripts/create-sample-node.sh"
    exit 1
fi

echo "创建示例节点..."

# 创建香港节点
curl -X POST "${API_URL}/api/admin/nodes" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "香港-01",
        "address": "hk1.example.com",
        "port": 8443,
        "region": "香港",
        "weight": 10,
        "max_users": 100
    }'

echo ""
echo "完成！"
