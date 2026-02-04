#!/bin/bash

# 初始化示例数据脚本
# 用于快速创建测试节点和代理

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

API_URL=${API_URL:-"http://localhost:8080"}
ADMIN_TOKEN=${ADMIN_TOKEN:-""}

show_help() {
    echo "初始化示例数据脚本"
    echo ""
    echo "用法:"
    echo "  $0 nodes              创建示例节点"
    echo "  $0 proxies            创建示例代理"
    echo "  $0 all                创建所有示例数据"
    echo ""
    echo "环境变量:"
    echo "  API_URL               API 地址 (默认: http://localhost:8080)"
    echo "  ADMIN_TOKEN           管理员 Token (必需)"
    echo ""
    echo "示例:"
    echo "  ADMIN_TOKEN=xxx ./scripts/init-sample-data.sh nodes"
    echo ""
}

check_token() {
    if [ -z "$ADMIN_TOKEN" ]; then
        echo -e "${RED}错误: 需要设置 ADMIN_TOKEN${NC}"
        echo "请先登录获取 Token，或设置环境变量:"
        echo "  export ADMIN_TOKEN=your-admin-token"
        exit 1
    fi
}

create_nodes() {
    echo -e "${YELLOW}创建示例节点...${NC}"
    echo ""
    
    # 节点 1: 香港
    echo "创建节点: 香港-01"
    curl -s -X POST "${API_URL}/api/admin/nodes" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "香港-01",
            "address": "hk1.example.com",
            "port": 18443,
            "region": "香港",
            "weight": 10,
            "max_users": 100,
            "tags": ["高速", "稳定"]
        }' | jq '.' || echo "创建失败"
    echo ""
    
    # 节点 2: 日本
    echo "创建节点: 日本-01"
    curl -s -X POST "${API_URL}/api/admin/nodes" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "日本-01",
            "address": "jp1.example.com",
            "port": 18443,
            "region": "日本",
            "weight": 8,
            "max_users": 80,
            "tags": ["游戏", "低延迟"]
        }' | jq '.' || echo "创建失败"
    echo ""
    
    # 节点 3: 美国
    echo "创建节点: 美国-01"
    curl -s -X POST "${API_URL}/api/admin/nodes" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "美国-01",
            "address": "us1.example.com",
            "port": 18443,
            "region": "美国",
            "weight": 5,
            "max_users": 50,
            "tags": ["流媒体"]
        }' | jq '.' || echo "创建失败"
    echo ""
    
    echo -e "${GREEN}✓ 节点创建完成${NC}"
}

create_proxies() {
    echo -e "${YELLOW}创建示例代理...${NC}"
    echo ""
    
    # 获取节点列表
    echo "获取节点列表..."
    nodes=$(curl -s "${API_URL}/api/admin/nodes" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}")
    
    # 提取第一个节点 ID
    node_id=$(echo "$nodes" | jq -r '.[0].id // .data[0].id // .nodes[0].id // 1')
    
    if [ "$node_id" = "null" ] || [ -z "$node_id" ]; then
        echo -e "${RED}错误: 没有找到节点，请先创建节点${NC}"
        return 1
    fi
    
    echo "使用节点 ID: $node_id"
    echo ""
    
    # 代理 1: VLESS
    echo "创建代理: VLESS-443"
    curl -s -X POST "${API_URL}/api/proxies" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"VLESS-443\",
            \"protocol\": \"vless\",
            \"port\": 443,
            \"node_id\": ${node_id},
            \"host\": \"example.com\",
            \"settings\": {
                \"uuid\": \"$(uuidgen | tr '[:upper:]' '[:lower:]')\",
                \"network\": \"ws\",
                \"security\": \"tls\",
                \"ws_settings\": {
                    \"path\": \"/vless\"
                }
            },
            \"enabled\": true,
            \"remark\": \"VLESS WebSocket TLS\"
        }" | jq '.' || echo "创建失败"
    echo ""
    
    # 代理 2: VMess
    echo "创建代理: VMess-8443"
    curl -s -X POST "${API_URL}/api/proxies" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        -H "Conten