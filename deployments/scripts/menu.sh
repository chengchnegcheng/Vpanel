#!/bin/bash
# V Panel 菜单管理脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 获取项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DOCKER_DIR="$PROJECT_ROOT/deployments/docker"

# 清屏函数
clear_screen() {
    clear
}

# 显示标题
show_header() {
    clear_screen
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}       V Panel 管理菜单${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
}

# 暂停函数
pause() {
    echo ""
    read -p "按回车键继续..."
}

# Docker 相关操作
docker_menu() {
    while true; do
        show_header
        echo -e "${GREEN}Docker 部署管理${NC}"
        echo ""
        echo "  1) 启动服务"
        echo "  2) 停止服务"
        echo "  3) 重启服务"
        echo "  4) 查看日志"
        echo "  5) 查看状态"
        echo "  6) 清理数据 (危险操作)"
        echo "  0) 返回主菜单"
        echo ""
        read -p "请选择操作 [0-6]: " choice

        case $choice in
            1)
                echo -e "${GREEN}启动 V Panel...${NC}"
                cd "$DOCKER_DIR"
                if docker compose version &> /dev/null; then
                    docker compose up -d --build
                else
                    docker-compose up -d --build
                fi
                echo -e "${GREEN}V Panel 启动成功！${NC}"
                echo -e "访问地址: ${YELLOW}http://localhost:8080${NC}"
                pause
                ;;
            2)
                echo -e "${YELLOW}停止 V Panel...${NC}"
                cd "$DOCKER_DIR"
                if docker compose version &> /dev/null; then
                    docker compose down
                else
                    docker-compose down
                fi
                echo -e "${GREEN}V Panel 已停止${NC}"
                pause
                ;;
            3)
                echo -e "${YELLOW}重启 V Panel...${NC}"
                cd "$DOCKER_DIR"
                if docker compose version &> /dev/null; then
                    docker compose down
                    docker compose up -d --build
                else
                    docker-compose down
                    docker-compose up -d --build
                fi
                echo -e "${GREEN}V Panel 已重启${NC}"
                pause
                ;;
            4)
                cd "$DOCKER_DIR"
                if docker compose version &> /dev/null; then
                    docker compose logs -f
                else
                    docker-compose logs -f
                fi
                ;;
            5)
                cd "$DOCKER_DIR"
                if docker compose version &> /dev/null; then
                    docker compose ps
                else
                    docker-compose ps
                fi
                pause
                ;;
            6)
                echo -e "${RED}警告: 这将删除所有数据！${NC}"
                read -p "确认删除? (y/N): " confirm
                if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
                    cd "$DOCKER_DIR"
                    if docker compose version &> /dev/null; then
                        docker compose down -v
                    else
                        docker-compose down -v
                    fi
                    echo -e "${GREEN}已清理所有容器和数据卷${NC}"
                fi
                pause
                ;;
            0)
                break
                ;;
            *)
                echo -e "${RED}无效选择，请重试${NC}"
                pause
                ;;
        esac
    done
}

# 开发环境操作
dev_menu() {
    while true; do
        show_header
        echo -e "${GREEN}本地开发环境${NC}"
        echo ""
        echo "  1) 编译并启动服务"
        echo "  2) 仅编译"
        echo "  3) 直接运行 (go run)"
        echo "  4) 运行测试"
        echo "  5) 启动前端开发服务器"
        echo "  6) 安装所有依赖"
        echo "  0) 返回主菜单"
        echo ""
        read -p "请选择操作 [0-6]: " choice

        cd "$PROJECT_ROOT"

        case $choice in
            1)
                echo -e "${GREEN}编译并启动 V Panel...${NC}"
                go build -o v ./cmd/v/main.go
                echo -e "${GREEN}启动服务...${NC}"
                ./v
                ;;
            2)
                echo -e "${GREEN}编译 V Panel...${NC}"
                go build -o v ./cmd/v/main.go
                echo -e "${GREEN}编译完成: ./v${NC}"
                pause
                ;;
            3)
                echo -e "${GREEN}直接运行 (不编译)...${NC}"
                go run ./cmd/v/main.go
                ;;
            4)
                echo -e "${GREEN}运行测试...${NC}"
                go test ./... -v
                pause
                ;;
            5)
                echo -e "${GREEN}启动前端开发服务器...${NC}"
                cd web && npm run dev
                ;;
            6)
                echo -e "${GREEN}安装依赖...${NC}"
                go mod download
                cd web && npm install
                echo -e "${GREEN}依赖安装完成${NC}"
                pause
                ;;
            0)
                break
                ;;
            *)
                echo -e "${RED}无效选择，请重试${NC}"
                pause
                ;;
        esac
    done
}

