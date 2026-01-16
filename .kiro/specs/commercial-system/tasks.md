# Implementation Plan: Commercial System

## Overview

本实现计划将商业化系统分解为可执行的编码任务，按照数据库 → 后端服务 → API → 前端的顺序实现。使用 Go 语言实现后端，Vue 3 + Element Plus 实现前端。

## Tasks

### Phase 1: 数据库和基础设施

- [x] 1. 创建数据模型和数据库迁移
  - [x] 1.1 创建套餐模型
    - 在 `internal/database/models.go` 添加 Plan 结构体
    - 定义所有字段和 GORM 标签
    - _Requirements: 1.1, 1.2_
  - [x] 1.2 创建订单模型
    - 添加 Order 结构体
    - 定义订单状态常量
    - _Requirements: 3.2, 5.4_
  - [x] 1.3 创建余额交易模型
    - 添加 BalanceTransaction 结构体
    - _Requirements: 6.5, 6.6_
  - [x] 1.4 创建优惠券模型
    - 添加 Coupon 和 CouponUsage 结构体
    - _Requirements: 8.2, 8.3_
  - [x] 1.5 创建邀请相关模型
    - 添加 InviteCode、Referral、Commission 结构体
    - _Requirements: 9.1, 10.1_
  - [x] 1.6 创建发票模型
    - 添加 Invoice 结构体
    - _Requirements: 11.1, 11.2_
  - [x] 1.7 创建数据库迁移文件
    - 在 `internal/database/migrations/` 创建迁移文件
    - 创建所有表和索引
    - 添加 users 表的 balance 字段
    - _Requirements: 1.1-14.10_

- [x] 2. 创建 Repository 层
  - [x] 2.1 创建 PlanRepository
    - 在 `internal/database/repository/` 创建 `plan_repository.go`
    - 实现 CRUD 和查询方法
    - _Requirements: 1.1-1.10_
  - [x] 2.2 创建 OrderRepository
    - 创建 `order_repository.go`
    - 实现订单 CRUD 和状态更新方法
    - _Requirements: 3.1-5.10_
  - [x] 2.3 创建 BalanceRepository
    - 创建 `balance_repository.go`
    - 实现余额查询和交易记录方法
    - _Requirements: 6.1-6.10_
  - [x] 2.4 创建 CouponRepository
    - 创建 `coupon_repository.go`
    - 实现优惠券 CRUD 和使用记录方法
    - _Requirements: 8.1-8.10_
  - [x] 2.5 创建 InviteRepository
    - 创建 `invite_repository.go`
    - 实现邀请码、推荐关系、佣金的 CRUD 方法
    - _Requirements: 9.1-10.10_
  - [x] 2.6 创建 InvoiceRepository
    - 创建 `invoice_repository.go`
    - 实现发票 CRUD 方法
    - _Requirements: 11.1-11.10_

- [x] 3. Checkpoint - 确保数据库层完成
  - 运行数据库迁移
  - 验证表结构
  - 如有问题请询问用户

### Phase 2: 核心业务服务

- [x] 4. 实现套餐服务
  - [x] 4.1 创建 PlanService
    - 在 `internal/commercial/plan/` 创建 `service.go`
    - 实现套餐 CRUD 操作
    - 实现月均价格计算
    - _Requirements: 1.1-1.10_
  - [x] 4.2 编写属性测试：月均价格计算
    - **Property 11: Plan Price Per Month Calculation**
    - **Validates: Requirements 2.4**
  - [x] 4.3 编写单元测试
    - 测试套餐创建、更新、删除
    - 测试状态切换
    - _Requirements: 1.1, 1.6_

