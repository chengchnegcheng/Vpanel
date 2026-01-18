# 路由访问说明

## 功能描述

系统现在默认以**用户门户**为主入口，管理后台地址相对隐藏，提供更好的用户体验和安全性。

## 访问路径

### 用户门户（默认入口）
- **根路径**: `http://localhost:8080/` → 自动跳转到 `/user/login`
- **用户登录**: `http://localhost:8080/user/login`
- **用户注册**: `http://localhost:8080/user/register`
- **用户仪表板**: `http://localhost:8080/user/dashboard`（需要登录）

### 管理后台（隐藏入口）
- **管理员登录**: `http://localhost:8080/login`
- **管理后台**: `http://localhost:8080/admin/`（需要管理员登录）
- **管理仪表板**: `http://localhost:8080/admin/dashboard`（需要管理员登录）

## 路由跳转逻辑

### 1. 根路径 (`/`)
访问 `http://localhost:8080/` 时：
- **所有用户**: 自动跳转到 `/user/login`（用户门户登录页）

### 2. 用户门户路由 (`/user/*`)
- **未登录用户**: 访问需要认证的页面时，跳转到 `/user/login`
- **已登录用户**: 可以正常访问用户门户的所有页面
- **已登录用户访问登录/注册页**: 自动跳转到 `/user/dashboard`

### 3. 管理后台路由 (`/admin/*`)
- **未登录用户**: 访问管理后台时，跳转到 `/login`（管理员登录页）
- **已登录管理员**: 可以正常访问管理后台的所有页面
- **普通用户**: 无法访问管理后台，会被跳转到 `/user/login`

### 4. 管理员登录页 (`/login`)
- **未登录用户**: 可以访问并登录
- **已登录管理员**: 自动跳转到 `/admin/dashboard`

## 测试场景

### 场景 1: 未登录用户访问根路径
1. 清除所有登录状态
2. 访问 `http://localhost:8080/`
3. **预期结果**: 自动跳转到 `http://localhost:8080/user/login`

### 场景 2: 普通用户登录
1. 访问 `http://localhost:8080/user/login`
2. 使用普通用户账号登录
3. **预期结果**: 跳转到 `http://localhost:8080/user/dashboard`

### 场景 3: 管理员登录
1. 访问 `http://localhost:8080/login`（注意：不是 `/user/login`）
2. 使用管理员账号（admin）登录
3. **预期结果**: 跳转到 `http://localhost:8080/admin/dashboard`

### 场景 4: 普通用户尝试访问管理后台
1. 使用普通用户账号登录
2. 尝试访问 `http://localhost:8080/admin/dashboard`
3. **预期结果**: 被拒绝访问，跳转到 `/user/login`

### 场景 5: 已登录用户访问根路径
1. 使用普通用户账号登录
2. 访问 `http://localhost:8080/`
3. **预期结果**: 跳转到 `/user/login`，然后因为已登录，再跳转到 `/user/dashboard`

## 关键改进

### 1. 用户门户作为默认入口
- 根路径 `/` 默认跳转到用户门户登录页 `/user/login`
- 符合大多数用户的使用习惯
- 提供更友好的首次访问体验

### 2. 管理后台地址隐藏
- 管理员登录页使用 `/login` 而不是 `/admin/login`
- 管理后台路径 `/admin/*` 需要知道才能访问
- 提高了系统安全性，减少了暴力破解的风险

### 3. 清晰的角色分离
- 普通用户使用 `/user/*` 路径
- 管理员使用 `/admin/*` 路径
- 两个系统使用不同的认证令牌（`userToken` vs `token`）

## 相关文件

- `web/src/router/index.js` - 主路由配置和全局守卫
- `web/src/router/user.js` - 用户门户路由和守卫
- `web/src/views/Login.vue` - 管理员登录页面
- `web/src/views/user/Login.vue` - 用户门户登录页面

## 注意事项

1. **管理员登录地址**: `/login`（不是 `/admin/login`）
2. **用户登录地址**: `/user/login`
3. **认证令牌分离**: 
   - 管理员: `localStorage.token` + `localStorage.userRole = 'admin'`
   - 普通用户: `localStorage.userToken`
4. **管理后台隐藏**: 普通用户不应该知道 `/admin/` 路径的存在

## 开发建议

如果需要修改路由逻辑：
1. 修改 `web/src/router/index.js` 中的根路径重定向和全局前置守卫
2. 修改 `web/src/router/user.js` 中的 `userRouteGuard` 函数
3. 确保两个守卫的逻辑保持一致
4. 测试所有场景以确保跳转逻辑正确
