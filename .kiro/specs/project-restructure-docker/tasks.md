# Implementation Plan: V Panel 项目重构与 Docker 部署

## Overview

本实现计划将项目重构分为四个主要阶段：目录结构重组、代码合并与清理、Docker 支持、测试与验证。每个阶段包含具体的编码任务，确保增量式进展和可验证的里程碑。

## Tasks

- [x] 1. 创建新目录结构
  - [x] 1.1 创建标准 Go 项目目录
    - 创建 `cmd/v/` 目录用于应用入口
    - 创建 `internal/` 目录用于私有包
    - 创建 `pkg/` 目录用于公共包
    - 创建 `configs/` 目录用于配置模板
    - 创建 `deployments/docker/` 目录用于 Docker 文件
    - 创建 `scripts/` 目录用于构建脚本
    - _Requirements: 1.1, 12.1, 12.2_

  - [x] 1.2 创建 internal 子目录结构
    - 创建 `internal/api/handlers/` 用于 HTTP 处理器
    - 创建 `internal/api/middleware/` 用于中间件
    - 创建 `internal/auth/` 用于认证模块
    - 创建 `internal/config/` 用于配置管理
    - 创建 `internal/database/` 用于数据库层
    - 创建 `internal/database/migrations/` 用于迁移文件
    - 创建 `internal/database/repository/` 用于数据访问
    - 创建 `internal/logger/` 用于日志模块
    - 创建 `internal/monitor/` 用于系统监控
    - 创建 `internal/notification/` 用于通知服务
    - 创建 `internal/proxy/protocols/` 用于代理协议
    - 创建 `internal/server/` 用于 HTTP 服务器
    - 创建 `internal/xray/` 用于 Xray 管理
    - _Requirements: 1.1, 1.4, 12.3_

- [x] 2. 实现配置管理模块
  - [x] 2.1 创建配置结构和加载器
    - 在 `internal/config/config.go` 中定义 Config 结构
    - 实现 YAML 配置文件加载
    - 实现环境变量覆盖（V_ 前缀）
    - 实现配置验证逻辑
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

  - [x] 2.2 编写配置优先级属性测试
    - **Property 1: Configuration Precedence**
    - **Validates: Requirements 1.6, 5.1, 5.3**

  - [x] 2.3 编写配置验证属性测试
    - **Property 4: Configuration Validation at Startup**
    - **Validates: Requirements 5.4**

  - [x] 2.4 创建配置模板文件
    - 创建 `configs/config.yaml.example`
    - 创建 `configs/xray.json.example`
    - _Requirements: 12.6_

- [x] 3. Checkpoint - 配置模块完成
  - 确保所有测试通过，如有问题请询问用户

- [x] 4. 实现日志模块
  - [x] 4.1 创建统一日志接口和实现
    - 在 `internal/logger/logger.go` 中定义 Logger 接口
    - 实现 JSON 格式日志输出
    - 实现日志级别过滤
    - 支持结构化字段
    - _Requirements: 9.1, 9.2, 9.3_

  - [x] 4.2 编写 JSON 日志格式属性测试
    - **Property 5: JSON Log Format**
    - **Validates: Requirements 9.2**

  - [x] 4.3 编写日志级别过滤属性测试
    - **Property 6: Log Level Filtering**
    - **Validates: Requirements 9.3**

- [x] 5. 实现错误处理模块
  - [x] 5.1 创建错误类型和处理函数
    - 在 `pkg/errors/errors.go` 中定义错误类型
    - 实现 AppError 结构
    - 实现错误包装函数
    - _Requirements: 2.4_

  - [x] 5.2 编写数据库错误上下文属性测试
    - **Property 2: Database Error Context**
    - **Validates: Requirements 2.4**

- [x] 6. 实现数据库层
  - [x] 6.1 创建数据库连接和迁移
    - 在 `internal/database/db.go` 中实现数据库连接
    - 移动迁移文件到 `internal/database/migrations/`
    - 实现自动迁移逻辑
    - _Requirements: 2.1, 2.2, 11.6_

  - [x] 6.2 创建 Repository 接口和实现
    - 在 `internal/database/repository/` 中定义接口
    - 实现 UserRepository
    - 实现 ProxyRepository
    - 实现 TrafficRepository
    - _Requirements: 2.3_

  - [x] 6.3 编写 Repository 单元测试
    - 使用 SQLite 内存数据库测试
    - 测试 CRUD 操作
    - _Requirements: 2.3_