- [x] 5. 实现订单服务
  - [x] 5.1 创建 OrderService 基础结构
    - 在 `internal/commercial/order/` 创建 `service.go`
    - 实现订单号生成
    - 实现订单创建
    - _Requirements: 3.1-3.3_
  - [x] 5.2 编写属性测试：订单号唯一性
    - **Property 1: Order ID Uniqueness**
    - **Validates: Requirements 3.3**
  - [x] 5.3 实现订单状态管理
    - 实现状态转换逻辑
    - 实现订单过期处理
    - _Requirements: 5.4, 3.7, 3.8_
  - [x] 5.4 编写属性测试：订单状态转换
    - **Property 5: Order Status Transitions**
    - **Validates: Requirements 5.4**
  - [x] 5.5 编写属性测试：订单过期
    - **Property 7: Order Expiration**
    - **Validates: Requirements 3.7, 3.8**

- [x] 6. 实现余额服务
  - [x] 6.1 创建 BalanceService
    - 在 `internal/commercial/balance/` 创建 `service.go`
    - 实现余额查询
    - 实现充值、扣款、退款操作
    - _Requirements: 6.1-6.5_
  - [x] 6.2 编写属性测试：余额非负
    - **Property 3: Balance Non-Negative Invariant**
    - **Validates: Requirements 6.9**
  - [x] 6.3 编写属性测试：交易一致性
    - **Property 4: Balance Transaction Consistency**
    - **Validates: Requirements 6.4, 6.5**
  - [x] 6.4 实现管理员余额调整
    - 实现手动调整功能
    - 记录操作日志
    - _Requirements: 6.8_

- [x] 7. 实现优惠券服务
  - [x] 7.1 创建 CouponService
    - 在 `internal/commercial/coupon/` 创建 `service.go`
    - 实现优惠券 CRUD
    - 实现折扣计算
    - _Requirements: 8.1-8.3_
  - [x] 7.2 编写属性测试：折扣计算
    - **Property 2: Coupon Discount Calculation**
    - **Validates: Requirements 3.5, 8.2**
  - [x] 7.3 实现优惠券验证
    - 验证有效期、使用限制、最低金额
    - _Requirements: 8.4-8.8_
  - [x] 7.4 编写属性测试：使用限制
    - **Property 6: Coupon Usage Limit**
    - **Validates: Requirements 8.5, 8.6**
  - [x] 7.5 编写属性测试：验证规则
    - **Property 14: Coupon Validation Rules**
    - **Validates: Requirements 8.4, 8.7**
  - [x] 7.6 实现批量生成优惠码
    - 生成指定数量的唯一优惠码
    - _Requirements: 8.10_

- [x] 8. Checkpoint - 确保核心服务测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 3: 支付和邀请服务

- [x] 9. 实现支付网关
  - [x] 9.1 创建支付网关接口
    - 在 `internal/commercial/payment/` 创建 `gateway.go`
    - 定义 PaymentGateway 接口
    - _Requirements: 4.1-4.4_
  - [x] 9.2 实现支付宝网关
    - 创建 `alipay.go`
    - 实现支付创建、回调验证、退款
    - _Requirements: 4.1_
  - [x] 9.3 实现微信支付网关
    - 创建 `wechat.go`
    - 实现支付创建、回调验证、退款
    - _Requirements: 4.2_
  - [x] 9.4 创建 PaymentService
    - 创建 `service.go`
    - 实现支付流程管理
    - 实现回调处理
    - _Requirements: 4.5-4.9_
  - [x] 9.5 编写属性测试：回调幂等性
    - **Property 10: Payment Callback Idempotency**
    - **Validates: Requirements 14.8**
  - [x] 9.6 编写属性测试：订阅激活
    - **Property 13: Subscription Activation on Payment**
    - **Validates: Requirements 4.7**

- [x] 10. 实现邀请服务
  - [x] 10.1 创建 InviteService
    - 在 `internal/commercial/invite/` 创建 `service.go`
    - 实现邀请码生成和查询
    - 实现推荐关系记录
    - _Requirements: 9.1-9.5_
  - [x] 10.2 编写属性测试：邀请码唯一性
    - **Property 8: Invite Code Uniqueness**
    - **Validates: Requirements 9.1**
  - [x] 10.3 实现邀请统计
    - 统计邀请数量、转化率
    - _Requirements: 9.8_

