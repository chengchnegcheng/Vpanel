# 贡献指南

感谢你考虑为 V Panel 做出贡献！

## 行为准则

请阅读并遵守我们的[行为准则](CODE_OF_CONDUCT.md)。

## 如何贡献

### 报告 Bug

1. 检查 [Issues](https://github.com/chengchnegcheng/V/issues) 确保 Bug 未被报告
2. 使用 Bug 报告模板创建新 Issue
3. 提供详细的复现步骤和环境信息
4. 如果可能，提供最小可复现示例

### 提出新功能

1. 检查 [Issues](https://github.com/chengchnegcheng/V/issues) 确保功能未被提出
2. 使用功能请求模板创建新 Issue
3. 清楚地描述功能和使用场景
4. 等待维护者反馈后再开始实现

### 提交代码

#### 开发流程

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交变更 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

#### 代码规范

**Go 代码**
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 使用 `gofmt` 格式化代码
- 通过 `golangci-lint` 检查
- 添加必要的注释和文档

**JavaScript/Vue 代码**
- 遵循 [Vue.js 风格指南](https://vuejs.org/style-guide/)
- 使用 ESLint 和 Prettier
- 组件命名使用 PascalCase
- 文件名使用 kebab-case

#### 提交信息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型 (type)**
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具相关
- `perf`: 性能优化
- `ci`: CI/CD 相关

**示例**
```
feat(auth): add JWT token refresh

- Implement token refresh endpoint
- Add refresh token to database
- Update frontend to use refresh token

Closes #123
```

#### 测试要求

- 所有新功能必须包含测试
- Bug 修复应包含回归测试
- 测试覆盖率应达到 80% 以上
- 确保所有测试通过

```bash
# 运行后端测试
go test -v ./...

# 运行前端测试
cd web
npm run test:unit

# 运行 E2E 测试
npm run test:e2e
```

#### Pull Request 检查清单

- [ ] 代码遵循项目规范
- [ ] 添加了必要的测试
- [ ] 所有测试通过
- [ ] 更新了相关文档
- [ ] 提交信息符合规范
- [ ] 没有合并冲突
- [ ] CI/CD 检查通过

### 文档贡献

文档同样重要！你可以：

- 修复文档中的错误
- 改进现有文档
- 添加示例和教程
- 翻译文档

## 开发环境设置

### 后端开发

```bash
# 安装 Go 1.23+
# 克隆仓库
git clone https://github.com/chengchnegcheng/V.git
cd V

# 安装依赖
go mod download

# 运行开发服务器
go run ./cmd/v/main.go -config configs/config.yaml.example
```

### 前端开发

```bash
# 安装 Node.js 20+
cd web

# 安装依赖
npm install

# 运行开发服务器
npm run dev
```

### 使用 Docker

```bash
# 构建镜像
docker build -t vpanel:dev -f deployments/docker/Dockerfile .

# 运行容器
docker run -p 8080:8080 vpanel:dev
```

## 项目结构

```
v/
├── cmd/                    # 主程序入口
├── internal/               # 私有包
│   ├── api/               # API 层
│   ├── database/          # 数据库层
│   └── ...
├── pkg/                    # 公共包
├── web/                    # 前端代码
├── configs/               # 配置文件
├── deployments/           # 部署文件
└── scripts/               # 脚本
```

## 获取帮助

- 查看 [文档](../Docs/)
- 在 [Discussions](https://github.com/chengchnegcheng/V/discussions) 提问
- 加入我们的社区

## 许可证

通过贡献代码，你同意你的贡献将在 MIT 许可证下发布。
