#!/bin/bash

# API全面测试脚本
# 测试所有前端调用的API端点，确保没有404错误

BASE_URL="http://localhost:8081"
ADMIN_USER="admin"
ADMIN_PASS="admin123"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 获取token
echo "正在登录..."
TOKEN=$(curl -s "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${ADMIN_USER}\",\"password\":\"${ADMIN_PASS}\"}" \
  | jq -r '.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo -e "${RED}登录失败${NC}"
  exit 1
fi

echo -e "${GREEN}登录成功${NC}"
echo ""

# 测试API端点
test_api() {
  local method=$1
  local endpoint=$2
  local description=$3
  local data=$4
  
  echo -n "测试 ${method} ${endpoint} - ${description}... "
  
  if [ "$method" = "GET" ]; then
    response=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $TOKEN" "${BASE_URL}${endpoint}")
  else
    response=$(curl -s -w "\n%{http_code}" -X "$method" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d "$data" "${BASE_URL}${endpoint}")
  fi
  
  http_code=$(echo "$response" | tail -n1)
  body=$(echo "$response" | head -n-1)
  
  if [ "$http_code" = "404" ]; then
    echo -e "${RED}失败 (404 Not Found)${NC}"
    echo "  响应: $body"
    return 1
  elif [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
    echo -e "${GREEN}成功 ($http_code)${NC}"
    return 0
  elif [ "$http_code" -ge 400 ] && [ "$http_code" -lt 500 ]; then
    echo -e "${YELLOW}客户端错误 ($http_code)${NC}"
    return 0
  else
    echo -e "${YELLOW}其他状态 ($http_code)${NC}"
    return 0
  fi
}

echo "=========================================="
echo "开始API测试"
echo "=========================================="
echo ""

# 统计
total=0
failed=0

# 礼品卡相关API
echo "=== 礼品卡 API ==="
test_api GET "/api/gift-cards/stats" "礼品卡统计" && ((total++)) || ((failed++))
test_api GET "/api/admin/gift-cards/stats" "管理员礼品卡统计" && ((total++)) || ((failed++))
test_api GET "/api/gift-cards" "用户礼品卡列表" && ((total++)) || ((failed++))
echo ""

# 报告相关API
echo "=== 报告 API ==="
test_api GET "/api/admin/reports/revenue" "收入报告" && ((total++)) || ((failed++))
test_api GET "/api/admin/reports/orders" "订单统计" && ((total++)) || ((failed++))
test_api GET "/api/admin/reports/failed-payments" "失败支付统计" && ((total++)) || ((failed++))
test_api GET "/api/admin/reports/pause-stats" "暂停统计" && ((total++)) || ((failed++))
echo ""

# 用户门户统计API
echo "=== 用户门户统计 API ==="
test_api GET "/api/portal/stats/traffic" "流量统计" && ((total++)) || ((failed++))
test_api GET "/api/portal/stats/usage" "使用统计" && ((total++)) || ((failed++))
test_api GET "/api/portal/stats/daily" "每日流量" && ((total++)) || ((failed++))
echo ""

# 试用相关API
echo "=== 试用 API ==="
test_api GET "/api/admin/trials/stats" "试用统计" && ((total++)) || ((failed++))
test_api GET "/api/trial" "试用配置" && ((total++)) || ((failed++))
echo ""

# 设备和IP相关API
echo "=== 设备和IP API ==="
test_api GET "/api/user/devices" "用户设备列表" && ((total++)) || ((failed++))
test_api GET "/api/user/ip-history" "IP历史记录" && ((total++)) || ((failed++))
test_api GET "/api/user/ip-stats" "IP统计" && ((total++)) || ((failed++))
echo ""

# 统计相关API
echo "=== 统计 API ==="
test_api GET "/api/stats/dashboard" "仪表板统计" && ((total++)) || ((failed++))
test_api GET "/api/stats/user" "用户统计" && ((total++)) || ((failed++))
test_api GET "/api/stats/traffic" "流量统计" && ((total++)) || ((failed++))
echo ""

# 订阅相关API
echo "=== 订阅 API ==="
test_api GET "/api/subscription/link" "订阅链接" && ((total++)) || ((failed++))
test_api GET "/api/subscription/info" "订阅信息" && ((total++)) || ((failed++))
echo ""

# 系统相关API
echo "=== 系统 API ==="
test_api GET "/api/system/info" "系统信息" && ((total++)) || ((failed++))
test_api GET "/api/system/status" "系统状态" && ((total++)) || ((failed++))
test_api GET "/api/system/stats" "系统统计" && ((total++)) || ((failed++))
echo ""

# Xray相关API
echo "=== Xray API ==="
test_api GET "/api/xray/status" "Xray状态" && ((total++)) || ((failed++))
test_api GET "/api/xray/versions" "Xray版本列表" && ((total++)) || ((failed++))
test_api GET "/api/xray/version" "Xray当前版本" && ((total++)) || ((failed++))
echo ""

# 用户管理API
echo "=== 用户管理 API ==="
test_api GET "/api/users" "用户列表" && ((total++)) || ((failed++))
test_api GET "/api/auth/me" "当前用户信息" && ((total++)) || ((failed++))
echo ""

# 代理相关API
echo "=== 代理 API ==="
test_api GET "/api/proxies" "代理列表" && ((total++)) || ((failed++))
echo ""

# 节点相关API
echo "=== 节点 API ==="
test_api GET "/api/admin/nodes" "节点列表" && ((total++)) || ((failed++))
test_api GET "/api/admin/node-groups" "节点组列表" && ((total++)) || ((failed++))
echo ""

# 计划相关API
echo "=== 计划 API ==="
test_api GET "/api/plans" "计划列表" && ((total++)) || ((failed++))
test_api GET "/api/admin/plans" "管理员计划列表" && ((total++)) || ((failed++))
echo ""

# 订单相关API
echo "=== 订单 API ==="
test_api GET "/api/orders" "用户订单列表" && ((total++)) || ((failed++))
test_api GET "/api/admin/orders" "管理员订单列表" && ((total++)) || ((failed++))
echo ""

# 余额相关API
echo "=== 余额 API ==="
test_api GET "/api/balance" "用户余额" && ((total++)) || ((failed++))
test_api GET "/api/balance/transactions" "余额交易记录" && ((total++)) || ((failed++))
echo ""

# 优惠券相关API
echo "=== 优惠券 API ==="
test_api GET "/api/admin/coupons" "优惠券列表" && ((total++)) || ((failed++))
echo ""

# 邀请相关API
echo "=== 邀请 API ==="
test_api GET "/api/invite/code" "邀请码" && ((total++)) || ((failed++))
test_api GET "/api/invite/stats" "邀请统计" && ((total++)) || ((failed++))
echo ""

# 发票相关API
echo "=== 发票 API ==="
test_api GET "/api/invoices" "发票列表" && ((total++)) || ((failed++))
echo ""

# 设置相关API
echo "=== 设置 API ==="
test_api GET "/api/settings" "系统设置" && ((total++)) || ((failed++))
test_api GET "/api/settings/xray" "Xray设置" && ((total++)) || ((failed++))
test_api GET "/api/settings/protocols" "协议设置" && ((total++)) || ((failed++))
echo ""

# 日志相关API
echo "=== 日志 API ==="
test_api GET "/api/logs" "日志列表" && ((total++)) || ((failed++))
echo ""

# 证书相关API
echo "=== 证书 API ==="
test_api GET "/api/certificates" "证书列表" && ((total++)) || ((failed++))
echo ""

# 角色相关API
echo "=== 角色 API ==="
test_api GET "/api/roles" "角色列表" && ((total++)) || ((failed++))
test_api GET "/api/permissions" "权限列表" && ((total++)) || ((failed++))
echo ""

# IP限制相关API
echo "=== IP限制 API ==="
test_api GET "/api/admin/ip-restrictions/stats" "IP限制统计" && ((total++)) || ((failed++))
test_api GET "/api/admin/ip-whitelist" "IP白名单" && ((total++)) || ((failed++))
test_api GET "/api/admin/ip-blacklist" "IP黑名单" && ((total++)) || ((failed++))
echo ""

# 健康检查API
echo "=== 健康检查 API ==="
test_api GET "/health" "健康检查" && ((total++)) || ((failed++))
test_api GET "/ready" "就绪检查" && ((total++)) || ((failed++))
echo ""

echo "=========================================="
echo "测试完成"
echo "=========================================="
echo "总计: $total 个API"
echo -e "失败: ${RED}$failed${NC} 个API"
if [ $failed -eq 0 ]; then
  echo -e "${GREEN}所有API测试通过！${NC}"
  exit 0
else
  echo -e "${RED}有 $failed 个API测试失败${NC}"
  exit 1
fi