- [x] 11. 实现佣金服务
  - [x] 11.1 创建 CommissionService
    - 在 `internal/commercial/commission/` 创建 `service.go`
    - 实现佣金计算
    - 实现佣金确认和取消
    - _Requirements: 10.1-10.6_
  - [x] 11.2 编写属性测试：佣金计算
    - **Property 9: Commission Calculation**
    - **Validates: Requirements 10.1**
  - [x] 11.3 实现定时确认任务
    - 实现延迟结算逻辑
    - _Requirements: 10.6_

- [x] 12. 实现发票服务
  - [x] 12.1 创建 InvoiceService
    - 在 `internal/commercial/invoice/` 创建 `service.go`
    - 实现发票生成
    - 实现发票号生成
    - _Requirements: 11.1-11.6_
  - [x] 12.2 实现 PDF 生成
    - 使用 PDF 库生成发票文件
    - _Requirements: 11.3_

- [x] 13. 实现退款服务
  - [x] 13.1 在 OrderService 添加退款逻辑
    - 实现全额和部分退款
    - 处理佣金回退
    - _Requirements: 13.1-13.7_
  - [x] 13.2 编写属性测试：退款余额恢复
    - **Property 12: Refund Balance Restoration**
    - **Validates: Requirements 13.4, 13.5**

- [x] 14. Checkpoint - 确保支付和邀请服务测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 4: 后端 API 层

- [x] 15. 实现 API Handlers
  - [x] 15.1 创建 PlanHandler
    - 在 `internal/api/handlers/` 创建 `plan.go`
    - 实现套餐列表、详情、管理端点
    - _Requirements: 1.1-2.8_
  - [x] 15.2 创建 OrderHandler
    - 创建 `order.go`
    - 实现订单创建、查询、取消端点
    - _Requirements: 3.1-5.10_
  - [x] 15.3 创建 PaymentHandler
    - 创建 `payment.go`
    - 实现支付创建、回调、状态查询端点
    - _Requirements: 4.1-4.10_
  - [x] 15.4 创建 BalanceHandler
    - 创建 `balance.go`
    - 实现余额查询、充值、交易历史端点
    - _Requirements: 6.1-6.10_
  - [x] 15.5 创建 CouponHandler
    - 创建 `coupon.go`
    - 实现优惠券验证、管理端点
    - _Requirements: 8.1-8.10_
  - [x] 15.6 创建 InviteHandler
    - 创建 `invite.go`
    - 实现邀请码、推荐、佣金端点
    - _Requirements: 9.1-10.10_
  - [x] 15.7 创建 InvoiceHandler
    - 创建 `invoice.go`
    - 实现发票列表、下载端点
    - _Requirements: 11.1-11.10_
  - [x] 15.8 创建 ReportHandler
    - 创建 `report.go`
    - 实现财务报表端点
    - _Requirements: 12.1-12.10_

- [x] 16. 注册路由
  - [x] 16.1 在 routes.go 中添加商业化路由
    - 添加 `/api/plans`、`/api/orders` 等路由
    - 添加 `/api/admin/plans`、`/api/admin/orders` 等管理路由
    - 配置认证中间件
    - _Requirements: 1.1-14.10_

- [x] 17. 编写 API 集成测试
  - [x] 17.1 编写订单流程测试
    - 测试完整订单流程
    - _Requirements: 3.1-5.10_
  - [x] 17.2 编写支付流程测试
    - 测试支付回调处理
    - _Requirements: 4.1-4.10_

- [x] 18. Checkpoint - 确保后端 API 测试通过
  - 运行所有测试，确保通过
  - 如有问题请询问用户

### Phase 5: 前端基础设施

