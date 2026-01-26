#!/bin/bash
# 节点诊断脚本

NODE_IP="${1:-180.173.123.192}"
NODE_USER="${2:-root}"

echo "=========================================="
echo "  节点诊断工具"
echo "=========================================="
echo ""
echo "目标节点: $NODE_USER@$NODE_IP"
echo ""

# 检查 SSH 连接
echo "1. 检查 SSH 连接..."
if ! ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no "$NODE_USER@$NODE_IP" "echo 'SSH 连接成功'" 2>/dev/null; then
    echo "✗ SSH 连接失败"
    exit 1
fi
echo "✓ SSH 连接正常"
echo ""

# 检查 Agent 服务状态
echo "2. 检查 Agent 服务状态..."
ssh -o StrictHostKeyChecking=no "$NODE_USER@$NODE_IP" "systemctl status vpanel-agent --no-pager -l" 2>/dev/null || echo "✗ 无法获取服务状态"
echo ""

# 检查 Agent 配置
echo "3. 检查 Agent 配置..."
ssh -o StrictHostKeyChecking=no "$NODE_USER@$NODE_IP" "cat /etc/vpanel/agent.yaml" 2>/dev/null || echo "✗ 配置文件不存在"
echo ""

# 检查 Agent 日志
echo "4. 检查 Agent 日志（最近 20 行）..."
ssh -o StrictHostKeyChecking=no "$NODE_USER@$NODE_IP" "journalctl -u vpanel-agent -n 20 --no-pager" 2>/dev/null || echo "✗ 无法获取日志"
echo ""

# 检查网络连接
echo "5. 检查到 Panel 的网络连接..."
PANEL_URL=$(ssh -o StrictHostKeyChecking=no "$NODE_USER@$NODE_IP" "grep 'url:' /etc/vpanel/agent.yaml | awk '{print \$2}' | tr -d '\"'" 2>/dev/null)
echo "Panel URL: $PANEL_URL"

if [ -n "$PANEL_URL" ]; then
    ssh -o StrictHostKeyChecking=no "$NODE_USER@$NODE_IP" "curl -s -o /dev/null -w 'HTTP Status: %{http_code}\n' '$PANEL_URL/api/health' --connect-timeout 5" 2>/dev/null || echo "✗ 无法连接到 Panel"
fi
echo ""

# 检查端口监听
echo "6. 检查端口监听..."
ssh -o StrictHostKeyChecking=no "$NODE_USER@$NODE_IP" "ss -tlnp | grep -E ':(8081|8443)'" 2>/dev/null || echo "✗ Agent 端口未监听"
echo ""

echo "=========================================="
echo "  诊断完成"
echo "=========================================="
