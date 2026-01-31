#!/bin/bash
# V Panel 一键启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DOCKER_DIR="$PROJECT_ROOT/deployments/docker"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}       V Panel 一键启动脚本${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: Docker 未安装，请先安装 Docker${NC}"
    exit 1
fi

# 检查 Docker Compose 是否可用
if ! docker compose version &> /dev/null && ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}错误: Docker Compose 未安装${NC}"
    exit 1
fi

# 确定使用哪个 compose 命令
if docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

# 生成随机端口 (10000-65000)
generate_random_port() {
    echo $((10000 + RANDOM % 55000))
}

# 检查端口是否被占用
check_port() {
    local port=$1
    if command -v lsof &> /dev/null; then
        lsof -i :$port &> /dev/null && return 1
    elif command -v netstat &> /dev/null; then
        netstat -an | grep ":$port " &> /dev/null && return 1
    fi
    return 0
}

# 获取可用的随机端口
get_available_port() {
    local port
    local max_attempts=10
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        port=$(generate_random_port)
        if check_port $port; then
            echo $port
            return 0
        fi
        attempt=$((attempt + 1))
    done
    
    # 如果随机端口都被占用，使用默认端口
    echo "8080"
}

# 检查 .env 文件
if [ ! -f "$DOCKER_DIR/.env" ]; then
    echo -e "${YELLOW}创建 .env 配置文件...${NC}"
    cp "$DOCKER_DIR/.env.example" "$DOCKER_DIR/.env"
    echo -e "${GREEN}.env 文件已创建${NC}"
fi

# 读取或生成端口
cd "$DOCKER_DIR"
source .env 2>/dev/null || true

if [ -z "$V_SERVER_PORT" ]; then
    V_SERVER_PORT=$(get_available_port)
    echo -e "${YELLOW}生成随机端口: ${V_SERVER_PORT}${NC}"
    
    # 更新 .env 文件
    if grep -q "^V_SERVER_PORT=" .env; then
        sed -i.bak "s/^V_SERVER_PORT=.*/V_SERVER_PORT=$V_SERVER_PORT/" .env
    else
        echo "V_SERVER_PORT=$V_SERVER_PORT" >> .env
    fi
    rm -f .env.bak
fi

# 解析命令行参数
case "${1:-start}" in
    start)
        echo -e "${GREEN}启动 V Panel...${NC}"
        $COMPOSE_CMD up -d --build
        echo ""
        echo -e "${GREEN}========================================${NC}"
        echo -e "${GREEN}V Panel 启动成功！${NC}"
        echo -e "${GREEN}========================================${NC}"
        echo -e "访问地址: ${YELLOW}http://localhost:${V_SERVER_PORT}${NC}"
        echo -e "用户名:   ${YELLOW}admin${NC}"
        echo -e "密码:     ${YELLOW}查看 .env 文件中的 V_ADMIN_PASS${NC}"
        echo ""
        ;;
    stop)
        echo -e "${YELLOW}停止 V Panel...${NC}"
        $COMPOSE_CMD down
        echo -e "${GREEN}V Panel 已停止${NC}"
        ;;
    restart)
        echo -e "${YELLOW}重启 V Panel...${NC}"
        $COMPOSE_CMD down
        $COMPOSE_CMD up -d --build
        echo -e "${GREEN}V Panel 已重启${NC}"
        echo -e "访问地址: ${YELLOW}http://localhost:${V_SERVER_PORT}${NC}"
        ;;
    logs)
        $COMPOSE_CMD logs -f
        ;;
    status)
        $COMPOSE_CMD ps
        echo ""
        echo -e "访问地址: ${YELLOW}http://localhost:${V_SERVER_PORT}${NC}"
        ;;
    clean)
        echo -e "${RED}警告: 这将删除所有数据！${NC}"
        read -p "确认删除? (y/N): " confirm
        if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
            $COMPOSE_CMD down -v
            echo -e "${GREEN}已清理所有容器和数据卷${NC}"
        else
            echo "已取消"
        fi
        ;;
    *)
        echo "用法: $0 {start|stop|restart|logs|status|clean}"
        echo ""
        echo "  start   - 启动服务 (默认)"
        echo "  stop    - 停止服务"
        echo "  restart - 重启服务"
        echo "  logs    - 查看日志"
        echo "  status  - 查看状态"
        echo "  clean   - 清理所有数据"
        exit 1
        ;;
esac
