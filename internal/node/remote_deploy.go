// Package node provides node management functionality.
package node

import (
	"bytes"
	"context"
	"fmt"
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
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key,omitempty"`
	PanelURL   string `json:"panel_url"`
	NodeToken  string `json:"node_token"`
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
	}

	// Private key authentication
	if config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided")
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Use proper host key verification
		Timeout:         30 * time.Second,
	}

	port := config.Port
	if port == 0 {
		port = 22
	}

	addr := fmt.Sprintf("%s:%d", config.Host, port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return client, nil
}

// executeCommand executes a command on the remote server.
func (s *RemoteDeployService) executeCommand(client *ssh.Client, command string, logBuffer *bytes.Buffer) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Capture output
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	s.logger.Debug("Executing command", logger.F("command", command))

	if err := session.Run(command); err != nil {
		logBuffer.WriteString(fmt.Sprintf("✗ 命令失败: %s\n", command))
		logBuffer.WriteString(fmt.Sprintf("错误: %v\n", err))
		if stderr.Len() > 0 {
			logBuffer.WriteString(fmt.Sprintf("stderr: %s\n", stderr.String()))
		}
		return fmt.Errorf("command failed: %w", err)
	}

	if stdout.Len() > 0 {
		logBuffer.WriteString(stdout.String())
		logBuffer.WriteString("\n")
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
# Detect package manager
if command -v apt-get &> /dev/null; then
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    apt-get install -y -qq curl wget unzip systemctl
elif command -v yum &> /dev/null; then
    yum install -y -q curl wget unzip systemd
elif command -v dnf &> /dev/null; then
    dnf install -y -q curl wget unzip systemd
else
    echo "Unsupported package manager"
    exit 1
fi
`

	if err := s.executeCommand(client, script, logBuffer); err != nil {
		return err
	}

	logBuffer.WriteString("✓ 依赖安装完成\n")
	return nil
}

// installAgent downloads and installs the agent binary.
func (s *RemoteDeployService) installAgent(client *ssh.Client, config *DeployConfig, logBuffer *bytes.Buffer) error {
	// For now, we'll build and upload the agent
	// In production, you'd download from a release URL
	script := `
# Create directories
mkdir -p /usr/local/bin
mkdir -p /etc/vpanel
mkdir -p /var/log/vpanel

# Download agent (placeholder - replace with actual download URL)
# For now, we'll assume the agent binary is uploaded separately
echo "Agent binary should be uploaded to /usr/local/bin/vpanel-agent"
`

	if err := s.executeCommand(client, script, logBuffer); err != nil {
		return err
	}

	logBuffer.WriteString("✓ Agent 目录创建完成\n")
	return nil
}

// installXray installs Xray using the official script.
func (s *RemoteDeployService) installXray(client *ssh.Client, logBuffer *bytes.Buffer) error {
	script := `
# Check if Xray is already installed
if command -v xray &> /dev/null; then
    echo "Xray is already installed"
    xray version
else
    echo "Installing Xray..."
    bash -c "$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)" @ install
fi
`

	if err := s.executeCommand(client, script, logBuffer); err != nil {
		return err
	}

	logBuffer.WriteString("✓ Xray 安装完成\n")
	return nil
}

// configureAgent creates the agent configuration file.
func (s *RemoteDeployService) configureAgent(client *ssh.Client, config *DeployConfig, logBuffer *bytes.Buffer) error {
	// Create agent config
	agentConfig := fmt.Sprintf(`panel:
  url: "%s"
  token: "%s"

xray:
  binary_path: "/usr/local/bin/xray"
  config_path: "/etc/xray/config.json"

sync:
  interval: 5m
  validate_before_apply: true
  backup_before_apply: true

health:
  port: 8081
`, config.PanelURL, config.NodeToken)

	// Write config file
	script := fmt.Sprintf(`cat > /etc/vpanel/agent.yaml <<'EOF'
%s
EOF
chmod 644 /etc/vpanel/agent.yaml
`, agentConfig)

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
sleep 2
systemctl status vpanel-agent --no-pager
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

set -e

PANEL_URL="%s"
NODE_TOKEN="%s"

echo "开始部署 V Panel Agent..."

# 安装依赖
if command -v apt-get &> /dev/null; then
    apt-get update -qq
    apt-get install -y curl wget unzip
elif command -v yum &> /dev/null; then
    yum install -y curl wget unzip
fi

# 安装 Xray
if ! command -v xray &> /dev/null; then
    echo "安装 Xray..."
    bash -c "$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)" @ install
fi

# 创建目录
mkdir -p /etc/vpanel
mkdir -p /var/log/vpanel

# 创建配置文件
cat > /etc/vpanel/agent.yaml <<EOF
panel:
  url: "$PANEL_URL"
  token: "$NODE_TOKEN"

xray:
  binary_path: "/usr/local/bin/xray"
  config_path: "/etc/xray/config.json"

sync:
  interval: 5m

health:
  port: 8081
EOF

# 下载 Agent (需要替换为实际的下载地址)
# wget -O /usr/local/bin/vpanel-agent https://your-panel.com/downloads/vpanel-agent
# chmod +x /usr/local/bin/vpanel-agent

# 创建 systemd 服务
cat > /etc/systemd/system/vpanel-agent.service <<EOF
[Unit]
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
EOF

# 启动服务
systemctl daemon-reload
systemctl enable vpanel-agent
systemctl start vpanel-agent

echo "部署完成！"
echo "查看状态: systemctl status vpanel-agent"
echo "查看日志: journalctl -u vpanel-agent -f"
`, panelURL, nodeToken)
}

// TestConnection tests SSH connection without deploying.
func (s *RemoteDeployService) TestConnection(ctx context.Context, config *DeployConfig) error {
	client, err := s.connectSSH(config)
	if err != nil {
		return err
	}
	defer client.Close()

	// Try to execute a simple command
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	if err := session.Run("echo 'Connection test successful'"); err != nil {
		return fmt.Errorf("failed to execute test command: %w", err)
	}

	return nil
}
