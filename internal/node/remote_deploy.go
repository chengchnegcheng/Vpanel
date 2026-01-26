// Package node provides node management functionality.
package node

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"v/internal/logger"
)

// RemoteDeployService handles remote agent deployment.
type RemoteDeployService struct {
	logger logger.Logger
}

// NewRemoteDeployService creates a new remote deploy service.
func NewRemoteDeployService(log logger.Logger) *RemoteDeployService {
	return &RemoteDeployService{
		logger: log,
	}
}

// DeployConfig contains configuration for remote deployment.
type DeployConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	PrivateKey      string `json:"private_key,omitempty"`
	PanelURL        string `json:"panel_url"`
	NodeToken       string `json:"node_token"`
	AgentBinaryPath string `json:"agent_binary_path,omitempty"` // 本地 Agent 二进制文件路径
}

// DeployResult contains the result of a deployment.
type DeployResult struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Steps   []string `json:"steps"`
	Logs    string   `json:"logs"`
}

// Deploy deploys the agent to a remote server.
func (s *RemoteDeployService) Deploy(ctx context.Context, config *DeployConfig) (*DeployResult, error) {
	result := &DeployResult{
		Steps: []string{},
	}

	var logBuffer bytes.Buffer

	// Step 1: Connect to remote server
	result.Steps = append(result.Steps, "连接到远程服务器...")
	client, err := s.connectSSH(config)
	if err != nil {
		result.Message = fmt.Sprintf("SSH 连接失败: %v", err)
		return result, err
	}
	defer client.Close()

	logBuffer.WriteString("✓ SSH 连接成功\n")
	s.logger.Info("SSH connected", logger.F("host", config.Host))

	// Step 2: Check system requirements
	result.Steps = append(result.Steps, "检查系统要求...")
	if err := s.checkSystemRequirements(client, &logBuffer); err != nil {
		result.Message = fmt.Sprintf("系统检查失败: %v", err)
		result.Logs = logBuffer.String()
		return result, err
	}

	// Step 3: Install dependencies
	result.Steps = append(result.Steps, "安装依赖...")
	if err := s.installDependencies(client, &logBuffer); err != nil {
		result.Message = fmt.Sprintf("依赖安装失败: %v", err)
		result.Logs = logBuffer.String()
		return result, err
	}

	// Step 4: Download and install agent
	result.Steps = append(result.Steps, "下载并安装 Agent...")
	if err := s.installAgent(client, config, &logBuffer); err != nil {
		result.Message = fmt.Sprintf("Agent 安装失败: %v", err)
		result.Logs = logBuffer.String()
		return result, err
	}

	// Step 5: Install Xray
	result.Steps = append(result.Steps, "安装 Xray...")
	if err := s.installXray(client, &logBuffer); err != nil {
		result.Message = fmt.Sprintf("Xray 安装失败: %v", err)
		result.Logs = logBuffer.String()
		return result, err
	}

	// Step 6: Configure agent
	result.Steps = append(result.Steps, "配置 Agent...")
	if err := s.configureAgent(client, config, &logBuffer); err != nil {
		result.Message = fmt.Sprintf("Agent 配置失败: %v", err)
		result.Logs = logBuffer.String()
		return result, err
	}

	// Step 7: Start agent service
	result.Steps = append(result.Steps, "启动 Agent 服务...")
	if err := s.startAgentService(client, &logBuffer); err != nil {
		result.Message = fmt.Sprintf("Agent 启动失败: %v", err)
		result.Logs = logBuffer.String()
		return result, err
	}

	// Step 8: Verify installation
	result.Steps = append(result.Steps, "验证安装...")
	if err := s.verifyInstallation(client, &logBuffer); err != nil {
		result.Message = fmt.Sprintf("安装验证失败: %v", err)
		result.Logs = logBuffer.String()
		return result, err
	}

	result.Success = true
	result.Message = "Agent 部署成功"
	result.Logs = logBuffer.String()

	s.logger.Info("Agent deployed successfully", logger.F("host", config.Host))

	return result, nil
}

