# Puppeteer E2E 自动化测试文档

## 概述

已为 V Panel 项目配置完整的 Puppeteer E2E（端到端）自动化测试系统。

## 安装完成

✅ **已安装依赖**:
- puppeteer (浏览器自动化)
- jest (测试框架)

✅ **已创建测试框架**:
- 配置文件
- 辅助函数库
- 5 个测试套件
- 完整文档

## 快速开始

### 1. 确保应用运行

```bash
# 使用菜单脚本
./vpanel.sh

# 或使用启动脚本
./scripts/start.sh start
```

### 2. 运行快速验证

```bash
cd web
node tests/e2e/quick-test.js
```

### 3. 运行测试

```bash
# 基础检查（推荐首次运行）
npm run test:e2e tests/basic-check.test.js

# 所有测试
npm run test:e2e

# 有头模式（可以看到浏览器操作）
npm run test:e2e:headed
```

## 测试套件

| 测试套件 | 功能 | 文件 |
|---------|------|------|
| 基础检查 | 验证应用基本功能 | `basic-check.test.js` |
| 管理员登录 | 测试管理后台登录流程 | `admin-login.test.js` |
| 用户门户 | 测试用户门户功能 | `user-portal.test.js` |
| 节点管理 | 测试节点管理功能 | `node-management.test.js` |
| 订阅系统 | 测试订阅链接功能 | `subscription.test.js` |

## 测试覆盖

### ✅ 已实现的测试

**基础功能**:
- 首页访问
- 健康检查端点
- 页面导航

**管理后台**:
- 管理员登录
- 仪表板访问
- 用户管理页面
- 节点管理页面
- 订阅管理页面

**用户门户**:
- 用户门户首页
- 注册页面
- 登录页面

**节点管理**:
- 节点列表显示
- 添加节点对话框

**订阅系统**:
- 订阅管理页面
- 订阅链接显示

## 文件结构

```
web/tests/e2e/
├── puppeteer.config.js       # Puppeteer 配置
├── jest.config.js            # Jest 配置
├── run-tests.js              # 测试运行器
├── quick-test.js             # 快速验证脚本
├── .env.example              # 环境变量示例
├── helpers/
│   ├── browser.js            # 浏览器操作封装
│   └── auth.js               # 认证辅助函数
└── tests/
    ├── basic-check.test.js
    ├── admin-login.test.js
    ├── user-portal.test.js
    ├── node-management.test.js
    └── subscription.test.js
```

## 配置说明

### 环境变量

在 `web/tests/e2e/.env` 中配置（可选）：

```bash
BASE_URL=http://localhost:8080
ADMIN_USER=admin
ADMIN_PASS=admin123
HEADLESS=true
SCREENSHOT=true
```

### 运行模式

```bash
# 无头模式（后台运行，默认）
npm run test:e2e

# 有头模式（显示浏览器）
npm run test:e2e:headed

# 慢速模式（便于观察）
SLOW_MO=500 npm run test:e2e:headed

# 开发者工具模式
DEVTOOLS=true npm run test:e2e:headed
```

## 辅助功能

### BrowserHelper (浏览器操作)

```javascript
await browser.goto('/path')              // 导航
await browser.click('button')            // 点击
await browser.type('input', 'text')      // 输入
await browser.getText('.element')        // 获取文本
await browser.exists('.element')         // 检查元素
await browser.screenshot('name')         // 截图
await browser.waitForElement('.element') // 等待元素
```

### AuthHelper (认证操作)

```javascript
await auth.loginAsAdmin()                // 管理员登录
await auth.loginAsUser()                 // 用户登录
await auth.logout()                      // 登出
await auth.isLoggedIn()                  // 检查登录状态
```

## 截图功能

所有测试自动截图，保存在：
```
web/tests/e2e/screenshots/
```

截图时机：
- 页面加载后
- 登录前后
- 关键操作前后
- 测试失败时

## 调试技巧

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

## 编写自定义测试

基本模板：

```javascript
import { BrowserHelper } from '../helpers/browser.js';
import { AuthHelper } from '../helpers/auth.js';

describe('我的测试', () => {
  let browser;
  let auth;

  beforeAll(async () => {
    browser = new BrowserHelper();
    await browser.launch();
    auth = new AuthHelper(browser);
  });

  afterAll(async () => {
    await browser.close();
  });

  test('测试用例', async () => {
    // 登录（如果需要）
    await auth.loginAsAdmin();
    
    // 导航到页面
    await browser.goto('/admin/dashboard');
    
    // 执行操作
    await browser.click('.button');
    await browser.type('input', 'value');
    
    // 截图
    await browser.screenshot('test-step');
    
    // 断言
    const text = await browser.getText('.result');
    expect(text).toBe('Expected');
  });
});
```

## 常见问题

### 1. 应用未运行

**错误**: 测试超时或连接失败

**解决**:
```bash
./vpanel.sh  # 启动应用
```

### 2. 端口冲突

**错误**: 无法访问 localhost:8080

**解决**:
```bash
# 修改配置
export BASE_URL=http://localhost:YOUR_PORT
npm run test:e2e
```

### 3. 登录失败

**错误**: 登录测试失败

**解决**:
- 检查管理员凭证是否正确
- 查看截图了解失败原因
- 使用有头模式观察登录过程

### 4. 元素找不到

**错误**: Element not found

**解决**:
- 使用有头模式查看页面
- 检查选择器是否正确
- 增加等待时间

### 5. 浏览器启动失败

**错误**: Failed to launch browser

**解决**:
```bash
# macOS
brew install chromium

# Linux
sudo apt-get install chromium-browser
```

## CI/CD 集成

### GitHub Actions

```yaml
- name: Run E2E Tests
  run: |
    cd web
    npm run test:e2e
  env:
    BASE_URL: http://localhost:8080
    HEADLESS: true
```

### GitLab CI

```yaml
e2e-tests:
  script:
    - cd web
    - npm install
    - npm run test:e2e
  variables:
    BASE_URL: http://localhost:8080
    HEADLESS: "true"
```

## 性能优化

1. **并行测试**: 修改 `jest.config.js` 中的 `maxWorkers`
2. **选择性测试**: 只运行相关测试套件
3. **缓存依赖**: CI/CD 中缓存 node_modules
4. **减少截图**: 生产环境禁用截图

## 最佳实践

1. ✅ 先运行基础检查确保应用正常
2. ✅ 使用有头模式调试失败的测试
3. ✅ 查看截图了解测试失败原因
4. ✅ 保持测试独立，不依赖其他测试
5. ✅ 使用有意义的选择器（data-testid）
6. ✅ 适当使用等待，避免固定延迟
7. ✅ 在关键步骤截图

## 扩展测试

可以添加更多测试：

- 代理管理测试
- 流量统计测试
- 证书管理测试
- 系统设置测试
- 用户权限测试
- 支付流程测试
- 订单管理测试

参考现有测试文件编写新的测试用例。

## 相关文档

- **快速指南**: `PUPPETEER_GUIDE.md`
- **详细文档**: `web/tests/e2e/README.md`
- **设置完成**: `web/tests/e2e/SETUP_COMPLETE.md`
- **配置文件**: `web/tests/e2e/puppeteer.config.js`

## 支持

遇到问题时：

1. 查看测试截图
2. 查看应用日志 (`logs/app.log`)
3. 使用有头模式观察
4. 查看浏览器控制台（DEVTOOLS=true）
5. 参考详细文档

---

**配置完成**: 2026-01-17
**测试框架**: Puppeteer + Jest
**测试套件**: 5 个
**辅助函数**: 2 个
