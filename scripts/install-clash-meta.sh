#!/bin/bash

# Clash Meta 一键安装脚本
# 适用于 Linux 服务器

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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
        print_error "请使用 root 权限运行此脚本"
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
    print_info "检测到系统架构: $ARCH"
}

# 获取最新版本号
get_latest_version() {
    print_info "获取最新版本信息..."
    
    # 方法1: 使用 GitHub API
    LATEST_VERSION=$(curl -s --connect-timeout 10 https://api.github.com/repos/MetaCubeX/mihomo/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    # 方法2: 如果 API 失败，尝试从 HTML 页面获取
    if [ -z "$LATEST_VERSION" ]; then
        print_warning "GitHub API 访问失败，尝试备用方法..."
        LATEST_VERSION=$(curl -sL --connect-timeout 10 https://github.com/MetaCubeX/mihomo/releases/latest | grep -oP 'tag/\K[^"]+' | head -1)
    fi
    
    # 方法3: 如果还是失败，使用固定版本
    if [ -z "$LATEST_VERSION" ]; then
        print_warning "无法自动获取版本，使用默认版本"
        LATEST_VERSION="v1.18.10"
        print_info "使用版本: $LATEST_VERSION"
        read -p "是否继续？(y/n): " -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        print_success "最新版本: $LATEST_VERSION"
    fi
}

# 下载 Clash Meta
download_clash_meta() {
    print_info "下载 Clash Meta..."
    
    local download_url="https://github.com/MetaCubeX/mihomo/releases/download/${LATEST_VERSION}/mihomo-linux-${ARCH}-${LATEST_VERSION}.gz"
    local temp_file="/tmp/mihomo-linux-${ARCH}.gz"
    
    print_info "下载地址: $download_url"
    
    # 尝试下载，显示进度
    if ! curl -L --progress-bar --connect-timeout 30 --max-time 300 -o "$temp_file" "$download_url"; then
        print_error "下载失败"
        print_info "请检查网络连接或手动下载："
        print_info "$download_url"
        exit 1
    fi
    
    # 验证文件是否下载成功
    if [ ! -f "$temp_file" ] || [ ! -s "$temp_file" ]; then
        print_error "下载的文件无效"
        exit 1
    fi
    
    print_success "下载完成"
}

# 安装 Clash Meta
install_clash_meta() {
    print_info "安装 Clash Meta..."
    
    local temp_file="/tmp/mihomo-linux-${ARCH}.gz"
    
    # 解压
    gunzip -f "$temp_file"
    
    # 移动到系统路径
    mv "/tmp/mihomo-linux-${ARCH}" /usr/local/bin/mihomo
    chmod +x /usr/local/bin/mihomo
    
    # 创建软链接
    ln -sf /usr/local/bin/mihomo /usr/local/bin/clash-meta
    
    print_success "安装完成"
}

# 创建配置目录
create_config_dir() {
    print_info "创建配置目录..."
    
    mkdir -p /etc/clash-meta
    mkdir -p /etc/clash-meta/profiles
    
    print_success "配置目录创建完成: /etc/clash-meta"
}

# 创建示例配置文件
create_sample_config() {
    print_info "创建示例配置文件..."
    
    cat > /etc/clash-meta/config.yaml << 'EOF'
# Clash Meta 配置文件
# 适用于命令行版本

# 混合端口（HTTP + SOCKS5）
mixed-port: 7890

# HTTP 代理端口
port: 7891

# SOCKS5 代理端口
socks-port: 7892

# 允许局域网连接
allow-lan: true

# 绑定地址
bind-address: '*'

# 运行模式: rule / global / direct
mode: rule

# 日志级别: info / warning / error / debug / silent
log-level: info

# IPv6 支持
ipv6: true

# 外部控制器
external-controller: 0.0.0.0:9090

# 外部控制器密钥（建议修改）
secret: ""

# 统一延迟
unified-delay: true

# TCP 并发
tcp-concurrent: true

# DNS 配置
dns:
  enable: true
  listen: 0.0.0.0:1053
  ipv6: true
  enhanced-mode: fake-ip
  fake-ip-range: 198.18.0.1/16
  fake-ip-filter:
    - '*.lan'
    - 'localhost.ptlogin2.qq.com'
  nameserver:
    - 223.5.5.5
    - 119.29.29.29
    - 114.114.114.114
  fallback:
    - 8.8.8.8
    - 1.1.1.1
    - tls://dns.google
  fallback-filter:
    geoip: true
    geoip-code: CN

# TUN 配置（可选，需要 root 权限）
tun:
  enable: false
  stack: system
  dns-hijack:
    - any:53
  auto-route: true
  auto-detect-interface: true

# 代理节点（请替换为您的节点信息）
proxies: []

# 代理组
proxy-groups:
  - name: PROXY
    type: select
    proxies:
      - DIRECT

# 规则
rules:
  - DOMAIN-SUFFIX,local,DIRECT
  - IP-CIDR,127.0.0.0/8,DIRECT
  - IP-CIDR,172.16.0.0/12,DIRECT
  - IP-CIDR,192.168.0.0/16,DIRECT
  - IP-CIDR,10.0.0.0/8,DIRECT
  - GEOIP,CN,DIRECT
  - MATCH,PROXY
EOF

    print_success "示例配置文件已创建: /etc/clash-meta/config.yaml"
    print_warning "请编辑配置文件或使用订阅链接更新配置"
}

# 创建 systemd 服务
create_systemd_service() {
    print_info "创建 systemd 服务..."
    
    cat > /etc/systemd/system/clash-meta.service << 'EOF'
[Unit]
Description=Clash Meta Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/mihomo -d /etc/clash-meta
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    
    print_success "systemd 服务已创建"
}

# 下载订阅配置
download_subscription() {
    if [ -n "$SUBSCRIPTION_URL" ]; then
        print_info "下载订阅配置..."
        
        if curl -L --connect-timeout 30 --max-time 60 -o /etc/clash-meta/config.yaml "$SUBSCRIPTION_URL"; then
            # 验证配置文件
            if [ -s /etc/clash-meta/config.yaml ]; then
                print_success "订阅配置下载成功"
            else
                print_warning "订阅配置文件为空，请检查订阅链接"
                create_sample_config
            fi
        else
            print_warning "订阅配置下载失败，使用示例配置"
            create_sample_config
        fi
    fi
}

# 配置系统代理
configure_system_proxy() {
    print_info "配置系统代理环境变量..."
    
    cat > /etc/profile.d/clash-meta.sh << 'EOF'
# Clash Meta 代理配置
export http_proxy=http://127.0.0.1:7890
export https_proxy=http://127.0.0.1:7890
export no_proxy=localhost,127.0.0.1,::1
EOF

    print_success "系统代理配置已添加到 /etc/profile.d/clash-meta.sh"
    print_info "运行 'source /etc/profile.d/clash-meta.sh' 或重新登录以启用代理"
}

# 显示使用说明
show_usage() {
    echo ""
    echo "=========================================="
    echo "  Clash Meta 安装完成！"
    echo "=========================================="
    echo ""
    echo "配置文件位置: /etc/clash-meta/config.yaml"
    echo "可执行文件: /usr/local/bin/mihomo"
    echo ""
    echo "常用命令:"
    echo "  启动服务: systemctl start clash-meta"
    echo "  停止服务: systemctl stop clash-meta"
    echo "  重启服务: systemctl restart clash-meta"
    echo "  查看状态: systemctl status clash-meta"
    echo "  开机自启: systemctl enable clash-meta"
    echo "  查看日志: journalctl -u clash-meta -f"
    echo ""
    echo "更新订阅配置:"
    echo "  wget -O /etc/clash-meta/config.yaml \"您的订阅链接\""
    echo "  systemctl restart clash-meta"
    echo ""
    echo "Web 控制面板:"
    echo "  访问: http://服务器IP:9090/ui"
    echo "  推荐面板: https://github.com/MetaCubeX/metacubexd"
    echo ""
    echo "启用系统代理:"
    echo "  source /etc/profile.d/clash-meta.sh"
    echo ""
    echo "测试连接:"
    echo "  curl -I https://www.google.com"
    echo ""
    echo "=========================================="
}

# 主函数
main() {
    echo ""
    echo "=========================================="
    echo "  Clash Meta 一键安装脚本"
    echo "=========================================="
    echo ""
    
    # 检查 root 权限
    check_root
    
    # 检测系统架构
    detect_arch
    
    # 询问是否提供订阅链接
    read -p "是否提供订阅链接？(y/n): " -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "请输入订阅链接: " SUBSCRIPTION_URL
        if [ -z "$SUBSCRIPTION_URL" ]; then
            print_warning "未输入订阅链接，将使用示例配置"
        fi
    fi
    
    # 获取最新版本
    get_latest_version
    
    # 下载
    download_clash_meta
    
    # 安装
    install_clash_meta
    
    # 创建配置目录
    create_config_dir
    
    # 下载订阅或创建示例配置
    if [ -n "$SUBSCRIPTION_URL" ]; then
        download_subscription
    else
        create_sample_config
    fi
    
    # 创建 systemd 服务
    create_systemd_service
    
    # 配置系统代理
    configure_system_proxy
    
    # 显示使用说明
    show_usage
    
    # 询问是否立即启动
    echo ""
    read -p "是否立即启动 Clash Meta 服务？(y/n): " -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        systemctl start clash-meta
        systemctl enable clash-meta
        sleep 2
        systemctl status clash-meta
    fi
}

# 运行主函数
main