// connectSSH establishes an SSH connection.
func (s *RemoteDeployService) connectSSH(config *DeployConfig) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	// Password authentication
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
		// Also try keyboard-interactive for servers that require it
		authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range answers {
				answers[i] = config.Password
			}
			return answers, nil
		}))
	}

	// Private key authentication
	if config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("私钥解析失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("未提供认证方式，请提供密码或私钥")
	}

	port := config.Port
	if port == 0 {
		port = 22
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
		// 添加更多加密算法和密钥交换算法以提高兼容性
		Config: ssh.Config{
			// 支持更多加密算法（按优先级排序）
			Ciphers: []string{
				"aes128-gcm@openssh.com",
				"chacha20-poly1305@openssh.com",
				"aes256-gcm@openssh.com",
				"aes128-ctr",
				"aes192-ctr",
				"aes256-ctr",
				"aes128-cbc",
				"aes192-cbc",
				"aes256-cbc",
				"3des-cbc",
			},
			// 添加密钥交换算法
			KeyExchanges: []string{
				"curve25519-sha256",
				"curve25519-sha256@libssh.org",
				"ecdh-sha2-nistp256",
				"ecdh-sha2-nistp384",
				"ecdh-sha2-nistp521",
				"diffie-hellman-group14-sha256",
				"diffie-hellman-group14-sha1",
				"diffie-hellman-group1-sha1",
			},
		},
	}

	addr := fmt.Sprintf("%s:%d", config.Host, port)
	
	s.logger.Info("尝试连接 SSH",
		logger.F("address", addr),
		logger.F("username", config.Username))

	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		// Provide more helpful error messages
		if strings.Contains(err.Error(), "connection refused") {
			return nil, fmt.Errorf("连接被拒绝，请检查: 1) SSH 服务是否运行 2) 端口 %d 是否正确 3) 防火墙是否允许", port)
		}
		if strings.Contains(err.Error(), "no supported methods remain") {
			return nil, fmt.Errorf("认证失败，请检查: 1) 用户名是否正确 2) 密码/私钥是否正确 3) 服务器是否允许密码认证")
		}
		if strings.Contains(err.Error(), "handshake failed") {
			return nil, fmt.Errorf("SSH 握手失败，请检查: 1) 服务器 SSH 版本是否兼容 2) 网络连接是否稳定 3) 是否有中间代理")
		}
		if strings.Contains(err.Error(), "timeout") {
			return nil, fmt.Errorf("连接超时，请检查: 1) 服务器地址是否正确 2) 网络是否可达 3) 防火墙规则")
		}
		return nil, fmt.Errorf("SSH 连接失败: %w", err)
	}

	s.logger.Info("SSH 连接成功", logger.F("address", addr))
	return client, nil
}

// executeCommand executes a command on the remote server.
func (s *RemoteDeployService) executeCommand(client *ssh.Client, command string, logBuffer *bytes.Buffer) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	// Capture output
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	s.logger.Debug("执行命令", logger.F("command", command))

	if err := session.Run(command); err != nil {
		// 记录详细错误信息
		logBuffer.WriteString(fmt.Sprintf("✗ 命令执行失败\n"))
		if stdout.Len() > 0 {
			logBuffer.WriteString(fmt.Sprintf("输出: %s\n", stdout.String()))
		}
		if stderr.Len() > 0 {
			logBuffer.WriteString(fmt.Sprintf("错误: %s\n", stderr.String()))
		}
		return fmt.Errorf("命令执行失败: %w", err)
	}

	// 记录成功输出
	if stdout.Len() > 0 {
		logBuffer.WriteString(stdout.String())
		if !strings.HasSuffix(stdout.String(), "\n") {
			logBuffer.WriteString("\n")
		}
	}
	if stderr.Len() > 0 {
		logBuffer.WriteString(stderr.String())
		if !strings.HasSuffix(stderr.String(), "\n") {
			logBuffer.WriteString("\n")
		}
	}

	return nil
}

// checkSystemRequirements checks if the system meets requirements.
func (s *RemoteDeployService) checkSystemRequirements(client *ssh.Client, logBuffer *bytes.Buffer) error {
	commands := []string{
		"uname -a",                    // OS info
		"cat /etc/os-release || true", // Distribution info
		"free -h",                     // Memory info
		"df -h /",                     // Disk space
	}

	for _, cmd := range commands {
		if err := s.executeCommand(client, cmd, logBuffer); err != nil {
			// Don't fail on info commands
			s.logger.Warn("System check command failed", logger.F("command", cmd))
		}
	}

	logBuffer.WriteString("✓ 系统检查完成\n")
	return nil
}

