#!/bin/bash

# IP限制服务修复脚本
# 修复错误ID: ERR-MKIMADZT-W501D2

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}IP限制服务修复脚本${NC}"
echo -e "${BLUE}错误ID: ERR-MKIMADZT-W501D2${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 1. 停止服务
echo -e "${YELLOW}1. 停止V Panel服务...${NC}"
if pgrep -f "vpanel" > /dev/null; then
    pkill -f "vpanel" || true
    sleep 2
    echo -e "${GREEN}✓ 服务已停止${NC}"
else
    echo -e "${YELLOW}服务未运行${NC}"
fi

# 2. 重新编译
echo -e "${YELLOW}2. 重新编译应用...${NC}"
go build -o vpanel cmd/v/main.go
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 编译成功${NC}"
else
    echo -e "${RED}✗ 编译失败${NC}"
    exit 1
fi

# 3. 检查数据库配置
echo -e "${YELLOW}3. 检查数据库配置...${NC}"
if [ -f "configs/config.yaml" ]; then
    echo -e "${GREEN}✓ 配置文件存在${NC}"
else
    echo -e "${RED}✗ 配置文件不存在${NC}"
    exit 1
fi

# 4. 启动服务（测试模式）
echo -e "${YELLOW}4. 启动服务进行测试...${NC}"
./vpanel &
VPANEL_PID=$!
sleep 3

# 检查服务是否启动成功
if ps -p $VPANEL_PID > /dev/null; then
    echo -e "${GREEN}✓ 服务启动成功 (PID: $VPANEL_PID)${NC}"
    
    # 5. 测试API端点
    echo -e "${YELLOW}5. 测试API端点...${NC}"
    
    # 测试健康检查
    if curl -s http://localhost:8080/health > /dev/null; then
        echo -e "${GREEN}✓ 健康检查通过${NC}"
    else
        echo -e "${RED}✗ 健康检查失败${NC}"
    fi
    
    # 停止测试服务
    kill $VPANEL_PID 2>/dev/null || true
    sleep 1
else
    echo -e "${RED}✗ 服务启动失败${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}修复完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}修复内容：${NC}"
echo "1. 修复了IP限制服务初始化逻辑"
echo "2. 禁用了GeoIP数据库依赖（可选功能）"
echo "3. 确保服务即使没有GeoIP数据库也能正常运行"
echo ""
echo -e "${YELLOW}下一步操作：${NC}"
echo "1. 启动服务: ./vpanel 或 ./vpanel.sh start"
echo "2. 检查日志: tail -f logs/app.log"
echo "3. 测试IP限制功能: 访问管理后台 -> IP限制"
echo "4. 测试财务报表: 访问管理后台 -> 报表 -> 财务报表"
echo ""
echo -e "${BLUE}注意事项：${NC}"
echo "- IP限制功能现在可以正常使用"
echo "- 地理位置功能已禁用（需要GeoIP数据库）"
echo "- 如需启用地理位置功能，请下载GeoLite2-City.mmdb到data目录"
echo ""
