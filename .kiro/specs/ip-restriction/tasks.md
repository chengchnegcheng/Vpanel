# Implementation Plan: IP Restriction System

## Overview

本实现计划将 IP 限制系统分解为可执行的开发任务，包括数据模型、核心服务、API 接口和前端界面。

## Tasks

- [x] 1. 数据模型和数据库迁移
  - [x] 1.1 创建 IP 相关数据模型
    - 创建 `internal/ip/models.go`
    - 定义 IPWhitelist, IPBlacklist, ActiveIP, IPHistory, SubscriptionIPAccess, GeoCache 模型
    - _Requirements: 1.1, 2.3, 4.2, 5.2_
  - [x] 1.2 创建数据库迁移文件
    - 创建 IP 相关表的迁移
    - 添加必要的索引 (user_id, ip, created_at)
    - _Requirements: 1.1, 2.3_
  - [x] 1.3 扩展 User 模型
    - 添加 max_concurrent_ips 字段到 User 模型
    - _Requirements: 1.1, 1.7_
  - [x] 1.4 扩展 Plan 模型
    - 添加 default_max_concurrent_ips 字段到 Plan 模型
    - _Requirements: 1.2_

- [x] 2. IP 验证器实现
  - [x] 2.1 实现 CIDR 匹配功能
    - 创建 `internal/ip/cidr.go`
    - 实现 IPv4 和 IPv6 CIDR 匹配
    - _Requirements: 4.2, 5.2_
  - [ ]* 2.2 编写 CIDR 匹配属性测试
    - **Property 9: CIDR Range Matching**
    - **Validates: Requirements 4.2, 5.2**
  - [x] 2.3 实现白名单验证
    - 创建 `internal/ip/validator.go`
    - 实现 IsWhitelisted 方法，支持全局和用户级白名单
    - _Requirements: 4.1, 4.3, 4.4_
  - [x] 2.4 实现黑名单验证
    - 实现 IsBlacklisted 方法，支持过期检查
    - _Requirements: 5.1, 5.3, 5.5_
  - [ ]* 2.5 编写白名单/黑名单属性测试
    - **Property 10: Whitelist Bypass**
    - **Property 12: Blacklist Rejection**
    - **Property 13: Temporary Blacklist Expiration**
    - **Validates: Requirements 4.3, 5.3, 5.5**

- [x] 3. IP 追踪器实现
  - [x] 3.1 实现活跃 IP 追踪
    - 创建 `internal/ip/tracker.go`
    - 实现 AddActiveIP, RemoveActiveIP, GetActiveIPCount 方法
    - _Requirements: 1.5, 3.1_
  - [x] 3.2 实现不活跃 IP 清理
    - 实现 CleanupInactiveIPs 方法
    - 添加定时清理任务
    - _Requirements: 1.6_
  - [ ]* 3.3 编写不活跃清理属性测试
    - **Property 2: Inactive IP Cleanup**
    - **Validates: Requirements 1.6**
  - [x] 3.4 实现 IP 历史记录
    - 实现 RecordIPHistory, GetIPHistory 方法
    - _Requirements: 2.1, 2.2, 2.3_
  - [ ]* 3.5 编写 IP 记录属性测试
    - **Property 4: IP Activity Recording**
    - **Validates: Requirements 2.1, 2.2, 2.3**

- [x] 4. 地理位置服务实现
  - [x] 4.1 集成 MaxMind GeoLite2
    - 创建 `internal/ip/geolocation.go`
    - 实现 Lookup 方法
    - _Requirements: 7.4_
  - [x] 4.2 实现地理位置缓存
    - 实现缓存层，减少数据库查询
    - _Requirements: 7.6_
  - [ ]* 4.3 编写地理位置缓存属性测试
    - **Property 17: Geolocation Cache**
    - **Validates: Requirements 7.6**
  - [x] 4.4 实现地理位置限制检查
    - 实现 CheckGeoRestriction 方法
    - _Requirements: 7.2, 7.3_
  - [ ]* 4.5 编写地理位置限制属性测试
    - **Property 16: Geo Restriction Enforcement**
    - **Validates: Requirements 7.2, 7.3**

- [ ] 5. Checkpoint - 核心服务测试
  - 确保所有属性测试通过
  - 如有问题请询问用户

- [x] 6. IP 限制核心服务
  - [x] 6.1 实现 IP 限制服务
    - 创建 `internal/ip/service.go`
    - 实现 CheckAccess 方法，整合所有验证逻辑
    - _Requirements: 1.3, 1.4, 4.3, 5.3, 7.2_
  - [x] 6.2 实现并发 IP 限制检查
    - 在 CheckAccess 中实现并发 IP 限制逻辑
    - 支持无限制选项 (0 或 -1)
    - _Requirements: 1.3, 1.4, 1.8_
  - [ ]* 6.3 编写并发 IP 限制属性测试
    - **Property 1: Concurrent IP Limit Enforcement**
    - **Property 3: Unlimited IP Option**
    - **Validates: Requirements 1.3, 1.4, 1.8**
  - [x] 6.4 实现设备踢出功能
    - 实现 KickIP 方法
    - 添加临时阻止逻辑
    - _Requirements: 3.3, 3.4_
  - [ ]* 6.5 编写设备踢出属性测试
    - **Property 7: Device Kick Temporary Block**
    - **Validates: Requirements 3.4**
  - [x] 6.6 实现可疑模式检测
    - 检测短时间内多国家 IP 访问
    - _Requirements: 2.8_
  - [ ]* 6.7 编写可疑模式检测属性测试
    - **Property 6: Suspicious Pattern Detection**
    - **Validates: Requirements 2.8**


