#!/bin/bash
# 修复 Agent 连接问题

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${GREEN}=== V Panel Agent 连接问题诊断 ===${NC}"
echo ""

# 1. 检查配置文件位置
echo -e "${CYAN}1. 检查配置文件...${NC}"
if [ -f "/etc/vpanel/agent.yaml" ]; then
    CONFIG_FILE="/etc/vpanel/agent.yaml"
    echo -e "${GREEN}✓ 找到配置文件: $CONFIG_FILE${NC}"
elif [ -f "/opt/vpanel-agent/configs/agent.yaml" ]; then
    CONFIG_FILE="/opt/vpanel-agent/configs/agent.yaml"
    echo -e "${GREEN}✓ 找到配置文件: $CONFIG_FILE${NC}"
else
    echo -e "${RED}✗ 未找到配置文件${NC}"
    exit 1
fi

# 2. 读取配置
echo ""
echo -e "${CYAN}2. 读取配置信息...${NC}"
PANEL_URL=$(grep "url:" "$CONFIG_FILE" | awk '{print $2}' | tr -d '"')
NODE_TOKEN=$(grep "token:" "$CONFIG_FILE" | awk '{print $2}' | tr -d '"')

echo -e "Panel URL: ${YELLOW}$PANEL_URL${NC}"
if [ "$NODE_TOKEN" = "your-node-token-here" ] || [ -z "$NODE_TOKEN" ]; then
    echo -e "Node Token: ${RED}未配置${NC}"
else
    echo -e "Node Token: ${GREEN}已配置${NC}"
fi

# 3. 测试网络连接
echo ""
echo -e "${CYAN}3. 测试网络连接...${NC}"

# 提取主机和端口
if [[ $PANEL_URL =~ ^https?://([^:/]+):?([0-9]+)?(/.*)?$ ]]; then
    PANEL_HOST="${BASH_REMATCH[1]}"
    PANEL_PORT="${BASH_REMATCH[2]}"
    
    # 如果没有指定端口，根据协议设置默认端口
    if [ -z "$PANEL_PORT" ]; then
        if [[ $PANEL_URL =~ ^https:// ]]; then
            PANEL_PORT="443"
        else
            PANEL_PORT="80"
        fi
    fi
    
    echo -e "目标主机: ${YELLOW}$PANEL_HOST${NC}"
    echo -e "目标端口: ${YELLOW}$PANEL_PORT${NC}"
    
    # 测试 DNS 解析
    if host "$PANEL_HOST" &>/dev/null; then
        echo -e "${GREEN}✓ DNS 解析成功${NC}"
    else
        echo -e "${RED}✗ DNS 解析失败${NC}"
    fi
    
    # 测试端口连接
    if timeout 5 bash -c "cat < /dev/null > /dev/tcp/$PANEL_HOST/$PANEL_PORT" 2>/dev/null; then
        echo -e "${GREEN}✓ 端口连接成功${NC}"
    else
        echo -e "${RED}✗ 端口连接失败${NC}"
        echo -e "${YELLOW}提示: 请检查防火墙和网络配置${NC}"
    fi
    
    # 测试 HTTP 连接
    if command -v curl &>/dev/null; then
        echo ""
        echo -e "${CYAN}测试 HTTP 连接...${NC}"
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "$PANEL_URL/health" 2>/dev/null || echo "000")
        if [ "$HTTP_CODE" = "200" ]; then
            echo -e "${GREEN}✓ Panel 服务器响应正常 (HTTP $HTTP_CODE)${NC}"
        elif [ "$HTTP_CODE" = "000" ]; then
            echo -e "${RED}✗ 无法连接到 Panel 服务器${NC}"
        else
            echo -e "${YELLOW}⚠ Panel 服务器响应异常 (HTTP $HTTP_CODE)${NC}"
        fi
    fi
fi

# 4. 检查 Agent 服务状态
echo ""
echo -e "${CYAN}4. 检查 Agent 服务状态...${NC}"
if systemctl is-active --quiet vpanel-agent; then
    echo -e "${GREEN}✓ Agent 服务运行中${NC}"
    
    # 显示最近的错误日志
    echo ""
    echo -e "${CYAN}最近的错误日志:${NC}"
    journalctl -u vpanel-agent --since "5 minutes ago" --no-pager | grep -i "error" | tail -5
else
    echo -e "${RED}✗ Agent 服务未运行${NC}"
fi

# 5. 提供修复建议
echo ""
echo -e "${GREEN}=== 修复建议 ===${NC}"
echo ""

if [ "$NODE_TOKEN" = "your-node-token-here" ] || [ -z "$NODE_TOKEN" ]; then
    echo -e "${YELLOW}1. 配置 Node Token:${NC}"
    echo "   - 在 Panel 管理界面创建节点"
    echo "   - 复制节点的 Token"
    echo "   - 编辑配置文件: vim $CONFIG_FILE"
    echo "   - 修改 token 字段"
    echo ""
fi

if [[ $PANEL_URL == *"localhost"* ]] || [[ $PANEL_URL == *"127.0.0.1"* ]]; then
    echo -e "${YELLOW}2. 修改 Panel URL:${NC}"
    echo "   - localhost 只能在本机访问"
    echo "   - 需要改为服务器的实际 IP 或域名"
    echo "   - 编辑配置文件: vim $CONFIG_FILE"
    echo "   - 修改 panel.url 字段"
    echo ""
fi

echo -e "${YELLOW}3. 重启 Agent 服务:${NC}"
echo "   systemctl restart vpanel-agent"
echo ""

echo -e "${YELLOW}4. 查看实时日志:${NC}"
echo "   journalctl -u vpanel-agent -f"
echo ""
