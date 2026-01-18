# 用户门户与管理后台访问优化 - 更新说明

## 更新日期
2026-01-17

## 更新内容

### 1. 统一登录入口
- ✅ 移除了独立的管理员登录页面 `/login` 和 `/register`
- ✅ 所有用户（包括管理员）统一使用用户门户登录页 `/user/login`
- ✅ 根据用户角色自动跳转到对应的页面

### 2. 智能路由跳转
- ✅ 访问根路径 `/` 时：
  - 未登录 → 跳转到 `/user/login`
  - 普通用户 → 跳转到 `/user/dashboard`
  - 管理员 → 跳转到 `/admin/dashboard`
- ✅ 访问旧的 `/login` 或 `/register` 自动重定向到 `/user/login`

### 3. 管理后台访问控制
- ✅ 未登录访问管理后台 → 跳转到用户登录页
- ✅ 普通用户访问管理后台 → 跳转到用户门户
- ✅ 管理员可以正常访问管理后台

### 4. 登录逻辑优化
- ✅ 用户在门户登录后，系统自动识别用户角色
- ✅ 管理员登录后自动跳转到 `/admin/dashboard`
- ✅ 普通用户登录后跳转到 `/user/dashboard`
- ✅ 支持 2FA 双因素认证

### 5. 登出逻辑优化
- ✅ 所有登出操作统一跳转到 `/user/login`
- ✅ 清除所有认证相关的 localStorage 数据

## 修改的文件

### 路由配置
- `web/src/router/index.js` - 主路由配置
  - 移除了 `/login` 和 `/register` 路由
  - 优化了根路径跳转逻辑
  - 增强了路由守卫的访问控制

### Store
- `web/src/stores/userPortal.js` - 用户门户 Store
  - 登录时自动识别管理员角色
  - 为管理员设置额外的 token 和角色标识

### 视图组件
- `web/src/views/user/Login.vue` - 用户登录页
  - 根据用户角色智能跳转
  - 支持管理员登录后跳转到管理后台

### API 配置
- `web/src/api/base.js` - API 基础配置
  - 401 错误统一跳转到用户登录页

### 布局组件
- `web/src/layouts/MainLayout.vue` - 管理后台布局
- `web/src/layouts/DefaultLayout.vue` - 默认布局
- `web/src/views/Home.vue` - 首页
- `web/src/views/Register.vue` - 注册页
  - 登出操作统一跳转到用户登录页

## 使用说明

### 普通用户
1. 访问 `http://localhost:8080/` 或 `http://localhost:8080/user/login`
2. 输入用户名和密码登录
3. 登录成功后自动跳转到用户门户仪表板

### 管理员
1. 访问 `http://localhost:8080/` 或 `http://localhost:8080/user/login`
2. 输入管理员账号和密码登录
3. 登录成功后自动跳转到管理后台仪表板
4. 管理员可以访问所有管理功能

### 安全特性
- ✅ 管理后台地址被隐藏，无法直接访问 `/login`
- ✅ 未登录用户无法访问管理后台
- ✅ 普通用户无法访问管理后台
- ✅ 所有访问控制在前端路由守卫中实现
- ⚠️ 建议在后端 API 中也实现相应的权限验证

## 测试建议

请参考 `Docs/portal-admin-integration-test.md` 进行完整的功能测试。

主要测试点：
1. 根路径访问跳转是否正确
2. 管理员登录后是否跳转到管理后台
3. 普通用户是否无法访问管理后台
4. 旧的 `/login` 路径是否正确重定向
5. 登出后是否正确清除认证信息

## 后续优化建议

1. **后端权限验证**：在后端 API 中实现基于角色的访问控制（RBAC）
2. **IP 白名单**：为管理后台添加 IP 白名单限制
3. **审计日志**：记录管理员的所有操作
4. **会话管理**：实现更安全的会话管理机制
5. **双因素认证**：为管理员账号强制启用 2FA
6. **专用管理员入口**：可以考虑添加隐藏的管理员登录入口（如 `/admin/login`），但不在前端暴露链接

## 回滚方案

如果需要回滚到之前的版本：
1. 恢复 `web/src/router/index.js` 中的 `/login` 和 `/register` 路由
2. 恢复各个文件中的 `router.push('/login')` 调用
3. 恢复 `web/src/stores/userPortal.js` 中的登录逻辑

## 联系方式

如有问题或建议，请联系开发团队。