- [x] 7. 订阅链接 IP 限制
  - [x] 7.1 实现订阅链接 IP 追踪
    - 创建 SubscriptionIPAccess 仓库
    - 实现 IP 访问记录和统计
    - _Requirements: 6.4_
  - [x] 7.2 实现订阅链接 IP 限制检查
    - 在订阅处理中集成 IP 限制检查
    - _Requirements: 6.1, 6.2, 6.3_
  - [ ]* 7.3 编写订阅 IP 限制属性测试
    - **Property 15: Subscription IP Limit**
    - **Validates: Requirements 6.3, 6.4**
  - [x] 7.4 实现订阅 IP 列表重置
    - 在重新生成订阅 token 时清除 IP 列表
    - _Requirements: 6.6_

- [x] 8. 自动黑名单功能
  - [x] 8.1 实现失败尝试追踪
    - 追踪 IP 的失败访问尝试
    - _Requirements: 5.4_
  - [x] 8.2 实现自动黑名单触发
    - 超过阈值时自动添加到黑名单
    - _Requirements: 5.4_
  - [ ]* 8.3 编写自动黑名单属性测试
    - **Property 14: Auto-Blacklist Trigger**
    - **Validates: Requirements 5.4**

- [ ] 9. Checkpoint - 服务层测试
  - 确保所有属性测试通过
  - 如有问题请询问用户

- [-] 10. 管理员 API 接口
  - [x] 10.1 实现 IP 统计 API
    - GET /api/admin/ip-restrictions/stats
    - _Requirements: 9.1_
  - [x] 10.2 实现用户在线 IP API
    - GET /api/admin/users/:id/online-ips
    - POST /api/admin/users/:id/kick-ip
    - _Requirements: 9.2, 9.3_
  - [x] 10.3 实现白名单管理 API
    - CRUD /api/admin/ip-whitelist
    - 支持批量导入
    - _Requirements: 9.4, 4.5_
  - [x] 10.4 实现黑名单管理 API
    - CRUD /api/admin/ip-blacklist
    - _Requirements: 9.5_
  - [x] 10.5 实现 IP 限制设置 API
    - GET/PUT /api/admin/settings/ip-restriction
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

- [x] 11. 用户 API 接口
  - [x] 11.1 实现用户设备列表 API
    - GET /api/user/devices
    - _Requirements: 9.6, 3.1, 3.2_
  - [x] 11.2 实现用户踢出设备 API
    - POST /api/user/devices/:ip/kick
    - _Requirements: 9.7, 3.3_

- [x] 12. IP 限制中间件
  - [x] 12.1 创建 IP 限制中间件
    - 创建 `internal/middleware/ip_restriction.go`
    - 在请求处理前检查 IP 限制
    - _Requirements: 1.3, 5.3, 7.2_
  - [x] 12.2 集成到订阅处理
    - 在订阅 API 中应用 IP 限制
    - _Requirements: 6.1, 6.2, 6.3_

- [ ] 13. Checkpoint - API 测试
  - 确保所有 API 端点正常工作
  - 如有问题请询问用户

- [x] 14. 管理后台界面
  - [x] 14.1 创建 IP 限制设置页面
    - 创建 `web/src/views/admin/IPRestriction.vue`
    - 实现全局设置配置
    - _Requirements: 8.1, 8.2, 8.3, 8.4_
  - [x] 14.2 创建白名单管理页面
    - 实现白名单列表、添加、删除、导入
    - _Requirements: 4.1, 4.5, 4.6_
  - [x] 14.3 创建黑名单管理页面
    - 实现黑名单列表、添加、删除
    - 显示原因和过期时间
    - _Requirements: 5.1, 5.6_
  - [x] 14.4 创建用户 IP 详情组件
    - 在用户详情页显示在线 IP 和历史
    - _Requirements: 2.5, 9.2_
  - [x] 14.5 创建 IP 统计仪表板
    - 显示 IP 限制相关统计
    - 按国家显示访问统计
    - _Requirements: 8.5, 7.7_

- [x] 15. 用户前台界面
  - [x] 15.1 创建设备管理页面
    - 创建 `web/src/views/user/Devices.vue`
    - 显示在线设备列表
    - _Requirements: 3.1, 3.2_
  - [x] 15.2 实现设备踢出功能
    - 添加踢出按钮和确认对话框
    - _Requirements: 3.3_
  - [x] 15.3 显示剩余设备槽位
    - 在仪表板和设备页面显示
    - _Requirements: 3.5, 3.6_
  - [x] 15.4 显示订阅链接访问 IP
    - 在订阅页面显示访问过的 IP
    - _Requirements: 6.5_

- [x] 16. 通知集成
  - [x] 16.1 实现新设备连接通知
    - 发送邮件/Telegram 通知
    - _Requirements: 10.1, 3.7_
  - [x] 16.2 实现 IP 限制达到通知
    - 提示用户断开其他设备
    - _Requirements: 10.2_
  - [x] 16.3 实现可疑活动告警
    - 向管理员发送告警
    - _Requirements: 10.3_
  - [x] 16.4 实现用户通知偏好设置
    - 允许用户配置 IP 相关通知
    - _Requirements: 10.5_

- [x] 17. Final Checkpoint
  - 确保所有测试通过
  - 确保所有功能正常工作
  - 如有问题请询问用户

## Notes

- 任务标记 `*` 为可选的属性测试任务
- 每个 Checkpoint 用于验证阶段性成果
- 属性测试使用 Go 的 `testing/quick` 或 `gopter` 库