// installDependencies installs required dependencies.
func (s *RemoteDeployService) installDependencies(client *ssh.Client, logBuffer *bytes.Buffer) error {
	// Detect package manager and install dependencies
	script := `
set -e
echo "检测操作系统和包管理器..."

# 检测操作系统
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    OS_VERSION=$VERSION_ID
    echo "操作系统: $NAME $VERSION_ID"
else
    echo "警告: 无法检测操作系统版本"
fi

# 根据不同系统安装依赖
if command -v apt-get &> /dev/null; then
    echo "使用 apt-get 安装依赖 (Ubuntu/Debian)..."
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq || true
    apt-get install -y curl wget unzip tar ca-certificates 2>&1 || {
        echo "警告: 部分依赖安装失败，继续..."
    }
elif command -v dnf &> /dev/null; then
    echo "使用 dnf 安装依赖 (Fedora/CentOS 8+/AlmaLinux/Rocky)..."
    dnf install -y curl wget unzip tar ca-certificates 2>&1 || {
        echo "警告: 部分依赖安装失败，继续..."
    }
elif command -v yum &> /dev/null; then
    echo "使用 yum 安装依赖 (CentOS/RHEL 7)..."
    yum install -y curl wget unzip tar ca-certificates 2>&1 || {
        echo "警告: 部分依赖安装失败，继续..."
    }
elif command -v apk &> /dev/null; then
    echo "使用 apk 安装依赖 (Alpine)..."
    apk add --no-cache curl wget unzip tar ca-certificates || {
        echo "警告: 部分依赖安装失败，继续..."
    }
elif command -v zypper &> /dev/null; then
    echo "使用 zypper 安装依赖 (openSUSE)..."
    zypper install -y curl wget unzip tar ca-certificates || {
        echo "警告: 部分依赖安装失败，继续..."
    }
else
    echo "警告: 未检测到支持的包管理器，跳过依赖安装"
fi

# 验证关键命令是否可用
echo "验证依赖..."
command -v curl >/dev/null 2>&1 || echo "警告: curl 未安装"
command -v wget >/dev/null 2>&1 || echo "警告: wget 未安装"
command -v unzip >/dev/null 2>&1 || echo "警告: unzip 未安装"
command -v tar >/dev/null 2>&1 || echo "警告: tar 未安装"

echo "依赖检查完成"
`

	if err := s.executeCommand(client, script, logBuffer); err != nil {
		// 不因为依赖安装失败而中断，继续尝试
		s.logger.Warn("依赖安装有警告，但继续执行", logger.F("error", err))
		logBuffer.WriteString("⚠ 依赖安装有警告，继续执行\n")
		return nil
	}

	logBuffer.WriteString("✓ 依赖安装完成\n")
	return nil
}

