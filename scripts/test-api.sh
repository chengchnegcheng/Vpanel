#!/bin/bash
# API 端点测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
API_BASE="${1:-http://localhost:8080}"
ADMIN_TOKEN="${2:-}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}     V Panel API 测试${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}API 地址:${NC} $API_BASE"
echo ""

# 测试函数
test_endpoint() {
    local method=$1
    local endpoint=$2
    local description=$3
    local auth=$4
    local expected_status=$5
    
    local url="${API_BASE}${endpoint}"
    local headers=""
    
    if [ "$auth" = "admin" ] && [ -n "$ADMIN_TOKEN" ]; then
        headers="-H \"Authorization: Bearer $ADMIN_TOKEN\""
    fi
    
    echo -n "测试: $description ... "
    
    local response
    local status
    
    if [ "$method" = "GET" ]; then
        response=$(eval curl -s -w "\n%{http_code}" $headers "$url" 2>&1)
    else
        response=$(eval curl -s -w "\n%{http_code}" -X "$method" $headers "$url" 2>&1)
    fi
    
    status=$(echo "$response" | tail -n1)
    
    if [ "$status" = "$expected_status" ]; then
        echo -e "${GREEN}✓${NC} (HTTP $status)"
    elif [ "$status" = "401" ] && [ "$auth" = "admin" ] && [ -z "$ADMIN_TOKEN" ]; then
        echo -e "${YELLOW}⊙${NC} (HTTP $status - 需要认证)"
    else
        echo -e "${RED}✗${NC} (HTTP $status, 期望 $expected_status)"
        if [ "$status" = "500" ] || [ "$status" = "503" ]; then
            echo "  响应: $(echo "$response" | head -n-1)"
        fi
    fi
}

# 公开端点测试
echo -e "${YELLOW}=== 公开端点 ===${NC}"
test_endpoint "GET" "/health" "健康检查" "none" "200"
test_endpoint "GET" "/ready" "就绪检查" "none" "200"
echo ""

# 认证端点测试
echo -e "${YELLOW}=== 认证端点 ===${NC}"
test_endpoint "POST" "/api/auth/login" "管理员登录" "none" "400"
test_endpoint "POST" "/api/portal/auth/login" "用户登录" "none" "400"
echo ""

if [ -z "$ADMIN_TOKEN" ]; then
    echo -e "${YELLOW}提示: 未提供 admin token，跳过需要认证的端点测试${NC}"
    echo -e "${YELLOW}使用方法: $0 <API_BASE> <ADMIN_TOKEN>${NC}"
    echo ""
    exit 0
fi

# 管理后台端点测试
echo -e "${YELLOW}=== 管理后台 - 用户管理 ===${NC}"
test_endpoint "GET" "/api/users" "获取用户列表" "admin" "200"
test_endpoint "GET" "/api/roles" "获取角色列表" "admin" "200"
echo ""

echo -e "${YELLOW}=== 管理后台 - 代理管理 ===${NC}"
test_endpoint "GET" "/api/proxies" "获取代理列表" "admin" "200"
echo ""

echo -e "${YELLOW}=== 管理后台 - 系统管理 ===${NC}"
test_endpoint "GET" "/api/system/info" "系统信息" "admin" "200"
test_endpoint "GET" "/api/system/status" "系统状态" "admin" "200"
test_endpoint "GET" "/api/settings" "系统设置" "admin" "200"
echo ""

echo -e "${YELLOW}=== 管理后台 - IP 限制 ===${NC}"
test_endpoint "GET" "/api/admin/ip-restrictions/stats" "IP 限制统计" "admin" "200"
test_endpoint "GET" "/api/admin/ip-whitelist" "IP 白名单列表" "admin" "200"
test_endpoint "GET" "/api/admin/ip-blacklist" "IP 黑名单列表" "admin" "200"
test_endpoint "GET" "/api/admin/settings/ip-restriction" "IP 限制设置" "admin" "200"
echo ""

echo -e "${YELLOW}=== 管理后台 - 订阅管理 ===${NC}"
test_endpoint "GET" "/api/admin/subscriptions" "订阅列表" "admin" "200"
echo ""

echo -e "${YELLOW}=== 管理后台 - 节点管理 ===${NC}"
test_endpoint "GET" "/api/admin/nodes" "节点列表" "admin" "200"
test_endpoint "GET" "/api/admin/node-groups" "节点分组列表" "admin" "200"
echo ""

echo -e "${YELLOW}=== 管理后台 - 商业化 ===${NC}"
test_endpoint "GET" "/api/admin/plans" "套餐列表" "admin" "200"
test_endpoint "GET" "/api/admin/orders" "订单列表" "admin" "200"
test_endpoint "GET" "/api/admin/coupons" "优惠券列表" "admin" "200"
test_endpoint "GET" "/api/admin/gift-cards" "礼品卡列表" "admin" "200"
echo ""

echo -e "${YELLOW}=== 管理后台 - 日志 ===${NC}"
test_endpoint "GET" "/api/logs" "日志列表" "admin" "200"
test_endpoint "GET" "/api/audit-logs" "审计日志" "admin" "200"
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}     测试完成${NC}"
echo -e "${BLUE}========================================${NC}"
