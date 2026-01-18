#!/bin/bash

# 测试IP限制和财务报表修复
# 错误ID: ERR-MKIMADZT-W501D2

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}测试IP限制和财务报表修复${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查服务是否运行
if ! pgrep -f "vpanel" > /dev/null; then
    echo -e "${YELLOW}服务未运行，正在启动...${NC}"
    ./vpanel &
    VPANEL_PID=$!
    sleep 3
    STARTED_BY_SCRIPT=true
else
    echo -e "${GREEN}✓ 服务正在运行${NC}"
    STARTED_BY_SCRIPT=false
fi

BASE_URL="http://localhost:8080"

# 测试1: 健康检查
echo -e "${YELLOW}测试1: 健康检查${NC}"
HEALTH_RESPONSE=$(curl -s $BASE_URL/health)
if echo "$HEALTH_RESPONSE" | grep -q "ok"; then
    echo -e "${GREEN}✓ 健康检查通过${NC}"
else
    echo -e "${RED}✗ 健康检查失败${NC}"
    echo "$HEALTH_RESPONSE"
fi

# 测试2: 登录获取token（需要管理员账户）
echo -e "${YELLOW}测试2: 登录测试${NC}"
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}' 2>/dev/null || echo '{"error":"login failed"}')

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}✓ 登录成功${NC}"
    
    # 测试3: IP限制统计
    echo -e "${YELLOW}测试3: IP限制统计API${NC}"
    IP_STATS=$(curl -s -X GET $BASE_URL/api/admin/ip-restrictions/stats \
        -H "Authorization: Bearer $TOKEN" 2>/dev/null || echo '{"error":"failed"}')
    
    if echo "$IP_STATS" | grep -q "total_active_ips"; then
        echo -e "${GREEN}✓ IP限制API正常${NC}"
    elif echo "$IP_STATS" | grep -q "IP restriction service is not available"; then
        echo -e "${RED}✗ IP限制服务不可用${NC}"
        echo "$IP_STATS"
    else
        echo -e "${YELLOW}⚠ IP限制API响应异常${NC}"
        echo "$IP_STATS"
    fi
    
    # 测试4: 财务报表
    echo -e "${YELLOW}测试4: 财务报表API${NC}"
    REVENUE_REPORT=$(curl -s -X GET "$BASE_URL/api/admin/reports/revenue?start=2024-01-01&end=2024-12-31" \
        -H "Authorization: Bearer $TOKEN" 2>/dev/null || echo '{"error":"failed"}')
    
    if echo "$REVENUE_REPORT" | grep -q "revenue"; then
        echo -e "${GREEN}✓ 财务报表API正常${NC}"
    elif echo "$REVENUE_REPORT" | grep -q "Failed to retrieve revenue data"; then
        echo -e "${RED}✗ 财务报表查询失败${NC}"
        echo "$REVENUE_REPORT"
    else
        echo -e "${YELLOW}⚠ 财务报表API响应异常${NC}"
        echo "$REVENUE_REPORT"
    fi
    
    # 测试5: 订单统计
    echo -e "${YELLOW}测试5: 订单统计API${NC}"
    ORDER_STATS=$(curl -s -X GET $BASE_URL/api/admin/reports/orders \
        -H "Authorization: Bearer $TOKEN" 2>/dev/null || echo '{"error":"failed"}')
    
    if echo "$ORDER_STATS" | grep -q "total"; then
        echo -e "${GREEN}✓ 订单统计API正常${NC}"
    else
        echo -e "${YELLOW}⚠ 订单统计API响应异常${NC}"
        echo "$ORDER_STATS"
    fi
else
    echo -e "${YELLOW}⚠ 登录失败（可能需要先创建管理员账户）${NC}"
    echo "跳过需要认证的测试"
fi

# 如果是脚本启动的服务，停止它
if [ "$STARTED_BY_SCRIPT" = true ]; then
    echo ""
    echo -e "${YELLOW}停止测试服务...${NC}"
    kill $VPANEL_PID 2>/dev/null || true
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}测试完成${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}测试结果总结：${NC}"
echo "1. 健康检查 - 查看上方结果"
echo "2. IP限制API - 查看上方结果"
echo "3. 财务报表API - 查看上方结果"
echo ""
echo -e "${YELLOW}如果测试失败，请检查：${NC}"
echo "1. 数据库连接是否正常"
echo "2. 数据库表是否已创建"
echo "3. 应用日志: tail -f logs/app.log"
echo ""
