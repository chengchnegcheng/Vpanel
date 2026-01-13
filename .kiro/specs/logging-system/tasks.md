# Implementation Plan: Logging System

## Overview

本实现计划将日志记录系统分解为可执行的编码任务。采用增量开发方式，从数据层开始，逐步构建服务层、API 层和前端界面。

## Tasks

- [x] 1. 数据库层实现
  - [x] 1.1 创建数据库迁移文件 `008_logs_enhancement.sql`
    - 添加 request_id 和 fields 字段
    - 创建复合索引优化查询
    - _Requirements: 1.1, 6.4_

  - [x] 1.2 扩展 Log 模型 (`internal/database/models.go`)
    - 添加 RequestID 和 Fields 字段
    - 更新 GORM 标签
    - _Requirements: 1.1, 1.5_

  - [x] 1.3 实现 LogRepository (`internal/database/repository/log_repository.go`)
    - 实现 Create, CreateBatch, GetByID, List, Count, DeleteOlderThan, DeleteByFilter 方法
    - 实现 LogFilter 结构体和过滤逻辑
    - _Requirements: 1.1, 1.4, 2.1-2.5, 4.2, 4.3_

  - [x] 1.4 编写 LogRepository 属性测试
    - **Property 2: Batch Insertion Integrity**
    - **Property 3: Unique Identifier Generation**
    - **Property 4: Pagination Correctness**
    - **Property 5: Filter Correctness**
    - **Validates: Requirements 1.4, 1.5, 2.1-2.5**

- [x] 2. Checkpoint - 确保数据库层测试通过
  - 确保所有测试通过，如有问题请询问用户

- [ ] 3. 服务层实现
  - [ ] 3.1 扩展日志配置 (`internal/config/config.go`)
    - 添加 DatabaseEnabled, DatabaseLevel, RetentionDays, BufferSize, BatchSize, FlushInterval 配置项
    - _Requirements: 5.1, 5.2, 5.3, 5.5_

  - [ ] 3.2 实现 AsyncWriter (`internal/log/async_writer.go`)
    - 实现缓冲区管理和批量写入
    - 实现定时刷新和优雅关闭
    - _Requirements: 6.1, 6.2, 6.3_

  - [ ] 3.3 实现 LogService (`internal/log/service.go`)
    - 实现 Log, LogSync, Query, GetByID, Delete, Cleanup 方法
    - 实现清理调度器
    - 集成 AsyncWriter
    - _Requirements: 1.1, 1.2, 1.3, 2.1-2.6, 4.1-4.4, 5.1_

  - [ ]* 3.4 编写 LogService 属性测试
    - **Property 1: Log Persistence Completeness**
    - **Property 7: Retention-Based Cleanup**
    - **Property 8: Manual Cleanup Filter Accuracy**
    - **Property 9: Database Level Filtering**
    - **Property 10: Async Non-Blocking Writes**
    - **Property 11: Batch Writing Efficiency**
    - **Validates: Requirements 1.1, 1.2, 4.2, 4.3, 5.1, 6.1, 6.2**

- [x] 4. Checkpoint - 确保服务层测试通过
  - 确保所有测试通过，如有问题请询问用户

- [x] 5. API 层实现
  - [x] 5.1 实现 LogHandler (`internal/api/handlers/logs.go`)
    - 实现 ListLogs, GetLog, DeleteLogs, Cleanup, ExportLogs 处理函数
    - 实现请求验证和错误处理
    - _Requirements: 2.1-2.7, 3.5, 4.3_

  - [x] 5.2 注册日志 API 路由 (`internal/api/routes.go`)
    - 添加 /api/logs 相关路由
    - 配置管理员权限中间件
    - _Requirements: 2.7_

  - [x] 5.3 编写 LogHandler 属性测试
    - **Property 6: Export Format Validity**
    - **Property 12: Query Result Limits**
    - **Validates: Requirements 3.5, 6.5**

- [x] 6. Checkpoint - 确保 API 层测试通过
  - 确保所有测试通过，如有问题请询问用户

- [x] 7. 前端实现
  - [x] 7.1 创建日志 API 模块 (`web/src/api/modules/logs.js`)
    - 实现 getLogs, getLog, deleteLogs, cleanup, exportLogs 函数
    - _Requirements: 2.1-2.6, 3.5, 4.3_

  - [x] 7.2 创建日志管理页面 (`web/src/views/Logs.vue`)
    - 实现日志列表展示和分页
    - 实现过滤控件（级别、日期、来源、关键词）
    - 实现日志详情弹窗
    - 实现导出功能
    - 实现错误/致命日志高亮
    - _Requirements: 3.1-3.6_

  - [x] 7.3 添加日志页面路由 (`web/src/router/index.js`)
    - 添加 /logs 路由
    - 配置管理员权限
    - _Requirements: 3.1_

  - [x] 7.4 更新侧边栏导航 (`web/src/components/Sidebar.vue`)
    - 添加日志管理菜单项
    - _Requirements: 3.1_

- [x] 8. 集成和连接
  - [x] 8.1 集成 LogService 到应用启动流程 (`cmd/v/main.go`)
    - 初始化 LogService
    - 启动清理调度器
    - 配置优雅关闭
    - _Requirements: 4.1, 4.5_

  - [x] 8.2 集成日志记录到现有中间件 (`internal/api/middleware/middleware.go`)
    - 在请求日志中间件中使用 LogService
    - 记录请求上下文信息
    - _Requirements: 1.2_

- [x] 9. Final Checkpoint - 确保所有测试通过
  - 确保所有测试通过，如有问题请询问用户

## Notes

- 标记 `*` 的任务为可选测试任务，可跳过以加快 MVP 开发
- 每个任务都引用了具体的需求以便追溯
- Checkpoint 任务确保增量验证
- 属性测试验证通用正确性属性
- 单元测试验证具体示例和边界情况
