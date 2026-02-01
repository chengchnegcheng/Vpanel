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

# 安全读取 .env 文件中的变量
read_env_var() {
    local var_name=$1
    local env_file=$2
    # 只匹配 KEY=VALUE 格式，忽略注释和空行
    grep "^${var_name}=" "$env_file" 2>/dev/null | head -n1 | cut -d'=' -f2- | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
}

# 验证密码强度
validate_password() {
    local password=$1
    local min_length=12
    
    # 检查长度
    if [ ${#password} -lt $min_length ]; then
        return 1
    fi
    
    # 检查是否包含大写字母
    if ! echo "$password" | grep -q '[A-Z]'; then
        return 1
    fi
    
    # 检查是否包含小写字母
    if ! echo "$password" | grep -q '[a-z]'; then
        return 1
    fi
    
    # 检查是否包含数字
    if ! echo "$password" | grep -q '[0-9]'; then
        return 1
    fi
    
    # 检查是否包含特殊字符
    if ! echo "$password" | grep -q '[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]'; then
        return 1
    fi
    
    return 0
}

# 验证 JWT Secret
validate_jwt_secret() {
    local secret=$1
    local min_length=32
    
    if [ ${#secret} -lt $min_length ]; then
        return 1
    fi
    
    # 检查是否是默认值
    if [ "$secret" = "CHANGE_ME_OR_SYSTEM_WILL_REFUSE_TO_START" ] || \
       [ "$secret" = "your-secure-jwt-secret-change-me" ] || \
       [ "$secret" = "change-me-in-production" ]; then
        return 1
    fi
    
    return 0
}

# 生产环境安全检查
production_security_check() {
    local env_file="$1"
    local mode=$(read_env_var "V_SERVER_MODE" "$env_file")
    
    # 只在 release 模式下进行严格检查
    if [ "$mode" != "release" ]; then
        echo -e "${YELLOW}警告: 非生产模式 (${mode})，跳过安全检查${NC}"
        return 0
    fi
    
    echo -e "${CYAN}执行生产环境安全检查...${NC}"
    
    local has_error=0
    
    # 检查 JWT Secret
    local jwt_secret=$(read_env_var "V_JWT_SECRET" "$env_file")
    if ! validate_jwt_secret "$jwt_secret"; then
        echo -e "${RED}✗ JWT Secret 不安全！${NC}"
        echo -e "  必须至少 32 字符且不能使用默认值"
        echo -e "  生成方法: ${YELLOW}openssl rand -base64 32${NC}"
        has_error=1
    else
        echo -e "${GREEN}✓ JWT Secret 验证通过${NC}"
    fi
    
    # 检查管理员密码
    local admin_pass=$(read_env_var "V_ADMIN_PASS" "$env_file")
    if [ "$admin_pass" = "CHANGE_ME_OR_SYSTEM_WILL_REFUSE_TO_START" ] || \
       [ "$admin_pass" = "admin123" ] || \
       [ "$admin_pass" = "your-secure-admin-password" ]; then
        echo -e "${RED}✗ 管理员密码使用默认值！${NC}"
        echo -e "  必须修改为强密码"
        has_error=1
    elif ! validate_password "$admin_pass"; then
        echo -e "${RED}✗ 管理员密码强度不足！${NC}"
        echo -e "  要求: 至少12字符，包含大小写字母、数字和特殊字符"
        has_error=1
    else
        echo -e "${GREEN}✓ 管理员密码验证通过${NC}"
    fi
    
    # 检查端口配置
    local server_port=$(read_env_var "V_SERVER_PORT" "$env_file")
    if [ -z "$server_port" ]; then
        echo -e "${RED}✗ 服务端口未配置！${NC}"
        echo -e "  生产环境必须设置固定端口"
        has_error=1
    else
        echo -e "${GREEN}✓ 服务端口: ${server_port}${NC}"
    fi
    
    if [ $has_error -eq 1 ]; then
        echo ""
        echo -e "${RED}========================================${NC}"
        echo -e "${RED}  生产环境安全检查失败！${NC}"
        echo -e "${RED}  请修改 .env 文件后重试${NC}"
        echo -e "${RED}========================================${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✓ 生产环境安全检查通过${NC}"
    return 0
}

# 检查 .env 文件
if [ ! -f "$DOCKER_DIR/.env" ]; then
    echo -e "${YELLOW}创建 .env 配置文件...${NC}"
    cp "$DOCKER_DIR/.env.example" "$DOCKER_DIR/.env"
    echo -e "${GREEN}.env 文件已创建${NC}"
    echo -e "${YELLOW}警告: 请先编辑 .env 文件，修改默认密码和 JWT Secret！${NC}"
    exit 1
fi

# 切换到 Docker 目录
cd "$DOCKER_DIR" || {
    echo -e "${RED}错误: 无法进入 Docker 目录${NC}"
    exit 1
}

# 读取配置
V_SERVER_PORT=$(read_env_var "V_SERVER_PORT" ".env")
V_SERVER_MODE=$(read_env_var "V_SERVER_MODE" ".env")

# 如果端口为空，使用默认值
if [ -z "$V_SERVER_PORT" ]; then
    V_SERVER_PORT=8080
    echo -e "${YELLOW}警告: 端口未配置，使用默认端口 8080${NC}"
fi

# 解析命令行参数
case "${1:-start}" in
    start)
        # 执行生产环境安全检查
        if ! production_security_check ".env"; then
            exit 1
        fi
        
        echo ""
        echo -e "${GREEN}启动 V Panel...${NC}"
        if $COMPOSE_CMD up -d --build; then
            echo ""
            echo -e "${GREEN}========================================${NC}"
            echo -e "${GREEN}V Panel 启动成功！${NC}"
            echo -e "${GREEN}========================================${NC}"
            echo -e "访问地址: ${YELLOW}http://localhost:${V_SERVER_PORT}${NC}"
            echo -e "用户名:   ${YELLOW}admin${NC}"
            echo -e "密码:     ${YELLOW}查看 .env 文件中的 V_ADMIN_PASS${NC}"
            echo -e "模式:     ${YELLOW}${V_SERVER_MODE}${NC}"
            echo ""
        else
            echo -e "${RED}启动失败！${NC}"
            exit 1
        fi
        ;;
    stop)
        echo -e "${YELLOW}停止 V Panel...${NC}"
        if $COMPOSE_CMD down; then
            echo -e "${GREEN}V Panel 已停止${NC}"
        else
            echo -e "${RED}停止失败！${NC}"
            exit 1
        fi
        ;;
    restart)
        echo -e "${YELLOW}重启 V Panel...${NC}"
        if $COMPOSE_CMD down && $COMPOSE_CMD up -d --build; then
            echo -e "${GREEN}V Panel 已重启${NC}"
            echo -e "访问地址: ${YELLOW}http://localhost:${V_SERVER_PORT}${NC}"
        else
            echo -e "${RED}重启失败！${NC}"
            exit 1
        fi
        ;;
    logs)
        $COMPOSE_CMD logs -f
        ;;
    status)
        $COMPOSE_CMD ps
        echo ""
        echo -e "访问地址: ${YELLOW}http://localhost:${V_SERVER_PORT}${NC}"
        echo -e "模式:     ${YELLOW}${V_SERVER_MODE}${NC}"
        ;;
    clean)
        echo ""
        echo -e "${RED}========================================${NC}"
        echo -e "${RED}        警告: 危险操作！${NC}"
        echo -e "${RED}========================================${NC}"
        echo -e "${RED}这将删除所有数据，包括:${NC}"
        echo -e "  - 数据库文件"
        echo -e "  - 日志文件"
        echo -e "  - Xray 配置"
        echo ""
        echo -e "${YELLOW}建议: 在删除前先备份数据${NC}"
        echo -e "  备份命令: docker run --rm -v v-panel-data:/data -v \$(pwd):/backup alpine tar czf /backup/v-panel-backup-\$(date +%Y%m%d-%H%M%S).tar.gz /data"
        echo ""
        read -p "确认删除所有数据? 输入 'DELETE' 确认: " confirm
        if [ "$confirm" = "DELETE" ]; then
            if $COMPOSE_CMD down -v; then
                echo -e "${GREEN}已清理所有容器和数据卷${NC}"
            else
                echo -e "${RED}清理失败！${NC}"
                exit 1
            fi
        else
            echo -e "${YELLOW}已取消操作${NC}"
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
        echo "  clean   - 清理所有数据 (危险)"
        exit 1
        ;;
esac