// installAgent uploads and installs the agent binary.
func (s *RemoteDeployService) installAgent(client *ssh.Client, config *DeployConfig, logBuffer *bytes.Buffer) error {
	// 先创建目录
	script := `
# 创建目录
mkdir -p /usr/local/bin
mkdir -p /etc/vpanel
mkdir -p /var/log/vpanel

# 检测系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    armv7l)
        ARCH="arm"
        ;;
    *)
        echo "错误: 不支持的架构: $ARCH"
        exit 1
        ;;
esac

echo "检测到系统架构: $ARCH"
`

	if err := s.executeCommand(client, script, logBuffer); err != nil {
		return err
	}

	// 尝试通过 Panel URL 下载（如果配置了公网地址）
	if config.PanelURL != "" && !strings.Contains(config.PanelURL, "localhost") && !strings.Contains(config.PanelURL, "127.0.0.1") {
		logBuffer.WriteString("尝试从 Panel 服务器下载 Agent...\n")
		downloadScript := fmt.Sprintf(`
PANEL_URL="%s"
DOWNLOAD_URL="${PANEL_URL}/api/admin/nodes/agent/download?arch=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/armv7l/arm/')"

echo "下载地址: ${DOWNLOAD_URL}"

if command -v wget &> /dev/null; then
    wget --no-check-certificate -O /usr/local/bin/vpanel-agent "$DOWNLOAD_URL" 2>&1 && chmod +x /usr/local/bin/vpanel-agent && echo "✓ Agent 下载成功" && exit 0
elif command -v curl &> /dev/null; then
    curl -k -L -o /usr/local/bin/vpanel-agent "$DOWNLOAD_URL" 2>&1 && chmod +x /usr/local/bin/vpanel-agent && echo "✓ Agent 下载成功" && exit 0
fi

echo "下载失败，将使用上传方式"
exit 1
`, config.PanelURL)

		// 尝试下载，失败也不报错
		if err := s.executeCommand(client, downloadScript, logBuffer); err == nil {
			logBuffer.WriteString("✓ Agent 安装完成（通过下载）\n")
			return nil
		}
	}

	// 下载失败或未配置公网地址，使用上传方式
	logBuffer.WriteString("使用上传方式安装 Agent...\n")
	
	// 通过 base64 编码上传 Agent 二进制文件
	// 这里需要读取本地编译好的 Agent 文件
	agentPath := config.AgentBinaryPath
	if agentPath == "" {
		agentPath = "./bin/vpanel-agent" // 默认路径
	}

	// 读取 Agent 文件内容
	agentData, err := s.readAgentBinary(agentPath)
	if err != nil {
		logBuffer.WriteString(fmt.Sprintf("✗ 无法读取 Agent 文件: %v\n", err))
		logBuffer.WriteString("请确保已编译 Agent: cd cmd/agent && go build -o ../../bin/vpanel-agent\n")
		return fmt.Errorf("无法读取 Agent 文件: %w", err)
	}

	// 使用 base64 编码传输
	logBuffer.WriteString(fmt.Sprintf("正在上传 Agent (大小: %d bytes)...\n", len(agentData)))
	
	// 分块上传（每次 50KB）
	chunkSize := 50 * 1024
	totalChunks := (len(agentData) + chunkSize - 1) / chunkSize
	
	// 先清空目标文件
	if err := s.executeCommand(client, "rm -f /usr/local/bin/vpanel-agent.b64", logBuffer); err != nil {
		return fmt.Errorf("清空临时文件失败: %w", err)
	}

	for i := 0; i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(agentData) {
			end = len(agentData)
		}
		
		chunk := agentData[start:end]
		encoded := base64Encode(chunk)
		
		uploadScript := fmt.Sprintf("echo '%s' >> /usr/local/bin/vpanel-agent.b64", encoded)
		if err := s.executeCommand(client, uploadScript, logBuffer); err != nil {
			return fmt.Errorf("上传分块 %d/%d 失败: %w", i+1, totalChunks, err)
		}
		
		if (i+1)%10 == 0 || i == totalChunks-1 {
			logBuffer.WriteString(fmt.Sprintf("上传进度: %d/%d\n", i+1, totalChunks))
		}
	}

	// 解码并安装
	decodeScript := `
echo "正在解码 Agent..."
base64 -d /usr/local/bin/vpanel-agent.b64 > /usr/local/bin/vpanel-agent
chmod +x /usr/local/bin/vpanel-agent
rm -f /usr/local/bin/vpanel-agent.b64

# 验证文件
FILE_SIZE=$(stat -c%s /usr/local/bin/vpanel-agent 2>/dev/null || stat -f%z /usr/local/bin/vpanel-agent 2>/dev/null || echo "0")
echo "Agent 文件大小: ${FILE_SIZE} bytes"

if [ "$FILE_SIZE" -lt 1000 ]; then
    echo "错误: Agent 文件大小异常"
    exit 1
fi

echo "✓ Agent 上传成功"
`

	if err := s.executeCommand(client, decodeScript, logBuffer); err != nil {
		return err
	}

	logBuffer.WriteString("✓ Agent 安装完成（通过上传）\n")
	return nil
}