# 配置管理
config_menu() {
    while true; do
        show_header
        echo -e "${GREEN}配置管理${NC}"
        echo ""
        echo "  1) 创建/重置配置文件"
        echo "  2) 编辑配置文件"
        echo "  3) 查看当前配置"
        echo "  4) 创建 Docker .env 文件"
        echo "  0) 返回主菜单"
        echo ""
        read -p "请选择操作 [0-4]: " choice

        case $choice in
            1)
                if [ -f "$PROJECT_ROOT/configs/config.yaml.example" ]; then
                    cp "$PROJECT_ROOT/configs/config.yaml.example" "$PROJECT_ROOT/configs/config.yaml"
                    echo -e "${GREEN}配置文件已创建/重置${NC}"
                else
                    echo -e "${RED}错误: 找不到示例配置文件${NC}"
                fi
                pause
                ;;
            2)
                if [ -f "$PROJECT_ROOT/configs/config.yaml" ]; then
                    ${EDITOR:-vi} "$PROJECT_ROOT/configs/config.yaml"
                else
                    echo -e "${RED}错误: 配置文件不存在，请先创建${NC}"
                    pause
                fi
                ;;
            3)
                if [ -f "$PROJECT_ROOT/configs/config.yaml" ]; then
                    cat "$PROJECT_ROOT/configs/config.yaml"
                else
                    echo -e "${RED}错误: 配置文件不存在${NC}"
                fi
                pause
                ;;
            4)
                if [ -f "$DOCKER_DIR/.env.example" ]; then
                    cp "$DOCKER_DIR/.env.example" "$DOCKER_DIR/.env"
                    echo -e "${GREEN}Docker .env 文件已创建${NC}"
                else
                    echo -e "${RED}错误: 找不到 .env.example 文件${NC}"
                fi
                pause
                ;;
            0)
                break
                ;;
            *)
                echo -e "${RED}无效选择，请重试${NC}"
                pause
                ;;
        esac
    done
}

# 系统检查
system_check() {
    show_header
    echo -e "${GREEN}系统环境检查${NC}"
    echo ""
    
    # 检查 Go
    if command -v go &> /dev/null; then
        echo -e "${GREEN}✓${NC} Go: $(go version)"
    else
        echo -e "${RED}✗${NC} Go: 未安装"
    fi
    
    # 检查 Docker
    if command -v docker &> /dev/null; then
        echo -e "${GREEN}✓${NC} Docker: $(docker --version)"
    else
        echo -e "${RED}✗${NC} Docker: 未安装"
    fi
    
    # 检查 Docker Compose
    if docker compose version &> /dev/null; then
        echo -e "${GREEN}✓${NC} Docker Compose: $(docker compose version)"
    elif command -v docker-compose &> /dev/null; then
        echo -e "${GREEN}✓${NC} Docker Compose: $(docker-compose --version)"
    else
        echo -e "${RED}✗${NC} Docker Compose: 未安装"
    fi
    
    # 检查 Node.js
    if command -v node &> /dev/null; then
        echo -e "${GREEN}✓${NC} Node.js: $(node --version)"
    else
        echo -e "${RED}✗${NC} Node.js: 未安装"
    fi
    
    # 检查 npm
    if command -v npm &> /dev/null; then
        echo -e "${GREEN}✓${NC} npm: $(npm --version)"
    else
        echo -e "${RED}✗${NC} npm: 未安装"
    fi
    
    echo ""
    echo -e "${BLUE}项目路径:${NC} $PROJECT_ROOT"
    
    # 检查配置文件
    if [ -f "$PROJECT_ROOT/configs/config.yaml" ]; then
        echo -e "${GREEN}✓${NC} 配置文件存在"
    else
        echo -e "${YELLOW}!${NC} 配置文件不存在"
    fi
    
    # 检查数据库
    if [ -f "$PROJECT_ROOT/data/v.db" ]; then
        echo -e "${GREEN}✓${NC} 数据库文件存在"
    else
        echo -e "${YELLOW}!${NC} 数据库文件不存在"
    fi
    
    pause
}

# 主菜单
main_menu() {
    while true; do
        show_header
        echo -e "${BLUE}请选择操作:${NC}"
        echo ""
        echo "  1) Docker 部署管理"
        echo "  2) 本地开发环境"
        echo "  3) 配置管理"
        echo "  4) 系统环境检查"
        echo "  0) 退出"
        echo ""
        read -p "请选择 [0-4]: " choice

        case $choice in
            1)
                docker_menu
                ;;
            2)
                dev_menu
                ;;
            3)
                config_menu
                ;;
            4)
                system_check
                ;;
            0)
                echo -e "${GREEN}再见！${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}无效选择，请重试${NC}"
                pause
                ;;
        esac
    done
}

# 启动主菜单
main_menu
