# 已知问题和限制

## 当前版本的限制

### 1. Agent 二进制分发 ⚠️

**问题**: 远程部署功能目前不包含 Agent 二进制的自动下载

**原因**: Agent 二进制需要预先编译并上传到服务器或提供下载地址

**临时解决方案**:

#### 方案 A: 手动上传 Agent 二进制

```bash
# 1. 编译 Agent
go build -o vpanel-agent ./cmd/agent/main.go

# 2. 上传到目标服务器
scp vpanel-agent root@node-server:/usr/local/bin/

# 3. 设置权限
ssh root@node-server "chmod +x /usr/local/bin/vpanel-agent"

# 4. 然后使用远程部署配置和启动服务
```

#### 方案 B: 使用部署脚本

```bash
# 1. 下载部署脚本
curl -H "Authorization: Bearer <token>" \
  https://panel.example.com/api/admin/nodes/1/deploy/script \
  -o install-agent.sh

# 2. 手动编辑脚本，添加 Agent 下载地址
# 在脚本中找到这一行：
# # wget -O /usr/local/bin/vpanel-agent https://your-panel.com/downloads/vpanel-agent

# 3. 替换为实际的下载地址或手动上传

# 4. 执行脚本
bash install-agent.sh
```

#### 方案 C: 设置下载服务器

```bash
# 1. 在 Panel 服务器上提供 Agent 下载
mkdir -p /var/www/downloads
cp vpanel-agent /var/www/downloads/
chmod 644 /var/www/downloads/vpanel-agent

# 2. 配置 nginx 或其他 web 服务器提供下载
# 例如在 nginx 配置中：
location /downloads {
    alias /var/www/downloads;
    autoindex on;
}

# 3. 修改部署脚本使用这个地址
wget -O /usr/local/bin/vpanel-agent https://panel.example.com/downloads/vpanel-agent
```

**计划改进**:
- [ ] 在 Panel 中内置 Agent 二进制分发
- [ ] 支持从 GitHub Releases 自动下载
- [ ] 支持多架构二进制（amd64, arm64）

### 2. 并发部署限制 ⚠️

**问题**: 目前不支持同时部署多个节点

**影响**: 部署多个节点时需要逐个执行

**临时解决方案**: 依次部署每个节点

**计划改进**:
- [ ] 实现异步部署队列
- [ ] 支持批量部署
- [ ] 显示部署进度

### 3. Windows 节点支持 ⚠️

**问题**: 目前只支持 Linux 和 macOS 节点

**原因**: 
- Xray 安装脚本主要针对 Linux
- systemd 服务管理不适用于 Windows

**临时解决方案**: 在 Windows 上手动安装

**计划改进**:
- [ ] 支持 Windows 节点
- [ ] 使用 Windows Service 管理

### 4. 部署回滚 ⚠️

**问题**: 部署失败后无法自动回滚

**影响**: 需要手动清理或重新部署

**临时解决方案**: 
```bash
# 手动清理
systemctl stop vpanel-agent
rm -rf /etc/vpanel
rm /usr/local/bin/vpanel-agent
```

**计划改进**:
- [ ] 实现部署快照
- [ ] 支持一键回滚

### 5. SSH 密钥格式 ⚠️

**问题**: 只支持 RSA 私钥，不支持新格式（如 ed25519）

**临时解决方案**: 使用 RSA 密钥或密码认证

```bash
# 生成 RSA 密钥
ssh-keygen -t rsa -b 4096
```

**计划改进**:
- [ ] 支持所有 SSH 密钥格式
- [ ] 支持 SSH Agent 转发

## 配置相关问题

### 1. 代理端口冲突检测

**问题**: 只检查数据库中的端口冲突，不检查实际端口占用

**影响**: 可能配置已被其他程序占用的端口

**临时解决方案**: 
```bash
# 在节点上检查端口
netstat -tlnp | grep <port>
```

**计划改进**:
- [ ] Agent 报告实际端口占用情况
- [ ] Panel 显示端口使用状态

