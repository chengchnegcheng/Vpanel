# IP 限制错误修复总结

## 问题
访问 IP 白名单/黑名单管理页面时出现数据库操作失败错误。

## 原因
IP 限制相关的数据库模型没有在 GORM AutoMigrate 中注册，导致 GORM 无法识别和查询这些表。

## 解决方案
在 `internal/database/db.go` 中：
1. 添加了 `"v/internal/ip"` 包导入
2. 在 `AutoMigrate()` 函数中添加了 7 个 IP 限制模型

## 修复的模型
- IPWhitelist（白名单）
- IPBlacklist（黑名单）
- ActiveIP（活跃 IP）
- IPHistory（IP 历史）
- SubscriptionIPAccess（订阅访问）
- GeoCache（地理缓存）
- FailedAttempt（失败尝试）

## 下一步
1. 重启应用程序：`./v`
2. 测试 IP 限制管理页面是否正常工作
3. 验证 API 端点：
   - `/api/admin/ip-whitelist`
   - `/api/admin/ip-blacklist`

## 状态
✅ 代码已修复并编译成功
⏳ 需要重启应用程序以应用修复