// installXray installs Xray using the official script.
func (s *RemoteDeployService) installXray(client *ssh.Client, logBuffer *bytes.Buffer) error {
	script := `
# 检查 Xray 是否已安装
if command -v xray &> /dev/null; then
    echo "Xray 已安装，跳过"
    xray version | head -n 1
else
    echo "正在安装 Xray..."
    
    # 尝试使用官方安装脚本
    if curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh -o /tmp/install-xray.sh 2>&1; then
        bash /tmp/install-xray.sh install
        rm -f /tmp/install-xray.sh
    else
        echo "官方脚本下载失败，尝试手动安装..."
        
        # 获取最新版本
        XRAY_VERSION=$(curl -s https://api.github.com/repos/XTLS/Xray-core/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' || echo "v1.8.4")
        
        # 检测系统架构
        ARCH=$(uname -m)
        case "$ARCH" in
            x86_64)
                XRAY_ARCH="linux-64"
                ;;
            aarch64|arm64)
                XRAY_ARCH="linux-arm64-v8a"
                ;;
            armv7l)
                XRAY_ARCH="linux-arm32-v7a"
                ;;
            *)
                echo "错误: 不支持的架构: $ARCH"
                exit 1
                ;;
        esac
        
        XRAY_URL="https://github.com/XTLS/Xray-core/releases/download/${XRAY_VERSION}/Xray-${XRAY_ARCH}.zip"
        
        echo "下载 Xray ${XRAY_VERSION} for ${XRAY_ARCH}..."
        
        # 尝试使用 wget 或 curl 下载
        if command -v wget &> /dev/null; then
            wget -O /tmp/xray.zip "$XRAY_URL" || exit 1
        elif command -v curl &> /dev/null; then
            curl -L -o /tmp/xray.zip "$XRAY_URL" || exit 1
        else
            echo "错误: 未找到 wget 或 curl"
            exit 1
        fi
        
        echo "解压并安装..."
        mkdir -p /tmp/xray
        unzip -o /tmp/xray.zip -d /tmp/xray
        mv /tmp/xray/xray /usr/local/bin/
        chmod +x /usr/local/bin/xray
        rm -rf /tmp/xray /tmp/xray.zip
    fi
    
    # 验证安装
    if command -v xray &> /dev/null; then
        echo "✓ Xray 安装成功"
        xray version | head -n 1
    else
        echo "✗ Xray 安装失败"
        exit 1
    fi
fi

# 创建 Xray 配置目录
mkdir -p /etc/xray
mkdir -p /var/log/xray
echo "✓ Xray 目录创建完成"
`

	if err := s.executeCommand(client, script, logBuffer); err != nil {
		return err
	}

	logBuffer.WriteString("✓ Xray 安装完成\n")
	return nil
}

// configureAgent creates the agent configuration file.
func (s *RemoteDeployService) configureAgent(client *ssh.Client, config *DeployConfig, logBuffer *bytes.Buffer) error {
	// Create agent config - 正确的配置结构
	agentConfig := fmt.Sprintf(`node:
  token: "%s"

panel:
  url: "%s"

xray:
  binary_path: "/usr/local/bin/xray"
  config_path: "/etc/xray/config.json"

health:
  port: 8081
`, config.NodeToken, config.PanelURL)

	// 使用 base64 编码配置内容，避免 heredoc 格式问题
	encoded := base64Encode([]byte(agentConfig))
	
	// Write config file using base64
	script := fmt.Sprintf(`echo '%s' | base64 -d > /etc/vpanel/agent.yaml
chmod 644 /etc/vpanel/agent.yaml
echo "配置文件已创建"
cat /etc/vpanel/agent.yaml
`, encoded)

	if err := s.executeCommand(client, script, logBuffer); err != nil {
		return err
	}

	// Create systemd service
	serviceFile := `[Unit]
Description=V Panel Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/vpanel-agent -config /etc/vpanel/agent.yaml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`

	script = fmt.Sprintf(`cat > /etc/systemd/system/vpanel-agent.service <<'EOF'
%s
EOF
systemctl daemon-reload
`, serviceFile)

	if err := s.executeCommand(client, script, logBuffer); err != nil {
		return err
	}

	logBuffer.WriteString("✓ Agent 配置完成\n")
	return nil
}

// startAgentService starts the agent service.
func (s *RemoteDeployService) startAgentService(client *ssh.Client, logBuffer *bytes.Buffer) error {
	script := `
systemctl enable vpanel-agent
systemctl start vpanel-agent
sleep 3

# 检查服务状态
if systemctl is-active --quiet vpanel-agent; then
    systemctl status vpanel-agent --no-pager
else
    echo "✗ Agent 服务启动失败"
    echo ""
    echo "=== 服务状态 ==="
    systemctl status vpanel-agent --no-pager || true
    echo ""
    echo "=== 最近日志 ==="
    journalctl -u vpanel-agent -n 50 --no-pager || true
    echo ""
    echo "=== 配置文件检查 ==="
    cat /etc/vpanel/agent.yaml || true
    exit 1
fi
`

	if err := s.executeCommand(client, script, logBuffer); err != nil {
		return err
	}

	logBuffer.WriteString("✓ Agent 服务启动完成\n")
	return nil
}

