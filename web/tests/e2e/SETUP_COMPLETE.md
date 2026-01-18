# ✅ Puppeteer E2E 测试设置完成

## 📦 已安装的依赖

- ✅ puppeteer - 浏览器自动化
- ✅ jest - 测试框架

## 📁 创建的文件结构

```
web/tests/e2e/
├── puppeteer.config.js       # Puppeteer 配置
├── jest.config.js            # Jest 配置
├── run-tests.js              # 测试运行器
├── quick-test.js             # 快速验证脚本
├── .env.example              # 环境变量示例
├── .gitignore                # Git 忽略文件
├── README.md                 # 详细文档
├── helpers/                  # 辅助函数
│   ├── browser.js            # 浏览器操作封装
│   └── auth.js               # 认证辅助
└── tests/                    # 测试用例
    ├── basic-check.test.js   # 基础检查
    ├── admin-login.test.js   # 管理员登录
    ├── user-portal.test.js   # 用户门户
    ├── node-management.test.js # 节点管理
    └── subscription.test.js  # 订阅系统
```

## 🚀 快速开始

### 1. 验证设置

```bash
cd web
node tests/e2e/quick-test.js
```

### 2. 运行基础检查

```bash
npm run test:e2e tests/basic-check.test.js
```

### 3. 运行所有测试

```bash
# 无头模式
npm run test:e2e

# 有头模式（可以看到浏览器）
npm run test:e2e:headed
```

## 📋 可用的测试套件

| 测试文件 | 描述 | 运行命令 |
|---------|------|---------|
| `basic-check.test.js` | 基础功能检查 | `npm run test:e2e tests/basic-check.test.js` |
| `admin-login.test.js` | 管理员登录流程 | `npm run test:e2e tests/admin-login.test.js` |
| `user-portal.test.js` | 用户门户测试 | `npm run test:e2e tests/user-portal.test.js` |
| `node-management.test.js` | 节点管理测试 | `npm run test:e2e tests/node-management.test.js` |
| `subscription.test.js` | 订阅系统测试 | `npm run test:e2e tests/subscription.test.js` |

## 🎯 测试功能

### 基础功能
- ✅ 页面导航和加载
- ✅ 元素查找和交互
- ✅ 表单填写和提交
- ✅ 自动截图
- ✅ 等待和超时处理

### 认证功能
- ✅ 管理员登录
- ✅ 用户登录
- ✅ 登出
- ✅ 登录状态检查

### 页面测试
- ✅ 首页访问
- ✅ 管理后台
- ✅ 用户门户
- ✅ 节点管理
- ✅ 订阅系统

## 🔧 配置选项

### 环境变量

```bash
# 应用 URL
BASE_URL=http://localhost:8080

# 管理员凭证
ADMIN_USER=admin
ADMIN_PASS=admin123

# 浏览器模式
HEADLESS=false    # 显示浏览器
SLOW_MO=500       # 减慢操作（毫秒）
DEVTOOLS=true     # 打开开发者工具

# 截图
SCREENSHOT=true   # 启用截图
```

### 运行模式

```bash
# 无头模式（默认）
npm run test:e2e

# 有头模式
npm run test:e2e:headed

# 慢速模式（便于观察）
SLOW_MO=500 npm run test:e2e:headed

# 开发者工具模式
DEVTOOLS=true npm run test:e2e:headed
```

## 📸 截图

所有测试运行时会自动截图，保存在：
```
web/tests/e2e/screenshots/
```

截图命名格式：`{测试名称}-{时间戳}.png`

## 🐛 调试技巧

### 1. 查看浏览器操作
```bash
npm run test:e2e:headed tests/basic-check.test.js
```

### 2. 减慢操作速度
```bash
SLOW_MO=1000 npm run test:e2e:headed
```

### 3. 打开开发者工具
```bash
DEVTOOLS=true npm run test:e2e:headed
```

### 4. 查看截图
```bash
open web/tests/e2e/screenshots/
```

### 5. 运行单个测试
```bash
npm run test:e2e tests/admin-login.test.js
```

## 📚 文档

- **快速指南**: `PUPPETEER_GUIDE.md`（项目根目录）
- **详细文档**: `web/tests/e2e/README.md`
- **配置说明**: `web/tests/e2e/puppeteer.config.js`

## ⚠️ 注意事项

1. **确保应用运行**: 测试前确保 V Panel 应用正在运行
2. **端口配置**: 默认使用 `http://localhost:8080`
3. **凭证配置**: 默认管理员账号 `admin/admin123`
4. **超时设置**: 默认导航超时 30 秒
5. **截图目录**: 自动创建，无需手动创建

## 🔍 常见问题

### 应用未运行
```bash
# 启动应用
./vpanel.sh
# 或
./scripts/start.sh start
```

### 测试超时
- 检查应用是否正常运行
- 检查 BASE_URL 配置
- 增加超时时间

### 元素找不到
- 使用有头模式查看页面
- 检查选择器是否正确
- 查看截图了解页面状态

### 浏览器启动失败
```bash
# macOS
brew install chromium

# Linux
sudo apt-get install chromium-browser
```

## 🎉 下一步

1. **运行快速测试**: `node tests/e2e/quick-test.js`
2. **运行基础检查**: `npm run test:e2e tests/basic-check.test.js`
3. **运行所有测试**: `npm run test:e2e`
4. **编写自定义测试**: 参考 `web/tests/e2e/README.md`

## 📞 获取帮助

- 查看详细文档: `web/tests/e2e/README.md`
- 查看快速指南: `PUPPETEER_GUIDE.md`
- 查看测试截图: `web/tests/e2e/screenshots/`
- 查看应用日志: `logs/app.log`

---

**设置完成时间**: $(date)
**Puppeteer 版本**: 最新版本
**Jest 版本**: 最新版本