- [x] 19. 创建前端 API 模块
  - [x] 19.1 创建 plans.ts
    - 在 `web/src/api/modules/` 创建套餐 API
    - _Requirements: 1.1-2.8_
  - [x] 19.2 创建 orders.ts
    - 创建订单 API
    - _Requirements: 3.1-5.10_
  - [x] 19.3 创建 payments.ts
    - 创建支付 API
    - _Requirements: 4.1-4.10_
  - [x] 19.4 创建 balance.ts
    - 创建余额 API
    - _Requirements: 6.1-6.10_
  - [x] 19.5 创建 coupons.ts
    - 创建优惠券 API
    - _Requirements: 8.1-8.10_
  - [x] 19.6 创建 invites.ts
    - 创建邀请 API
    - _Requirements: 9.1-10.10_
  - [x] 19.7 创建 invoices.ts
    - 创建发票 API
    - _Requirements: 11.1-11.10_

- [x] 20. 创建 Pinia Stores
  - [x] 20.1 创建 plan.ts store
    - 在 `web/src/stores/` 创建套餐状态管理
    - _Requirements: 1.1-2.8_
  - [x] 20.2 创建 order.ts store
    - 创建订单状态管理
    - _Requirements: 3.1-5.10_
  - [x] 20.3 创建 balance.ts store
    - 创建余额状态管理
    - _Requirements: 6.1-6.10_
  - [x] 20.4 创建 invite.ts store
    - 创建邀请状态管理
    - _Requirements: 9.1-10.10_

- [x] 21. Checkpoint - 确保前端基础设施完成
  - 验证 API 模块和 Store
  - 如有问题请询问用户

### Phase 6: 前端用户页面

- [x] 22. 实现套餐页面
  - [x] 22.1 创建 Plans.vue
    - 在 `web/src/views/` 创建套餐列表页面
    - 显示套餐卡片、价格、功能
    - _Requirements: 2.1-2.8_
  - [x] 22.2 创建 PlanCard.vue 组件
    - 在 `web/src/components/commercial/` 创建套餐卡片组件
    - _Requirements: 2.2-2.4_

- [x] 23. 实现订单页面
  - [x] 23.1 创建 Orders.vue
    - 创建订单历史页面
    - 显示订单列表和状态
    - _Requirements: 5.1-5.3_
  - [x] 23.2 创建 OrderDetail.vue
    - 创建订单详情页面
    - 显示订单信息、支付详情
    - _Requirements: 5.3, 5.9_
  - [x] 23.3 创建 Payment.vue
    - 创建支付页面
    - 显示支付方式选择、优惠券输入
    - _Requirements: 3.4, 4.5_

- [x] 24. 实现余额页面
  - [x] 24.1 创建 Balance.vue
    - 创建余额和充值页面
    - 显示余额、交易历史
    - _Requirements: 6.1-6.6_
  - [x] 24.2 创建 BalanceCard.vue 组件
    - 创建余额显示卡片
    - _Requirements: 6.1_

- [x] 25. 实现邀请页面
  - [x] 25.1 创建 Invite.vue
    - 创建邀请推广页面
    - 显示邀请码、邀请链接、二维码
    - _Requirements: 9.2, 9.7_
  - [x] 25.2 创建 InviteCard.vue 组件
    - 创建邀请码显示组件
    - _Requirements: 9.2_
  - [x] 25.3 创建 CommissionList.vue 组件
    - 创建佣金历史列表
    - _Requirements: 10.7_

- [x] 26. 实现发票页面
  - [x] 26.1 创建 Invoices.vue
    - 创建发票历史页面
    - 支持下载 PDF
    - _Requirements: 11.3, 11.7_

- [x] 27. Checkpoint - 确保用户页面完成
  - 验证所有用户页面功能
  - 如有问题请询问用户

### Phase 7: 前端管理页面

- [x] 28. 实现套餐管理页面
  - [x] 28.1 创建 AdminPlans.vue
    - 在 `web/src/views/admin/` 创建套餐管理页面
    - 支持创建、编辑、删除、启用/禁用
    - _Requirements: 1.1, 1.6_

- [x] 29. 实现订单管理页面
  - [x] 29.1 创建 AdminOrders.vue
    - 创建订单管理页面
    - 支持搜索、筛选、状态更新
    - _Requirements: 5.6, 5.7_

