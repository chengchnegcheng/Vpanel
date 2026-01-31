#!/bin/bash

# Clash Meta 管理脚本
# 支持安装、卸载、启动、停止、重启、状态查看、日志查看、更新订阅等功能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/clash-meta"
SERVICE_NAME="clash-meta"
BINARY_NAME="mihomo"

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "此操作需要 root 权限"
        exit 1
    fi
}

# 检查是否已安装
check_installed() {
    if [ ! -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        print_error "Clash Meta 未安装"
        print_info "请先运行: $0 install"
        exit 1
    fi
}

# 检测系统架构
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            print_error "不支持的系统架构: $arch"
            exit 1
            ;;
    esac
}

# 获取最新版本号
get_latest_version() {
    print_info "获取最新版本信息..."
    
    # 临时取消所有代理设置（包括大小写变体）
    local old_http_proxy=$http_proxy
    local old_https_proxy=$https_proxy
    local old_HTTP_PROXY=$HTTP_PROXY
    local old_HTTPS_PROXY=$HTTPS_PROXY
    local old_all_proxy=$all_proxy
    local old_ALL_PROXY=$ALL_PROXY
    
    unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY
    
    # 方法1: 使用 GitHub API
    LATEST_VERSION=$(curl -s --connect-timeout 10 --noproxy '*' https://api.github.com/repos/MetaCubeX/mihomo/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    # 方法2: 如果 API 失败，尝试从 HTML 页面获取
    if [ -z "$LATEST_VERSION" ]; then
        print_warning "GitHub API 访问失败，尝试备用方法..."
        LATEST_VERSION=$(curl -sL --connect-timeout 10 --noproxy '*' https://github.com/MetaCubeX/mihomo/releases/latest | grep -oP 'tag/\K[^"]+' | head -1)
    fi
    
    # 恢复代理设置
    [ -n "$old_http_proxy" ] && export http_proxy=$old_http_proxy
    [ -n "$old_https_proxy" ] && export https_proxy=$old_https_proxy
    [ -n "$old_HTTP_PROXY" ] && export HTTP_PROXY=$old_HTTP_PROXY
    [ -n "$old_HTTPS_PROXY" ] && export HTTPS_PROXY=$old_HTTPS_PROXY
    [ -n "$old_all_proxy" ] && export all_proxy=$old_all_proxy
    [ -n "$old_ALL_PROXY" ] && export ALL_PROXY=$old_ALL_PROXY
    
    # 方法3: 如果还是失败，使用固定版本
    if [ -z "$LATEST_VERSION" ]; then
        print_warning "无法自动获取版本，使用默认版本"
        LATEST_VERSION="v1.18.10"
        print_info "使用版本: $LATEST_VERSION"
    else
        print_success "最新版本: $LATEST_VERSION"
    fi
}

# 安装 Clash Meta
install_clash_meta() {
    check_root
    
    print_info "开始安装 Clash Meta..."
    
    # 检测系统架构
    detect_arch
    print_info "系统架构: $ARCH"
    
    # 获取最新版本
    get_latest_version
    
    # 下载
    print_info "下载 Clash Meta..."
    local download_url="https://github.com/MetaCubeX/mihomo/releases/download/${LATEST_VERSION}/mihomo-linux-${ARCH}-${LATEST_VERSION}.gz"
    local temp_file="/tmp/mihomo-linux-${ARCH}.gz"
    
    print_info "下载地址: $download_url"
    
    # 临时取消所有代理设置（包括大小写变体）
    local old_http_proxy=$http_proxy
    local old_https_proxy=$https_proxy
    local old_HTTP_PROXY=$HTTP_PROXY
    local old_HTTPS_PROXY=$HTTPS_PROXY
    local old_all_proxy=$all_proxy
    local old_ALL_PROXY=$ALL_PROXY
    
    unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY
    
    if ! curl -L --progress-bar --connect-timeout 30 --max-time 300 --noproxy '*' -o "$temp_file" "$download_url"; then
        print_error "下载失败"
        print_info "请检查网络连接或手动下载："
        print_info "$download_url"
        
        # 恢复代理设置
        [ -n "$old_http_proxy" ] && export http_proxy=$old_http_proxy
        [ -n "$old_https_proxy" ] && export https_proxy=$old_https_proxy
        [ -n "$old_HTTP_PROXY" ] && export HTTP_PROXY=$old_HTTP_PROXY
        [ -n "$old_HTTPS_PROXY" ] && export HTTPS_PROXY=$old_HTTPS_PROXY
        [ -n "$old_all_proxy" ] && export all_proxy=$old_all_proxy
        [ -n "$old_ALL_PROXY" ] && export ALL_PROXY=$old_ALL_PROXY
        exit 1
    fi
    
    # 恢复代理设置
    [ -n "$old_http_proxy" ] && export http_proxy=$old_http_proxy
    [ -n "$old_https_proxy" ] && export https_proxy=$old_https_proxy
    [ -n "$old_HTTP_PROXY" ] && export HTTP_PROXY=$old_HTTP_PROXY
    [ -n "$old_HTTPS_PROXY" ] && export HTTPS_PROXY=$old_HTTPS_PROXY
    [ -n "$old_all_proxy" ] && export all_proxy=$old_all_proxy
    [ -n "$old_ALL_PROXY" ] && export ALL_PROXY=$old_ALL_PROXY
    
    # 解压
    print_info "解压文件..."
    gunzip -f "$temp_file"
    
    # 安装
    print_info "安装到系统..."
    mv "/tmp/mihomo-linux-${ARCH}" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    ln -sf "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/clash-meta"
    
    # 创建配置目录
    mkdir -p "$CONFIG_DIR"
    
    # 创建示例配置
    if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
        create_sample_config
    fi
    
    # 创建 systemd 服务
    create_systemd_service
    
    # 配置系统代理
    configure_system_proxy
    
    print_success "安装完成！"
    show_usage
}

# 创建示例配置文件
create_sample_config() {
    cat > "$CONFIG_DIR/config.yaml" << 'EOF'
# Clash Meta 配置文件

mixed-port: 7890
port: 7891
socks-port: 7892
allow-lan: true
bind-address: '*'
mode: rule
log-level: info
ipv6: true
external-controller: 0.0.0.0:9090
secret: ""
unified-delay: true
tcp-concurrent: true

dns:
  enable: true
  listen: 0.0.0.0:1053
  ipv6: true
  enhanced-mode: fake-ip
  fake-ip-range: 198.18.0.1/16
  nameserver:
    - 223.5.5.5
    - 119.29.29.29
  fallback:
    - 8.8.8.8
    - 1.1.1.1

tun:
  enable: false
  stack: system
  dns-hijack:
    - any:53
  auto-route: true
  auto-detect-interface: true

proxies: []

proxy-groups:
  - name: PROXY
    type: select
    proxies:
      - DIRECT

rules:
  - GEOIP,CN,DIRECT
  - MATCH,PROXY
EOF
    print_success "示例配置已创建: $CONFIG_DIR/config.yaml"
}

# 创建 systemd 服务
create_systemd_service() {
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Clash Meta Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=$INSTALL_DIR/$BINARY_NAME -d $CONFIG_DIR
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    print_success "systemd 服务已创建"
}

# 配置系统代理
configure_system_proxy() {
    cat > /etc/profile.d/clash-meta.sh << 'EOF'
# Clash Meta 代理配置
export http_proxy=http://127.0.0.1:7890
export https_proxy=http://127.0.0.1:7890
export no_proxy=localhost,127.0.0.1,::1
EOF
    print_success "系统代理配置已添加"
}

# 卸载
uninstall_clash_meta() {
    check_root
    
    print_warning "即将卸载 Clash Meta"
    read -p "确认继续？(y/n): " -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "已取消"
        exit 0
    fi
    
    # 停止服务
    if systemctl is-active --quiet $SERVICE_NAME; then
        systemctl stop $SERVICE_NAME
    fi
    
    # 禁用服务
    if systemctl is-enabled --quiet $SERVICE_NAME 2>/dev/null; then
        systemctl disable $SERVICE_NAME
    fi
    
    # 删除服务文件
    rm -f /etc/systemd/system/${SERVICE_NAME}.service
    systemctl daemon-reload
    
    # 删除二进制文件
    rm -f "$INSTALL_DIR/$BINARY_NAME"
    rm -f "$INSTALL_DIR/clash-meta"
    
    # 询问是否删除配置
    read -p "是否删除配置文件？(y/n): " -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$CONFIG_DIR"
        print_success "配置文件已删除"
    fi
    
    # 删除系统代理配置
    rm -f /etc/profile.d/clash-meta.sh
    
    print_success "卸载完成"
}

# 启动服务
start_service() {
    check_root
    check_installed
    
    print_info "启动 Clash Meta..."
    systemctl start $SERVICE_NAME
    sleep 1
    
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "服务已启动"
        systemctl status $SERVICE_NAME --no-pager
    else
        print_error "服务启动失败"
        systemctl status $SERVICE_NAME --no-pager
        exit 1
    fi
}

# 停止服务
stop_service() {
    check_root
    check_installed
    
    print_info "停止 Clash Meta..."
    systemctl stop $SERVICE_NAME
    print_success "服务已停止"
}

# 重启服务
restart_service() {
    check_root
    check_installed
    
    print_info "重启 Clash Meta..."
    systemctl restart $SERVICE_NAME
    sleep 1
    
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "服务已重启"
        systemctl status $SERVICE_NAME --no-pager
    else
        print_error "服务重启失败"
        systemctl status $SERVICE_NAME --no-pager
        exit 1
    fi
}

# 查看状态
show_status() {
    check_installed
    
    echo "=========================================="
    echo "  Clash Meta 状态"
    echo "=========================================="
    echo ""
    
    # 服务状态
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "服务状态: 运行中"
    else
        print_error "服务状态: 已停止"
    fi
    
    # 开机自启
    if systemctl is-enabled --quiet $SERVICE_NAME 2>/dev/null; then
        print_info "开机自启: 已启用"
    else
        print_info "开机自启: 未启用"
    fi
    
    # 版本信息
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        local version=$($INSTALL_DIR/$BINARY_NAME -v 2>&1 | head -1)
        print_info "版本: $version"
    fi
    
    # 配置文件
    if [ -f "$CONFIG_DIR/config.yaml" ]; then
        print_info "配置文件: $CONFIG_DIR/config.yaml"
    else
        print_warning "配置文件不存在"
    fi
    
    echo ""
    echo "详细状态:"
    systemctl status $SERVICE_NAME --no-pager
}

# 查看日志
show_logs() {
    check_installed
    
    print_info "查看 Clash Meta 日志 (Ctrl+C 退出)"
    journalctl -u $SERVICE_NAME -f
}

# 启用开机自启
enable_service() {
    check_root
    check_installed
    
    systemctl enable $SERVICE_NAME
    print_success "已启用开机自启"
}

# 禁用开机自启
disable_service() {
    check_root
    check_installed
    
    systemctl disable $SERVICE_NAME
    print_success "已禁用开机自启"
}

# 更新订阅
update_subscription() {
    check_root
    check_installed
    
    read -p "请输入订阅链接: " SUBSCRIPTION_URL
    
    if [ -z "$SUBSCRIPTION_URL" ]; then
        print_error "订阅链接不能为空"
        exit 1
    fi
    
    print_info "下载订阅配置..."
    
    # 备份当前配置
    if [ -f "$CONFIG_DIR/config.yaml" ]; then
        cp "$CONFIG_DIR/config.yaml" "$CONFIG_DIR/config.yaml.backup"
        print_info "已备份当前配置"
    fi
    
    # 临时取消所有代理设置（包括大小写变体）
    local old_http_proxy=$http_proxy
    local old_https_proxy=$https_proxy
    local old_HTTP_PROXY=$HTTP_PROXY
    local old_HTTPS_PROXY=$HTTPS_PROXY
    local old_all_proxy=$all_proxy
    local old_ALL_PROXY=$ALL_PROXY
    
    unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY
    
    # 下载新配置
    if curl -L --connect-timeout 30 --max-time 60 --noproxy '*' -o "$CONFIG_DIR/config.yaml" "$SUBSCRIPTION_URL"; then
        if [ -s "$CONFIG_DIR/config.yaml" ]; then
            print_success "订阅配置更新成功"
            
            # 重启服务
            if systemctl is-active --quiet $SERVICE_NAME; then
                print_info "重启服务..."
                systemctl restart $SERVICE_NAME
                print_success "服务已重启"
            fi
        else
            print_error "下载的配置文件为空"
            if [ -f "$CONFIG_DIR/config.yaml.backup" ]; then
                mv "$CONFIG_DIR/config.yaml.backup" "$CONFIG_DIR/config.yaml"
                print_info "已恢复备份配置"
            fi
        fi
    else
        print_error "订阅配置下载失败"
        if [ -f "$CONFIG_DIR/config.yaml.backup" ]; then
            mv "$CONFIG_DIR/config.yaml.backup" "$CONFIG_DIR/config.yaml"
            print_info "已恢复备份配置"
        fi
    fi
    
    # 恢复代理设置
    [ -n "$old_http_proxy" ] && export http_proxy=$old_http_proxy
    [ -n "$old_https_proxy" ] && export https_proxy=$old_https_proxy
    [ -n "$old_HTTP_PROXY" ] && export HTTP_PROXY=$old_HTTP_PROXY
    [ -n "$old_HTTPS_PROXY" ] && export HTTPS_PROXY=$old_HTTPS_PROXY
    [ -n "$old_all_proxy" ] && export all_proxy=$old_all_proxy
    [ -n "$old_ALL_PROXY" ] && export ALL_PROXY=$old_ALL_PROXY
}

# 编辑配置
edit_config() {
    check_installed
    
    if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
        print_error "配置文件不存在"
        exit 1
    fi
    
    # 检测可用的编辑器
    if command -v nano &> /dev/null; then
        nano "$CONFIG_DIR/config.yaml"
    elif command -v vim &> /dev/null; then
        vim "$CONFIG_DIR/config.yaml"
    elif command -v vi &> /dev/null; then
        vi "$CONFIG_DIR/config.yaml"
    else
        print_error "未找到可用的编辑器"
        print_info "配置文件位置: $CONFIG_DIR/config.yaml"
        exit 1
    fi
    
    print_success "配置已保存"
    
    # 询问是否重启服务
    if systemctl is-active --quiet $SERVICE_NAME; then
        read -p "是否重启服务以应用配置？(y/n): " -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            systemctl restart $SERVICE_NAME
            print_success "服务已重启"
        fi
    fi
}

# 显示使用说明
show_usage() {
    echo ""
    echo "=========================================="
    echo "  Clash Meta 管理脚本"
    echo "=========================================="
    echo ""
    echo "使用方法: $0 [命令]"
    echo ""
    echo "命令列表:"
    echo "  install          安装 Clash Meta"
    echo "  uninstall        卸载 Clash Meta"
    echo "  start            启动服务"
    echo "  stop             停止服务"
    echo "  restart          重启服务"
    echo "  status           查看状态"
    echo "  logs             查看日志"
    echo "  enable           启用开机自启"
    echo "  disable          禁用开机自启"
    echo "  update           更新订阅"
    echo "  edit             编辑配置文件"
    echo "  help             显示帮助信息"
    echo ""
    echo "配置文件: $CONFIG_DIR/config.yaml"
    echo "Web 面板: http://服务器IP:9090/ui"
    echo ""
}

# 显示交互式菜单
show_menu() {
    clear
    echo "=========================================="
    echo "       Clash Meta 管理脚本"
    echo "=========================================="
    echo ""
    
    # 检查安装状态
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        echo -e "${GREEN}[已安装]${NC}"
        
        # 检查服务状态
        if systemctl is-active --quiet $SERVICE_NAME 2>/dev/null; then
            echo -e "服务状态: ${GREEN}运行中${NC}"
        else
            echo -e "服务状态: ${RED}已停止${NC}"
        fi
        
        # 检查开机自启
        if systemctl is-enabled --quiet $SERVICE_NAME 2>/dev/null; then
            echo -e "开机自启: ${GREEN}已启用${NC}"
        else
            echo -e "开机自启: ${YELLOW}未启用${NC}"
        fi
    else
        echo -e "${YELLOW}[未安装]${NC}"
    fi
    
    echo ""
    echo "=========================================="
    echo ""
    echo "  1) 安装 Clash Meta"
    echo "  2) 卸载 Clash Meta"
    echo "  3) 启动服务"
    echo "  4) 停止服务"
    echo "  5) 重启服务"
    echo "  6) 查看状态"
    echo "  7) 查看日志"
    echo "  8) 启用开机自启"
    echo "  9) 禁用开机自启"
    echo " 10) 更新订阅"
    echo " 11) 编辑配置"
    echo "  0) 退出"
    echo ""
    echo "=========================================="
    echo ""
}

# 交互式菜单主循环
interactive_menu() {
    while true; do
        show_menu
        read -p "请选择操作 [0-11]: " choice
        echo ""
        
        case $choice in
            1)
                install_clash_meta
                read -p "按回车键继续..."
                ;;
            2)
                uninstall_clash_meta
                read -p "按回车键继续..."
                ;;
            3)
                start_service
                read -p "按回车键继续..."
                ;;
            4)
                stop_service
                read -p "按回车键继续..."
                ;;
            5)
                restart_service
                read -p "按回车键继续..."
                ;;
            6)
                show_status
                read -p "按回车键继续..."
                ;;
            7)
                show_logs
                ;;
            8)
                enable_service
                read -p "按回车键继续..."
                ;;
            9)
                disable_service
                read -p "按回车键继续..."
                ;;
            10)
                update_subscription
                read -p "按回车键继续..."
                ;;
            11)
                edit_config
                read -p "按回车键继续..."
                ;;
            0)
                print_info "退出脚本"
                exit 0
                ;;
            *)
                print_error "无效的选择"
                read -p "按回车键继续..."
                ;;
        esac
    done
}

# 主函数
main() {
    # 如果没有参数，显示交互式菜单
    if [ $# -eq 0 ]; then
        interactive_menu
        exit 0
    fi
    
    # 如果有参数，执行对应命令
    case "${1:-}" in
        install)
            install_clash_meta
            ;;
        uninstall)
            uninstall_clash_meta
            ;;
        start)
            start_service
            ;;
        stop)
            stop_service
            ;;
        restart)
            restart_service
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        enable)
            enable_service
            ;;
        disable)
            disable_service
            ;;
        update)
            update_subscription
            ;;
        edit)
            edit_config
            ;;
        menu)
            interactive_menu
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            print_error "未知命令: $1"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