// verifyInstallation verifies the installation.
func (s *RemoteDeployService) verifyInstallation(client *ssh.Client, logBuffer *bytes.Buffer) error {
	script := `
# Check if agent is running
if systemctl is-active --quiet vpanel-agent; then
    echo "✓ Agent 服务运行中"
else
    echo "✗ Agent 服务未运行"
    exit 1
fi

# Check if Xray is installed
if command -v xray &> /dev/null; then
    echo "✓ Xray 已安装: $(xray version | head -1)"
else
    echo "✗ Xray 未安装"
    exit 1
fi

# Check config file
if [ -f /etc/vpanel/agent.yaml ]; then
    echo "✓ 配置文件存在"
else
    echo "✗ 配置文件不存在"
    exit 1
fi
`

	if err := s.executeCommand(client, script, logBuffer); err != nil {
		return err
	}

	logBuffer.WriteString("✓ 安装验证完成\n")
	return nil
}

// UploadAgentBinary uploads the agent binary to the remote server.
func (s *RemoteDeployService) UploadAgentBinary(client *ssh.Client, localPath string) error {
	// Open SFTP session
	sftp, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SFTP session: %w", err)
	}
	defer sftp.Close()

	// TODO: Implement file upload using SFTP
	// This requires the golang.org/x/crypto/ssh package with SFTP support

	return nil
}

// GetDeployScript returns the deployment script for manual installation.
func (s *RemoteDeployService) GetDeployScript(panelURL, nodeToken string) string {
	return fmt.Sprintf(`#!/bin/bash
# V Panel Agent 自动部署脚本
# 支持: Ubuntu, Debian, CentOS, AlmaLinux, Rocky Linux, Fedora
# 使用方法: bash install-agent.sh

set -e

PANEL_URL="%s"
NODE_TOKEN="%s"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "  V Panel Agent 自动部署脚本"
echo "=========================================="
echo ""

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
    exit 1
fi

echo "开始部署 V Panel Agent..."

# 检测操作系统
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    echo -e "${GREEN}检测到操作系统: $NAME $VERSION_ID${NC}"
else
    echo -e "${YELLOW}警告: 无法检测操作系统版本${NC}"
    OS="unknown"
fi

# 检测系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    armv7l)
        ARCH="arm"
        ;;
    *)
        echo -e "${RED}错误: 不支持的架构: $ARCH${NC}"
        exit 1
        ;;
esac

echo "检测到系统架构: $ARCH"

# 安装依赖
echo "正在安装依赖..."
if command -v apt-get &> /dev/null; then
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    apt-get install -y -qq curl wget unzip tar ca-certificates
elif command -v dnf &> /dev/null; then
    dnf install -y -q curl wget unzip tar ca-certificates
elif command -v yum &> /dev/null; then
    yum install -y -q curl wget unzip tar ca-certificates
else
    echo -e "${RED}错误: 不支持的包管理器${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 依赖安装完成${NC}"

# 安装 Xray
if ! command -v xray &> /dev/null; then
    echo "正在安装 Xray..."
    if curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh -o /tmp/install-xray.sh 2>&1; then
        bash /tmp/install-xray.sh install
        rm -f /tmp/install-xray.sh
        echo -e "${GREEN}✓ Xray 安装完成${NC}"
    else
        echo -e "${YELLOW}警告: Xray 安装失败，请手动安装${NC}"
    fi
else
    echo -e "${GREEN}✓ Xray 已安装${NC}"
    xray version | head -1
fi

# 创建目录
echo "正在创建目录..."
mkdir -p /etc/vpanel
mkdir -p /etc/xray
mkdir -p /var/log/vpanel
mkdir -p /var/log/xray
mkdir -p /usr/local/bin

# 下载 Agent 二进制文件
echo "正在下载 Agent..."
# 注意: 请替换为实际的下载地址
# AGENT_VERSION="latest"
# DOWNLOAD_URL="https://github.com/your-org/vpanel/releases/download/${AGENT_VERSION}/vpanel-agent-linux-${ARCH}"
# wget -O /usr/local/bin/vpanel-agent "$DOWNLOAD_URL" || curl -L -o /usr/local/bin/vpanel-agent "$DOWNLOAD_URL"

# 临时方案: 提示用户手动上传
echo -e "${YELLOW}⚠ 注意: 当前需要手动上传 agent 二进制文件${NC}"
echo "请将编译好的 vpanel-agent 上传到 /usr/local/bin/vpanel-agent"
echo ""
echo "编译命令:"
echo "  cd cmd/agent && go build -o vpanel-agent"
echo ""
echo "上传命令:"
echo "  scp vpanel-agent root@YOUR_SERVER:/usr/local/bin/"
echo ""
read -p "按回车键继续 (确认已上传 agent 二进制文件)..."

# 设置权限
chmod +x /usr/local/bin/vpanel-agent

# 创建配置文件
echo "正在创建配置文件..."
cat > /etc/vpanel/agent.yaml <<EOF
node:
  token: "$NODE_TOKEN"

panel:
  url: "$PANEL_URL"

xray:
  binary_path: "/usr/local/bin/xray"
  config_path: "/etc/xray/config.json"

health:
  port: 8081

log:
  level: "info"
  file: "/var/log/vpanel/agent.log"
EOF

chmod 600 /etc/vpanel/agent.yaml
echo -e "${GREEN}✓ 配置文件创建完成${NC}"

# 创建 systemd 服务
echo "正在创建 systemd 服务..."
cat > /etc/systemd/system/vpanel-agent.service <<EOF
[Unit]
Description=V Panel Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/vpanel-agent -config /etc/vpanel/agent.yaml
Restart=on-failure
RestartSec=5s
LimitNOFILE=65536

StandardOutput=journal
StandardError=journal
SyslogIdentifier=vpanel-agent

[Install]
WantedBy=multi-user.target
EOF

# 重载 systemd
systemctl daemon-reload
echo -e "${GREEN}✓ systemd 服务创建完成${NC}"

# 启动服务
echo "正在启动 Agent 服务..."
systemctl enable vpanel-agent
systemctl start vpanel-agent

# 等待服务启动
sleep 2

# 检查服务状态
if systemctl is-active --quiet vpanel-agent; then
    echo -e "${GREEN}✓ Agent 服务启动成功${NC}"
else
    echo -e "${RED}✗ Agent 服务启动失败${NC}"
    echo "查看日志: journalctl -u vpanel-agent -n 50"
    exit 1
fi

echo ""
echo "=========================================="
echo "  部署完成！"
echo "=========================================="
echo ""
echo "常用命令:"
echo "  查看状态: systemctl status vpanel-agent"
echo "  查看日志: journalctl -u vpanel-agent -f"
echo "  重启服务: systemctl restart vpanel-agent"
echo "  停止服务: systemctl stop vpanel-agent"
echo ""
`, panelURL, nodeToken)
}