- [x] 7. Checkpoint - 基础设施模块完成
  - 确保所有测试通过，如有问题请询问用户

- [x] 8. 实现认证模块
  - [x] 8.1 创建认证服务
    - 在 `internal/auth/service.go` 中实现认证逻辑
    - 实现 JWT token 生成和验证
    - 实现密码哈希和验证
    - _Requirements: 5.4_

  - [x] 8.2 创建认证中间件
    - 在 `internal/api/middleware/auth.go` 中实现
    - 实现 token 验证中间件
    - 实现角色检查中间件
    - _Requirements: 3.3_

- [x] 9. 实现代理协议模块
  - [x] 9.1 创建协议接口和管理器
    - 在 `internal/proxy/protocol.go` 中定义 Protocol 接口
    - 在 `internal/proxy/manager.go` 中实现 ProxyManager
    - _Requirements: 4.2_

  - [x] 9.2 迁移 VMess 协议实现
    - 合并 `proxy/vmess/` 和 `proxy/vmess_server.go`
    - 移动到 `internal/proxy/protocols/vmess/`
    - 实现 Protocol 接口
    - _Requirements: 4.1, 4.3_

  - [x] 9.3 迁移 VLESS 协议实现
    - 合并 `proxy/vless/` 和 `proxy/vless.go`
    - 移动到 `internal/proxy/protocols/vless/`
    - 实现 Protocol 接口
    - _Requirements: 4.1, 4.3_

  - [x] 9.4 迁移 Trojan 协议实现
    - 合并 `proxy/trojan/` 和 `proxy/trojan.go`
    - 移动到 `internal/proxy/protocols/trojan/`
    - 实现 Protocol 接口
    - _Requirements: 4.1, 4.3_

  - [x] 9.5 迁移 Shadowsocks 协议实现
    - 合并 `proxy/shadowsocks/` 和 `proxy/shadowsocks_server.go`
    - 移动到 `internal/proxy/protocols/shadowsocks/`
    - 实现 Protocol 接口
    - _Requirements: 4.1, 4.3_

- [x] 10. Checkpoint - 核心模块完成
  - 确保所有测试通过，如有问题请询问用户

- [x] 11. 实现 API 层
  - [x] 11.1 创建 API 路由和中间件
    - 在 `internal/api/routes.go` 中定义所有路由
    - 在 `internal/api/middleware/` 中实现 CORS、日志、恢复中间件
    - _Requirements: 3.2, 3.3_

  - [x] 11.2 创建认证 Handler
    - 在 `internal/api/handlers/auth.go` 中实现
    - 实现登录、登出、获取用户信息接口
    - _Requirements: 3.1, 3.4_

  - [x] 11.3 创建代理 Handler
    - 在 `internal/api/handlers/proxy.go` 中实现
    - 实现代理 CRUD 接口
    - 实现分享链接生成接口
    - _Requirements: 3.1, 3.4_

  - [x] 11.4 创建系统 Handler
    - 在 `internal/api/handlers/system.go` 中实现
    - 实现系统信息、状态接口
    - _Requirements: 3.1, 3.4_

  - [x] 11.5 创建健康检查 Handler
    - 在 `internal/api/handlers/health.go` 中实现
    - 实现 /health 和 /ready 端点
    - _Requirements: 10.1, 10.2_

  - [x] 11.6 编写 API 输入验证属性测试
    - **Property 3: API Input Validation**
    - **Validates: Requirements 3.5**

  - [x] 11.7 编写错误日志上下文属性测试
    - **Property 7: Error Logging with Request Context**
    - **Validates: Requirements 9.5**

- [x] 12. 实现 HTTP 服务器
  - [x] 12.1 创建服务器启动和优雅关闭
    - 在 `internal/server/server.go` 中实现
    - 实现优雅关闭逻辑（30秒超时）
    - 实现信号处理
    - _Requirements: 10.3, 10.4_

  - [x] 12.2 编写优雅关闭属性测试
    - **Property 8: Graceful Shutdown Completion**
    - **Validates: Requirements 10.4**