- [x] 30. 实现优惠券管理页面
  - [x] 30.1 创建 AdminCoupons.vue
    - 创建优惠券管理页面
    - 支持创建、编辑、批量生成
    - _Requirements: 8.1, 8.9, 8.10_

- [x] 31. 实现邀请统计页面
  - [x] 31.1 创建 AdminInvites.vue
    - 创建邀请统计页面
    - 显示推荐网络、统计数据
    - _Requirements: 9.9_

- [x] 32. 实现财务报表页面
  - [x] 32.1 创建 AdminReports.vue
    - 创建财务报表页面
    - 显示收入图表、订单统计
    - _Requirements: 12.1-12.10_
  - [x] 32.2 创建 RevenueChart.vue 组件
    - 创建收入图表组件
    - _Requirements: 12.5_

- [x] 33. 实现退款管理页面
  - [x] 33.1 创建 AdminRefunds.vue
    - 创建退款管理页面（集成在 AdminOrders.vue 中）
    - 支持处理退款请求
    - _Requirements: 13.1-13.9_

- [x] 34. 添加路由配置
  - [x] 34.1 更新 router 配置
    - 添加所有商业化页面路由
    - 配置权限守卫
    - _Requirements: 1.1-14.10_

- [x] 35. Final Checkpoint - 完整功能测试
  - 运行所有后端测试
  - 运行前端构建
  - 手动测试完整流程
  - 如有问题请询问用户

## Notes

- 所有任务都是必须完成的，包括测试任务
- 每个任务都引用了具体的需求编号以便追溯
- Checkpoint 任务用于确保阶段性质量
- 属性测试验证系统的正确性属性
- 支付网关集成需要配置真实的商户凭证进行测试
- 前端开发可以与后端 API 并行进行（使用 mock 数据）


### Phase 8: 新增功能 (Requirements 15-20)

- [x] 36. 实现套餐试用功能
  - [x] 36.1 创建 Trial 数据模型
    - 添加 Trial 结构体和数据库迁移
    - _Requirements: 15.1, 15.2_
  - [x] 36.2 创建 TrialService
    - 在 `internal/commercial/trial/` 创建 `service.go`
    - 实现试用激活、过期检查
    - _Requirements: 15.3-15.5_
  - [x]* 36.3 编写属性测试：试用唯一性
    - **Property 15: Trial Uniqueness**
    - **Validates: Requirements 15.3**
  - [x] 36.4 创建 Trial API 端点
    - GET/POST /api/trial
    - _Requirements: 15.4, 15.5_
  - [x] 36.5 实现试用转化率统计
    - _Requirements: 15.9_

- [x] 37. 实现套餐升降级功能
  - [x] 37.1 创建 PlanChangeService
    - 在 `internal/commercial/planchange/` 创建 `service.go`
    - 实现升级价格计算
    - _Requirements: 16.2, 16.3_
  - [x]* 37.2 编写属性测试：升级价格计算
    - **Property 16: Plan Change Proration**
    - **Validates: Requirements 16.3**
  - [x] 37.3 实现降级调度
    - 实现下个周期生效的降级
    - _Requirements: 16.4_
  - [x] 37.4 创建 Plan Change API 端点
    - POST /api/plan-change/calculate
    - POST /api/plan-change/upgrade
    - POST /api/plan-change/downgrade
    - _Requirements: 16.1-16.6_
  - [x] 37.5 创建升级页面组件
    - 显示价格差异和新功能
    - _Requirements: 16.6_

- [x] 38. 实现支付失败处理
  - [x] 38.1 创建 PaymentRetryService
    - 在 `internal/commercial/payment/` 创建 `retry.go`
    - 实现重试调度
    - _Requirements: 17.2, 17.3_
  - [x]* 38.2 编写属性测试：重试次数限制
    - **Property 17: Payment Retry Limit**
    - **Validates: Requirements 17.2**
  - [x] 38.3 实现定时重试任务
    - 创建 cron job 执行重试
    - _Requirements: 17.2_
  - [x] 38.4 实现支付方式切换
    - 允许失败订单切换支付方式
    - _Requirements: 17.6_
  - [x] 38.5 添加失败统计到管理面板
    - _Requirements: 17.8_