### 2. 配置同步延迟

**问题**: 配置更新后需要等待最多 5 分钟才能同步

**影响**: 新配置不能立即生效

**临时解决方案**:
```bash
# 手动触发同步（在节点上）
systemctl restart vpanel-agent
```

**计划改进**:
- [ ] 实现配置推送（WebSocket）
- [ ] 提供手动同步按钮

### 3. TLS 证书管理

**问题**: 需要手动管理 TLS 证书

**影响**: 证书过期需要手动更新

**临时解决方案**: 使用 certbot 自动续期

```bash
# 安装 certbot
apt install certbot

# 获取证书
certbot certonly --standalone -d example.com

# 自动续期
certbot renew --dry-run
```

**计划改进**:
- [ ] 集成 Let's Encrypt
- [ ] 自动证书续期
- [ ] 证书到期提醒

## 性能相关问题

### 1. 大量代理配置

**问题**: 单个节点配置大量代理（>100）时，配置文件可能很大

**影响**: 配置同步和 Xray 重启时间较长

**临时解决方案**: 
- 将代理分散到多个节点
- 只启用必要的代理

**计划改进**:
- [ ] 配置压缩传输
- [ ] 增量配置更新
- [ ] Xray 热重载

### 2. 数据库查询优化

**问题**: 获取节点配置时需要查询所有代理

**影响**: 节点较多时可能影响性能

**临时解决方案**: 已添加索引优化

**计划改进**:
- [ ] 配置缓存
- [ ] 查询结果缓存

## 安全相关问题

### 1. SSH 密码存储

**问题**: 远程部署时 SSH 密码在内存中明文传输

**影响**: 存在安全风险

**建议**: 
- 使用 SSH 密钥认证
- 部署后立即修改密码
- 使用临时密码

**计划改进**:
- [ ] 密码加密存储
- [ ] 支持 SSH Agent
- [ ] 一次性部署令牌

### 2. Node Token 安全

**问题**: Node Token 一旦泄露，可能被用于伪造节点

**影响**: 安全风险

**建议**:
- 定期轮换 Token
- 限制 Token 使用 IP
- 监控异常连接

**计划改进**:
- [ ] Token 自动轮换
- [ ] IP 白名单
- [ ] 异常检测告警

## 监控和日志

### 1. 部署日志保存

**问题**: 部署日志只在部署时显示，不持久化

**影响**: 无法回溯查看历史部署日志

**临时解决方案**: 复制保存部署日志

**计划改进**:
- [ ] 部署日志持久化
- [ ] 部署历史记录
- [ ] 部署失败告警

### 2. Agent 监控

**问题**: 缺少 Agent 详细监控指标

**影响**: 难以诊断问题

**临时解决方案**: 查看系统日志

```bash
journalctl -u vpanel-agent -f
```

**计划改进**:
- [ ] Agent 监控面板
- [ ] 性能指标收集
- [ ] 告警系统

## 文档和工具

### 1. 缺少 Agent 命令行工具

**问题**: Agent 缺少管理命令

**影响**: 需要手动操作

**计划改进**:
- [ ] Agent CLI 工具
- [ ] 配置验证命令
- [ ] 状态查询命令

### 2. 缺少故障诊断工具

**问题**: 缺少自动化故障诊断

**影响**: 需要手动排查

**计划改进**:
- [ ] 健康检查脚本
- [ ] 自动诊断工具
- [ ] 问题修复建议

## 报告问题

如果你发现新的问题，请：

1. 查看日志
   ```bash
   # Panel 日志
   tail -f logs/vpanel.log
   
   # Agent 日志
   journalctl -u vpanel-agent -f
   
   # Xray 日志
   tail -f /var/log/xray/error.log
   ```

2. 收集信息
   - 操作系统版本
   - 错误信息
   - 复现步骤

3. 提交 Issue 或联系支持

## 更新日志

- 2026-01-19: 初始版本，记录已知问题
