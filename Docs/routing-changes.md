# 路由修改说明

## 修改概述

根据用户需求，已将系统的默认入口从管理后台改为用户门户，并隐藏了管理后台的访问地址。

## 主要变更

### 1. 根路径行为变更

**之前**:
- `http://localhost:8080/` → 根据登录状态智能跳转
  - Admin 已登录 → `/admin/dashboard`
  - 用户已登录 → `/user/dashboard`
  - 未登录 → `/user/login`

**现在**:
- `http://localhost:8080/` → 始终跳转到 `/user/login`（用户门户登录页）

### 2. 路由守卫逻辑简化

**之前**:
- Admin 用户访问用户门户会自动跳转到管理后台
- 复杂的角色判断逻辑

**现在**:
- 移除了 Admin 用户访问用户门户时的自动跳转
- 用户门户和管理后台完全独立
- 更清晰的角色分离

### 3. 访问路径说明

| 用户类型 | 登录地址 | 仪表板地址 |
|---------|---------|-----------|
| 普通用户 | `/user/login` | `/user/dashboard` |
| 管理员 | `/login` | `/admin/dashboard` |

## 修改的文件

1. **web/src/router/index.js**
   - 修改根路径重定向逻辑：始终跳转到 `/user/login`
   - 简化全局路由守卫：移除 Admin 用户的特殊处理
   - 优化管理员登录页的访问控制

2. **web/src/router/user.js**
   - 移除 Admin 用户访问用户门户时的自动跳转逻辑
   - 简化用户门户路由守卫

3. **README.md**
   - 更新访问地址说明
   - 明确管理员和用户的登录入口
   - 添加提示信息

4. **Docs/routing-guide.md** (新建)
   - 详细的路由访问说明
   - 测试场景
   - 开发建议

## 优势

### 1. 更好的用户体验
- 普通用户访问根路径直接看到用户门户登录页
- 符合大多数用户的使用习惯
- 减少了不必要的跳转

### 2. 提高安全性
- 管理后台地址相对隐藏（`/login` 而不是 `/admin/login`）
- 减少了暴力破解的风险
- 普通用户不容易发现管理后台的存在

### 3. 清晰的角色分离
- 用户门户：`/user/*`
- 管理后台：`/admin/*`
- 两个系统使用不同的认证令牌
- 互不干扰

## 测试建议

### 测试场景 1: 未登录用户
```bash
# 访问根路径
curl -I http://localhost:8080/
# 预期: 302 重定向到 /user/login
```

### 测试场景 2: 普通用户登录
1. 访问 `http://localhost:8080/user/login`
2. 输入用户名和密码
3. 预期：跳转到 `/user/dashboard`

### 测试场景 3: 管理员登录
1. 访问 `http://localhost:8080/login`
2. 输入 admin 账号和密码
3. 预期：跳转到 `/admin/dashboard`

### 测试场景 4: 权限控制
1. 使用普通用户登录
2. 尝试访问 `http://localhost:8080/admin/dashboard`
3. 预期：被拒绝，跳转到 `/user/login`

## 回滚方案

如果需要恢复之前的行为，可以：

1. 恢复 `web/src/router/index.js` 中的根路径智能跳转逻辑
2. 恢复 `web/src/router/user.js` 中的 Admin 用户自动跳转逻辑
3. 参考 Git 历史记录中的 `Docs/admin-redirect.md` 文档

## 注意事项

1. **管理员登录地址变更**: 确保管理员知道新的登录地址是 `/login`
2. **文档更新**: 所有相关文档已更新，包括 README.md
3. **向后兼容**: 现有的 API 和后端路由没有变化
4. **前端构建**: 修改后需要重新构建前端：
   ```bash
   cd web
   npm run build
   ```

## 相关文档

- [路由访问指南](./routing-guide.md) - 详细的路由说明
- [README.md](../README.md) - 项目主文档
- [admin-redirect.md](./admin-redirect.md) - 旧的路由逻辑说明（已废弃）