- [x] 39. 实现多币种支持
  - [x] 39.1 创建 CurrencyService
    - 在 `internal/commercial/currency/` 创建 `service.go`
    - 实现汇率获取和转换
    - _Requirements: 18.1, 18.6_
  - [x]* 39.2 编写属性测试：货币转换一致性
    - **Property 18: Currency Conversion Consistency**
    - **Validates: Requirements 18.6**
  - [x] 39.3 扩展 Plan 模型支持多币种价格
    - _Requirements: 18.2_
  - [x] 39.4 实现货币自动检测
    - 基于 IP 检测用户货币
    - _Requirements: 18.3_
  - [x] 39.5 添加货币选择器到前端
    - _Requirements: 18.4_
  - [x] 39.6 实现汇率定时更新任务
    - _Requirements: 18.7_

- [x] 40. 实现订阅暂停功能
  - [x] 40.1 创建 PauseService
    - 在 `internal/commercial/pause/` 创建 `service.go`
    - 实现暂停和恢复逻辑
    - _Requirements: 19.1, 19.2, 19.6_
  - [x]* 40.2 编写属性测试：暂停时长限制
    - **Property 19: Pause Duration Limit**
    - **Validates: Requirements 19.3**
  - [x] 40.3 实现暂停频率限制
    - _Requirements: 19.4_
  - [x] 40.4 实现自动恢复任务
    - _Requirements: 19.9_
  - [x] 40.5 创建 Pause API 端点
    - GET/POST /api/subscription/pause
    - POST /api/subscription/resume
    - _Requirements: 19.1, 19.6_
  - [x] 40.6 添加暂停功能到用户仪表板
    - _Requirements: 19.1_

- [x] 41. 实现礼品卡系统
  - [x] 41.1 创建 GiftCard 数据模型
    - 添加 GiftCard 结构体和数据库迁移
    - _Requirements: 20.1, 20.2_
  - [x] 41.2 创建 GiftCardService
    - 在 `internal/commercial/giftcard/` 创建 `service.go`
    - 实现创建、购买、兑换
    - _Requirements: 20.3-20.6_
  - [x]* 41.3 编写属性测试：礼品卡兑换
    - **Property 20: Gift Card Redemption**
    - **Validates: Requirements 20.6**
  - [x] 41.4 创建 Gift Card API 端点
    - POST /api/gift-cards/redeem
    - GET /api/gift-cards
    - POST /api/gift-cards/purchase
    - _Requirements: 20.3, 20.5_
  - [x] 41.5 创建管理员批量创建端点
    - POST /api/admin/gift-cards/batch
    - _Requirements: 20.1_
  - [x] 41.6 创建礼品卡兑换页面
    - _Requirements: 20.5_
  - [x] 41.7 创建管理员礼品卡管理页面
    - _Requirements: 20.9_

- [x] 42. Checkpoint - 新增功能测试
  - 运行所有新增功能的测试
  - 验证属性测试通过
  - 如有问题请询问用户

- [x] 43. Final Integration
  - [x] 43.1 更新用户仪表板
    - 添加试用状态显示
    - 添加暂停状态显示
    - _Requirements: 15.5, 19.1_
  - [x] 43.2 更新订阅页面
    - 添加升级/降级选项
    - 添加暂停/恢复按钮
    - _Requirements: 16.1, 19.1_
  - [x] 43.3 更新管理后台
    - 添加试用管理
    - 添加礼品卡管理
    - 添加暂停统计
    - _Requirements: 15.8, 19.10, 20.9_

- [x] 44. Final Checkpoint - 完整系统测试
  - 运行所有后端测试
  - 运行前端构建
  - 手动测试完整流程
  - 如有问题请询问用户

## Additional Notes

- Phase 8 中标记 `*` 的任务为可选的属性测试任务
- 新增功能可以在基础商业化系统完成后独立实现
- 多币种支持需要配置汇率 API
- 礼品卡功能可以作为独立模块部署