// TestConnection tests SSH connection without deploying.
func (s *RemoteDeployService) TestConnection(ctx context.Context, config *DeployConfig) error {
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// 使用 channel 来处理超时
	type result struct {
		err error
	}
	resultChan := make(chan result, 1)
	
	go func() {
		// 在 goroutine 内部检查 context
		select {
		case <-ctx.Done():
			return
		default:
		}
		
		client, err := s.connectSSH(config)
		if err != nil {
			select {
			case resultChan <- result{err: err}:
			case <-ctx.Done():
			}
			return
		}
		defer client.Close()

		// 再次检查 context
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Try to execute a simple command
		session, err := client.NewSession()
		if err != nil {
			select {
			case resultChan <- result{err: fmt.Errorf("failed to create session: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
		defer session.Close()

		if err := session.Run("echo 'Connection test successful'"); err != nil {
			select {
			case resultChan <- result{err: fmt.Errorf("failed to execute test command: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
		
		select {
		case resultChan <- result{err: nil}:
		case <-ctx.Done():
		}
	}()
	
	select {
	case res := <-resultChan:
		return res.err
	case <-ctx.Done():
		return fmt.Errorf("连接测试超时")
	}
}

// readAgentBinary 读取本地 Agent 二进制文件
func (s *RemoteDeployService) readAgentBinary(path string) ([]byte, error) {
	// 尝试多个可能的路径
	paths := []string{
		path,
		"./bin/vpanel-agent",
		"./bin/vpanel-agent-amd64",
		"./bin/vpanel-agent-arm64",
		"./vpanel-agent",
		"../bin/vpanel-agent",
		"../bin/vpanel-agent-amd64",
		"/usr/local/bin/vpanel-agent",
	}

	for _, p := range paths {
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err == nil {
			s.logger.Info("找到 Agent 文件", logger.F("path", p), logger.F("size", len(data)))
			return data, nil
		}
	}

	return nil, fmt.Errorf("未找到 Agent 二进制文件，请先编译: cd cmd/agent && go build -o ../../bin/vpanel-agent")
}

// base64Encode 对数据进行 base64 编码
func base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
