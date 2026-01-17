# Admin 用户自动跳转功能说明

## 功能描述

当 admin 用户访问用户门户（`http://localhost:8080`）时，系统会自动检测用户角色并跳转到管理后台。

## 实现逻辑

### 1. 根路径智能跳转 (`/`)

访问 `http://localhost:8080/` 时：
- **Admin 用户（已登录）**: 自动跳转到 `/admin/dashboard`
- **普通用户（已登录）**: 自动跳转到 `/user/dashboard`
- **未登录用户**: 跳转到 `/user/login`

### 2. 用户门户路由守卫

Admin 用户访问用户门户的任何需要认证的页面时：
- 自动跳转到 `/admin/dashboard`
- **例外**: 允许访问登录/注册页面（`/user/login`, `/user/register`），以便管理员可以切换账号

### 3. 判断依据

系统通过以下 localStorage 数据判断用户身份：
- `token`: 管理员认证令牌
- `userToken`: 普通用户认证令牌
- `userRole`: 用户角色（`admin` 或 `user`）

## 测试场景

### 场景 1: Admin 用户访问根路径
1. 使用 admin 账号登录管理后台
2. 访问 `http://localhost:8080/`
3. **预期结果**: 自动跳转到 `http://localhost:8080/admin/dashboard`

### 场景 2: Admin 用户访问用户门户
1. 使用 admin 账号登录管理后台
2. 访问 `http://localhost:8080/user/dashboard`
3. **预期结果**: 自动跳转到 `http://localhost:8080/admin/dashboard`

### 场景 3: Admin 用户访问用户登录页
1. 使用 admin 账号登录管理后台
2. 访问 `http://localhost:8080/user/login`
3. **预期结果**: 允许访问（不跳转），以便管理员可以切换账号

### 场景 4: 普通用户访问根路径
1. 使用普通用户账号登录
2. 访问 `http://localhost:8080/`
3. **预期结果**: 自动跳转到 `http://localhost:8080/user/dashboard`

### 场景 5: 未登录用户访问根路径
1. 清除所有登录状态
2. 访问 `http://localhost:8080/`
3. **预期结果**: 自动跳转到 `http://localhost:8080/user/login`

## 相关文件

- `web/src/router/index.js` - 主路由配置和全局守卫
- `web/src/router/user.js` - 用户门户路由和守卫
- `web/src/stores/user.js` - 用户状态管理

## 注意事项

1. 管理员和普通用户使用不同的认证令牌（`token` vs `userToken`）
2. 角色信息存储在 `localStorage` 的 `userRole` 字段
3. 登录/注册页面（`meta.guest: true`）允许管理员访问，不会触发跳转
4. 所有跳转逻辑在路由守卫中处理，确保一致性

## 开发建议

如果需要修改跳转逻辑：
1. 修改 `web/src/router/index.js` 中的全局前置守卫
2. 修改 `web/src/router/user.js` 中的 `userRouteGuard` 函数
3. 确保两个守卫的逻辑保持一致