- [x] 13. 创建应用入口
  - [x] 13.1 创建 main.go
    - 在 `cmd/v/main.go` 中实现应用入口
    - 初始化所有模块
    - 启动 HTTP 服务器
    - _Requirements: 12.2_

- [x] 14. Checkpoint - 应用核心完成
  - 确保所有测试通过，如有问题请询问用户

- [x] 15. 清理冗余文件和目录
  - [x] 15.1 删除根目录冗余文件
    - 删除根目录 `src/` 目录
    - 删除根目录 `package.json` 和 `package-lock.json`
    - 删除根目录 `router/` 目录
    - _Requirements: 11.1, 11.2, 11.3_

  - [x] 15.2 清理前端冗余文件
    - 删除 `web/src/views/*.vue.new` 文件
    - 删除重复的视图文件（如 StatsNew.vue, RolesNew.vue）
    - _Requirements: 11.4_

  - [x] 15.3 删除旧的后端目录
    - 删除旧的 `api/` 目录
    - 删除旧的 `auth/` 目录
    - 删除旧的 `backup/` 目录
    - 删除旧的 `cert/` 和 `certificate/` 目录
    - 删除旧的 `common/` 目录
    - 删除旧的 `config/` 目录
    - 删除旧的 `database/` 和 `db/` 目录
    - 删除旧的 `errors/` 目录
    - 删除旧的 `logger/` 目录
    - 删除旧的 `middleware/` 目录
    - 删除旧的 `model/` 目录
    - 删除旧的 `monitor/` 目录
    - 删除旧的 `notification/` 目录
    - 删除旧的 `protocol/` 目录
    - 删除旧的 `proxy/` 目录
    - 删除旧的 `security/` 目录
    - 删除旧的 `server/` 目录
    - 删除旧的 `settings/` 目录
    - 删除旧的 `ssl/` 目录
    - 删除旧的 `stats/` 目录
    - 删除旧的 `traffic/` 目录
    - 删除旧的 `user/` 目录
    - 删除旧的 `utils/` 目录
    - 删除旧的 `version/` 目录
    - 删除旧的 `audit/` 目录
    - 删除旧的 `main.go`
    - _Requirements: 11.7, 12.1_

- [x] 16. 创建 Docker 部署文件
  - [x] 16.1 创建 Dockerfile
    - 在 `deployments/docker/Dockerfile` 中实现多阶段构建
    - Stage 1: 构建前端
    - Stage 2: 构建后端
    - Stage 3: 最终镜像
    - 配置健康检查
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 8.1, 8.3_

  - [x] 16.2 创建 docker-compose.yml
    - 在 `deployments/docker/docker-compose.yml` 中实现
    - 定义服务配置
    - 定义命名卷
    - 配置健康检查
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

  - [x] 16.3 创建 .env.example
    - 在 `deployments/docker/.env.example` 中定义环境变量模板
    - _Requirements: 7.3_

  - [x] 16.4 创建入口脚本
    - 在 `deployments/scripts/entrypoint.sh` 中实现
    - 处理数据库初始化
    - _Requirements: 6.6_

- [x] 17. 创建构建脚本
  - [x] 17.1 创建构建脚本
    - 在 `scripts/build.sh` 中实现本地构建
    - 在 `scripts/docker-build.sh` 中实现 Docker 构建
    - _Requirements: 12.7_

- [x] 18. 更新文档
  - [x] 18.1 更新 README.md
    - 更新项目结构说明
    - 添加 Docker 部署说明
    - 更新开发指南
    - _Requirements: 7.5_

- [x] 19. Final Checkpoint - 项目重构完成
  - 确保所有测试通过
  - 验证 Docker 构建成功
  - 验证应用正常运行
  - 如有问题请询问用户

## Notes

- 每个任务都引用了具体的需求以确保可追溯性
- Checkpoint 任务用于确保增量验证
- 属性测试验证通用正确性属性
- 单元测试验证具体示例和边界情况
- 所有任务都是必须完成的，包括测试任务
