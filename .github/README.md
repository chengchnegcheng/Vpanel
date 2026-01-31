# GitHub 配置说明

本目录包含 V Panel 项目的 GitHub 配置文件。

## 目录结构

```
.github/
├── workflows/              # GitHub Actions 工作流
│   ├── build.yml          # 构建和发布
│   ├── test.yml           # 测试
│   ├── security.yml       # 安全扫描
│   ├── stale.yml          # 自动关闭过期 Issue/PR
│   └── label.yml          # 自动标签
├── ISSUE_TEMPLATE/        # Issue 模板
│   ├── bug_report.yml     # Bug 报告
│   └── feature_request.yml # 功能请求
├── PULL_REQUEST_TEMPLATE.md # PR 模板
├── CODE_OF_CONDUCT.md     # 行为准则
├── CONTRIBUTING.md        # 贡献指南
├── FUNDING.yml            # 赞助配置
├── dependabot.yml         # 依赖自动更新
└── labeler.yml            # 自动标签规则
```

## 工作流说明

### build.yml - 构建和发布

**触发条件:**
- Push 到 main/develop 分支
- 创建 tag (v*)
- Pull Request 到 main/develop

**功能:**
- 代码质量检查 (golangci-lint)
- 后端测试 (Go)
- 前端测试 (npm)
- 多平台构建 (Linux, Windows, macOS)
- Docker 镜像构建
- 自动发布 Release

### test.yml - 测试

**触发条件:**
- Push 到 main/develop 分支
- Pull Request 到 main/develop

**功能:**
- 单元测试
- 前端测试
- E2E 测试
- 代码覆盖率上传

### security.yml - 安全扫描

**触发条件:**
- Push 到 main/develop 分支
- Pull Request 到 main/develop
- 每周日定时运行

**功能:**
- Go 安全扫描 (gosec)
- 依赖漏洞扫描 (Trivy)
- CodeQL 分析

### stale.yml - 自动关闭过期 Issue/PR

**触发条件:**
- 每天定时运行

**功能:**
- Issue 60 天无活动标记为 stale
- PR 30 天无活动标记为 stale
- 7 天后自动关闭

### label.yml - 自动标签

**触发条件:**
- Issue/PR 创建或更新

**功能:**
- 根据文件变更自动添加标签
- 标签规则在 labeler.yml 中配置

## 配置 Secrets

在 GitHub 仓库设置中添加以下 Secrets：

### 必需的 Secrets

| Secret | 说明 | 用途 |
|--------|------|------|
| `GITHUB_TOKEN` | GitHub 自动提供 | Release、上传 artifact |
| `DOCKER_USERNAME` | Docker Hub 用户名 | 推送 Docker 镜像 |
| `DOCKER_PASSWORD` | Docker Hub 密码/Token | 推送 Docker 镜像 |
| `CODECOV_TOKEN` | Codecov Token | 上传代码覆盖率 |

### 可选的 Secrets

| Secret | 说明 | 用途 |
|--------|------|------|
| `SLACK_WEBHOOK` | Slack Webhook URL | 构建通知 |
| `TELEGRAM_TOKEN` | Telegram Bot Token | 构建通知 |

## 配置步骤

### 1. 启用 GitHub Actions

1. 进入仓库 Settings
2. 点击 Actions > General
3. 选择 "Allow all actions and reusable workflows"
4. 保存设置

### 2. 配置 Docker Hub

1. 登录 [Docker Hub](https://hub.docker.com/)
2. 创建 Access Token
3. 在 GitHub 仓库添加 Secrets:
   - `DOCKER_USERNAME`: 你的 Docker Hub 用户名
   - `DOCKER_PASSWORD`: 创建的 Access Token

### 3. 配置 Codecov

1. 登录 [Codecov](https://codecov.io/)
2. 添加你的 GitHub 仓库
3. 获取 Upload Token
4. 在 GitHub 仓库添加 Secret:
   - `CODECOV_TOKEN`: Codecov Upload Token

### 4. 配置 Dependabot

Dependabot 配置已包含在 `dependabot.yml` 中，会自动：
- 每周检查 Go 依赖更新
- 每周检查 npm 依赖更新
- 每周检查 Docker 基础镜像更新
- 每周检查 GitHub Actions 更新

### 5. 配置标签

标签规则在 `labeler.yml` 中定义，会根据文件变更自动添加标签：
- `backend` - Go 代码变更
- `frontend` - Vue/JS 代码变更
- `documentation` - 文档变更
- `configuration` - 配置文件变更
- `deployment` - 部署文件变更
- `tests` - 测试文件变更
- `ci/cd` - CI/CD 配置变更
- `database` - 数据库相关变更
- `security` - 安全相关变更
- `api` - API 相关变更
- `dependencies` - 依赖更新

## 发布流程

### 自动发布

1. 创建并推送 tag:
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

2. GitHub Actions 会自动:
   - 运行所有测试
   - 构建多平台二进制文件
   - 构建 Docker 镜像
   - 创建 GitHub Release
   - 上传构建产物

### 手动发布

1. 进入 Actions 页面
2. 选择 "Build and Release" 工作流
3. 点击 "Run workflow"
4. 选择分支或 tag
5. 点击 "Run workflow"

## 徽章说明

在 README.md 中添加的徽章：

- **Build Status**: 显示最新构建状态
- **Go Report Card**: Go 代码质量评分
- **codecov**: 代码覆盖率
- **License**: 项目许可证
- **GitHub release**: 最新版本
- **Docker Pulls**: Docker 镜像下载次数

## 贡献指南

详细的贡献指南请查看 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 问题反馈

- 使用 [Bug Report](https://github.com/chengchnegcheng/V/issues/new?template=bug_report.yml) 报告 Bug
- 使用 [Feature Request](https://github.com/chengchnegcheng/V/issues/new?template=feature_request.yml) 提出新功能

## 许可证

MIT License - 详见 [LICENSE](../LICENSE)
