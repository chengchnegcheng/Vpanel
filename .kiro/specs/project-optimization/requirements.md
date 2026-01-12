# Requirements Document

## Introduction

本文档定义了 V Panel 项目深度优化改进的需求规范。V Panel 是一个基于 Go 语言和 Vue.js 的高性能代理服务器管理面板，支持 VMess、VLESS、Trojan、Shadowsocks 等多种代理协议。通过本次优化，旨在提升项目的安全性、性能、可维护性和测试覆盖率。

## Glossary

- **V_Panel**: V Panel 应用程序主系统
- **Auth_Service**: 认证授权服务模块
- **Proxy_Manager**: 代理协议管理器
- **Repository**: 数据访问层仓库接口
- **API_Handler**: HTTP API 请求处理器
- **Middleware**: HTTP 中间件组件
- **Frontend_Store**: Vue.js 前端状态管理存储
- **Config_Manager**: 配置管理模块

## Requirements

### Requirement 1: 安全性增强

**User Story:** As a system administrator, I want enhanced security measures, so that the system is protected against common attack vectors and unauthorized access.

#### Acceptance Criteria

1. WHEN a new admin account is created with default credentials THEN the Auth_Service SHALL force password change on first login
2. WHEN a user attempts login THEN the Auth_Service SHALL implement rate limiting with maximum 5 attempts per minute per IP
3. WHEN storing authentication tokens THEN the Frontend_Store SHALL use httpOnly cookies instead of localStorage to prevent XSS attacks
4. WHEN processing user input THEN the API_Handler SHALL sanitize all input data before processing
5. WHEN admin performs sensitive operations THEN the V_Panel SHALL log the action with user ID, timestamp, IP address, and operation details
6. IF a JWT token is compromised THEN the Auth_Service SHALL support token revocation through a blacklist mechanism
7. WHEN generating JWT secrets THEN the Config_Manager SHALL require a minimum 32-character secret key

### Requirement 2: 错误处理标准化

**User Story:** As a developer, I want standardized error handling across the application, so that debugging is easier and users receive consistent error messages.

#### Acceptance Criteria

1. THE API_Handler SHALL return errors in a consistent JSON format containing error code, message, and optional details field
2. WHEN an error occurs THEN the API_Handler SHALL map internal errors to appropriate HTTP status codes
3. WHEN validation fails THEN the API_Handler SHALL return field-specific error messages indicating which fields failed validation
4. WHEN an unexpected error occurs THEN the V_Panel SHALL log the full error stack trace while returning a sanitized message to the client
5. THE Error_Response SHALL follow the format: `{"code": "ERROR_CODE", "message": "User-friendly message", "details": {}}`

### Requirement 3: 请求验证中间件

**User Story:** As a developer, I want centralized request validation, so that validation logic is not duplicated across handlers.

#### Acceptance Criteria

1. WHEN a request is received THEN the Middleware SHALL validate request body against defined schemas before reaching handlers
2. WHEN validation fails THEN the Middleware SHALL return a 400 Bad Request with detailed validation errors
3. THE Validation_Middleware SHALL support JSON schema validation for request bodies
4. WHEN defining API endpoints THEN the API_Handler SHALL declare validation rules using struct tags
5. THE Validation_Middleware SHALL validate query parameters, path parameters, and request headers

### Requirement 4: 缓存层实现

**User Story:** As a system administrator, I want caching for frequently accessed data, so that database load is reduced and response times are improved.

#### Acceptance Criteria

1. WHEN fetching user information THEN the Repository SHALL first check the cache before querying the database
2. WHEN user data is updated THEN the Repository SHALL invalidate the corresponding cache entry
3. THE Cache_Layer SHALL support configurable TTL (Time To Live) for different data types
4. WHEN cache is unavailable THEN the Repository SHALL fall back to direct database queries without error
5. THE Cache_Layer SHALL support both in-memory caching and Redis for distributed deployments
6. WHEN listing proxies THEN the Proxy_Manager SHALL cache the results with a configurable TTL

### Requirement 5: 数据库优化

**User Story:** As a system administrator, I want optimized database operations, so that the system performs well under high load.

#### Acceptance Criteria

1. THE Database SHALL have indexes on frequently queried columns: users.username, proxies.user_id, proxies.protocol, traffic.user_id, traffic.recorded_at
2. WHEN listing resources THEN the Repository SHALL implement cursor-based pagination for large datasets
3. THE Database SHALL implement migration versioning to track applied migrations
4. WHEN querying traffic statistics THEN the Repository SHALL use aggregation queries instead of fetching all records
5. THE Database_Config SHALL support connection pool tuning with configurable max_open_conns, max_idle_conns, and conn_max_lifetime

### Requirement 6: API 文档生成

**User Story:** As a developer, I want auto-generated API documentation, so that API consumers can easily understand and use the endpoints.

#### Acceptance Criteria

1. THE V_Panel SHALL generate OpenAPI/Swagger specification from code annotations
2. WHEN the server starts THEN the V_Panel SHALL serve Swagger UI at /api/docs endpoint
3. THE API_Documentation SHALL include request/response schemas, authentication requirements, and example payloads
4. WHEN API endpoints change THEN the Documentation SHALL be automatically updated from code annotations

### Requirement 7: 前端状态管理完善

**User Story:** As a frontend developer, I want comprehensive state management, so that application state is predictable and maintainable.

#### Acceptance Criteria

1. THE Frontend_Store SHALL implement separate stores for: user, proxies, system, settings, and notifications
2. WHEN fetching data THEN the Frontend_Store SHALL track loading and error states
3. THE Frontend_Store SHALL persist critical state to sessionStorage for page refresh resilience
4. WHEN API calls fail THEN the Frontend_Store SHALL implement retry logic with exponential backoff
5. THE Frontend_Store SHALL implement optimistic updates for better user experience

### Requirement 8: 前端性能优化

**User Story:** As a user, I want fast page loads and smooth interactions, so that I can efficiently manage the proxy server.

#### Acceptance Criteria

1. THE Frontend SHALL implement route-based code splitting to reduce initial bundle size
2. WHEN displaying large lists THEN the Frontend SHALL implement virtual scrolling
3. THE Frontend SHALL lazy load components that are not immediately visible
4. WHEN making API requests THEN the Frontend SHALL implement request debouncing for search inputs
5. THE Frontend_Build SHALL tree-shake unused code and dependencies

### Requirement 8.1: 前端显示优化

**User Story:** As a user, I want a polished and responsive UI, so that I can have a pleasant experience while managing the proxy server.

#### Acceptance Criteria

1. WHEN data is loading THEN the Frontend SHALL display skeleton loading placeholders instead of blank screens
2. WHEN a page has no data THEN the Frontend SHALL display meaningful empty state illustrations with action suggestions
3. THE Frontend SHALL implement responsive design that works on desktop, tablet, and mobile devices
4. WHEN switching between dark and light themes THEN the Frontend SHALL apply theme changes smoothly without page reload
5. THE Frontend SHALL display data tables with sortable columns, filterable rows, and adjustable column widths
6. WHEN displaying statistics THEN the Frontend SHALL use animated charts with smooth transitions
7. THE Frontend SHALL implement breadcrumb navigation for deep page hierarchies
8. WHEN forms are submitted THEN the Frontend SHALL display inline validation feedback before submission
9. THE Frontend SHALL implement consistent spacing, typography, and color schemes across all pages
10. WHEN displaying proxy status THEN the Frontend SHALL use color-coded indicators (green for active, red for inactive, yellow for warning)

### Requirement 9: 测试覆盖率提升

**User Story:** As a developer, I want comprehensive test coverage, so that code changes can be made with confidence.

#### Acceptance Criteria

1. THE Backend SHALL have unit tests for all API handlers with minimum 80% code coverage
2. THE Backend SHALL have integration tests for database repository operations
3. THE Frontend SHALL have unit tests for all Pinia stores
4. THE Frontend SHALL have component tests for critical UI components
5. THE V_Panel SHALL have property-based tests for data serialization and parsing operations
6. WHEN running tests THEN the CI_Pipeline SHALL fail if coverage drops below threshold

### Requirement 10: 监控与可观测性

**User Story:** As a system administrator, I want comprehensive monitoring, so that I can identify and resolve issues quickly.

#### Acceptance Criteria

1. THE V_Panel SHALL expose Prometheus metrics at /metrics endpoint
2. THE Metrics SHALL include: request latency, request count, error rate, active connections, database query duration
3. WHEN errors occur THEN the V_Panel SHALL include correlation IDs for distributed tracing
4. THE V_Panel SHALL implement structured logging with consistent field names across all components
5. THE Health_Check SHALL verify database connectivity, Xray process status, and disk space availability

### Requirement 11: 配置验证增强

**User Story:** As a system administrator, I want configuration validation on startup, so that misconfigurations are caught early.

#### Acceptance Criteria

1. WHEN the application starts THEN the Config_Manager SHALL validate all required configuration values
2. IF required configuration is missing THEN the Config_Manager SHALL fail fast with a clear error message
3. THE Config_Manager SHALL validate JWT secret meets minimum length requirements
4. THE Config_Manager SHALL validate database connection string format
5. WHEN environment variables override config file values THEN the Config_Manager SHALL log which values were overridden

### Requirement 12: 优雅降级与容错

**User Story:** As a system administrator, I want the system to handle failures gracefully, so that partial failures don't cause complete system outages.

#### Acceptance Criteria

1. WHEN Xray process crashes THEN the V_Panel SHALL attempt automatic restart with exponential backoff
2. WHEN database connection is lost THEN the V_Panel SHALL retry connection with configurable retry policy
3. IF cache service is unavailable THEN the V_Panel SHALL continue operating with direct database access
4. WHEN external service calls timeout THEN the API_Handler SHALL return appropriate error responses without blocking
5. THE V_Panel SHALL implement circuit breaker pattern for external service calls

### Requirement 13: 前端异常错误处理与提示

**User Story:** As a user, I want clear and helpful error messages, so that I understand what went wrong and how to fix it.

#### Acceptance Criteria

1. WHEN an API request fails THEN the Frontend SHALL display a user-friendly error message with the error code and suggested action
2. WHEN network connection is lost THEN the Frontend SHALL display a persistent offline indicator with retry option
3. WHEN a form submission fails THEN the Frontend SHALL highlight the specific fields that caused the error
4. WHEN session expires THEN the Frontend SHALL display a modal prompting re-login without losing unsaved work
5. WHEN an unexpected JavaScript error occurs THEN the Frontend SHALL catch it globally and display a recovery option
6. THE Frontend SHALL implement toast notifications for transient errors and modal dialogs for critical errors
7. WHEN API returns validation errors THEN the Frontend SHALL map backend error codes to localized user messages
8. WHEN a long-running operation fails THEN the Frontend SHALL provide option to retry or cancel
9. THE Frontend SHALL log client-side errors to the backend for debugging purposes
10. WHEN displaying error messages THEN the Frontend SHALL include a unique error ID for support reference
11. IF multiple errors occur simultaneously THEN the Frontend SHALL queue and display them without overwhelming the user
12. WHEN an operation partially succeeds THEN the Frontend SHALL clearly indicate which parts succeeded and which failed


### Requirement 14: API 接口统一与规范化

**User Story:** As a frontend developer, I want a unified and well-organized API layer, so that API calls are consistent and maintainable.

#### Acceptance Criteria

1. THE Frontend_API SHALL consolidate duplicate API definitions into single modules (e.g., merge `users` and `usersApi` into one)
2. THE Frontend_API SHALL organize endpoints by domain: auth, users, proxies, system, certificates, backups, logs, stats, monitor
3. THE Frontend_API SHALL use consistent naming conventions: camelCase for methods, kebab-case for URL paths
4. WHEN defining API endpoints THEN the Frontend_API SHALL include TypeScript type definitions for request and response payloads
5. THE Frontend_API SHALL implement a centralized error handler that maps HTTP status codes to user-friendly messages
6. THE Frontend_API SHALL support request cancellation for long-running requests
7. WHEN making concurrent requests THEN the Frontend_API SHALL implement request deduplication to prevent duplicate calls
8. THE Frontend_API SHALL implement request queue for offline support with automatic retry when connection is restored
9. THE API_Base_URL SHALL be configurable through environment variables with fallback to relative path `/api`
10. THE Frontend_API SHALL log all API requests and responses in development mode for debugging

### Requirement 15: 数据库连接健壮性

**User Story:** As a system administrator, I want robust database connections, so that the system remains stable under various conditions.

#### Acceptance Criteria

1. THE Database SHALL support multiple drivers: SQLite, PostgreSQL, and MySQL
2. WHEN database connection fails THEN the Database SHALL implement automatic reconnection with exponential backoff
3. THE Database SHALL implement connection health checks at configurable intervals
4. WHEN connection pool is exhausted THEN the Database SHALL queue requests with configurable timeout
5. THE Database SHALL log slow queries exceeding configurable threshold (default 200ms)
6. WHEN database migration fails THEN the Database SHALL rollback changes and report detailed error
7. THE Database SHALL implement migration versioning to track applied migrations and prevent duplicate runs
8. THE Database_Config SHALL validate connection string format before attempting connection
9. WHEN database is unavailable at startup THEN the V_Panel SHALL retry connection with configurable max attempts before failing
10. THE Database SHALL implement read replica support for high-availability deployments
11. THE Database SHALL implement query timeout to prevent long-running queries from blocking connections
12. WHEN closing database connection THEN the Database SHALL gracefully drain active connections

### Requirement 16: 数据库索引优化

**User Story:** As a system administrator, I want optimized database queries, so that the system performs well with large datasets.

#### Acceptance Criteria

1. THE Database SHALL create index on `users.username` for fast user lookup
2. THE Database SHALL create index on `users.email` for email-based queries
3. THE Database SHALL create composite index on `proxies(user_id, enabled)` for user proxy listing
4. THE Database SHALL create index on `proxies.protocol` for protocol-based filtering
5. THE Database SHALL create composite index on `traffic(user_id, recorded_at)` for user traffic queries
6. THE Database SHALL create composite index on `traffic(proxy_id, recorded_at)` for proxy traffic queries
7. THE Database SHALL create index on `logs.created_at` for time-based log queries
8. THE Database SHALL create composite index on `logs(level, created_at)` for filtered log queries
9. WHEN running migrations THEN the Database SHALL create indexes if they do not exist
10. THE Database SHALL implement query analysis to identify missing indexes in production


### Requirement 17: 账号管理功能完善

**User Story:** As an administrator, I want complete user account management, so that I can effectively manage all users in the system.

#### Acceptance Criteria

1. THE API_Handler SHALL implement `POST /api/users/:id/enable` endpoint to enable a user account
2. THE API_Handler SHALL implement `POST /api/users/:id/disable` endpoint to disable a user account
3. WHEN a user account is disabled THEN the Auth_Service SHALL reject login attempts for that user
4. THE API_Handler SHALL implement `POST /api/users/:id/reset-password` endpoint to reset user password
5. WHEN password is reset THEN the Auth_Service SHALL generate a temporary password and require change on next login
6. THE User_Model SHALL include traffic_limit field to set maximum allowed traffic per user
7. THE User_Model SHALL include traffic_used field to track current traffic usage
8. THE User_Model SHALL include expires_at field to set account expiration date
9. WHEN user traffic exceeds limit THEN the V_Panel SHALL disable user's proxy access
10. WHEN user account expires THEN the V_Panel SHALL disable user's proxy access
11. THE API_Handler SHALL implement `GET /api/users/:id/login-history` endpoint to retrieve login history
12. THE V_Panel SHALL log all login attempts with timestamp, IP address, user agent, and success status
13. THE API_Handler SHALL implement `DELETE /api/users/:id/login-history` endpoint to clear login history
14. WHEN creating a user THEN the API_Handler SHALL validate email format if provided
15. WHEN updating a user THEN the API_Handler SHALL prevent changing username to an existing one

### Requirement 18: 系统设置功能实现

**User Story:** As an administrator, I want to configure system settings, so that I can customize the system behavior.

#### Acceptance Criteria

1. THE API_Handler SHALL implement `GET /api/settings` endpoint to retrieve all system settings
2. THE API_Handler SHALL implement `PUT /api/settings` endpoint to update system settings
3. THE Settings_Model SHALL be persisted to database instead of memory
4. THE Settings SHALL include: site_name, site_description, allow_registration, default_traffic_limit, default_expiry_days
5. THE Settings SHALL include: smtp_host, smtp_port, smtp_user, smtp_password for email notifications
6. THE Settings SHALL include: telegram_bot_token, telegram_chat_id for Telegram notifications
7. WHEN settings are updated THEN the V_Panel SHALL apply changes without restart
8. THE API_Handler SHALL implement `POST /api/settings/backup` endpoint to create settings backup
9. THE API_Handler SHALL implement `POST /api/settings/restore` endpoint to restore settings from backup
10. THE Settings SHALL include: xray_config_template for customizing Xray configuration
11. THE Settings SHALL include: rate_limit_enabled, rate_limit_requests, rate_limit_window for API rate limiting

### Requirement 19: 角色管理持久化

**User Story:** As an administrator, I want roles to persist across restarts, so that custom roles are not lost.

#### Acceptance Criteria

1. THE Role_Model SHALL be stored in database instead of memory
2. THE Role_Model SHALL include: id, name, description, permissions (JSON array), is_system, created_at, updated_at
3. WHEN the application starts THEN the V_Panel SHALL create default system roles if they don't exist
4. THE API_Handler SHALL prevent deletion of system roles (admin, user, viewer)
5. THE API_Handler SHALL prevent modification of system role permissions
6. WHEN a role is deleted THEN the V_Panel SHALL reassign affected users to default role
7. THE API_Handler SHALL validate that permission keys exist before assigning to role
8. THE Role_Handler SHALL implement permission inheritance (admin inherits all permissions)

### Requirement 20: 统计数据实时查询

**User Story:** As an administrator, I want accurate statistics, so that I can monitor system usage effectively.

#### Acceptance Criteria

1. THE Stats_Handler SHALL query actual data from database instead of returning placeholder values
2. THE Dashboard_Stats SHALL include: total_users (count from users table), active_users (users with recent activity)
3. THE Dashboard_Stats SHALL include: total_proxies (count from proxies table), active_proxies (enabled proxies)
4. THE Dashboard_Stats SHALL include: total_traffic, upload_traffic, download_traffic (sum from traffic table)
5. THE Protocol_Stats SHALL aggregate traffic by protocol type from traffic table
6. THE User_Stats SHALL show per-user traffic consumption with proxy count
7. THE Traffic_Stats SHALL support period filtering: today, week, month, year, custom range
8. THE Stats_Handler SHALL implement caching for expensive aggregation queries
9. THE Timeline_Stats SHALL return hourly/daily traffic data points for charts
10. THE Stats_Handler SHALL calculate online_count based on recent WebSocket connections or API activity

### Requirement 21: 代理服务功能完善

**User Story:** As a user, I want complete proxy management, so that I can effectively use and monitor my proxies.

#### Acceptance Criteria

1. THE Proxy_Model SHALL include user_id field to associate proxy with owner
2. WHEN creating a proxy THEN the Proxy_Handler SHALL set user_id to current authenticated user
3. WHEN listing proxies THEN the Proxy_Handler SHALL filter by user_id for non-admin users
4. THE API_Handler SHALL implement `POST /api/proxies/:id/start` endpoint to start a proxy
5. THE API_Handler SHALL implement `POST /api/proxies/:id/stop` endpoint to stop a proxy
6. WHEN starting a proxy THEN the Proxy_Manager SHALL update Xray configuration and reload
7. WHEN stopping a proxy THEN the Proxy_Manager SHALL remove from Xray configuration and reload
8. WHEN creating a proxy THEN the Proxy_Handler SHALL check for port conflicts
9. IF port is already in use THEN the Proxy_Handler SHALL return error with conflicting proxy info
10. THE API_Handler SHALL implement `GET /api/proxies/:id/stats` endpoint to get proxy traffic statistics
11. THE Proxy_Stats SHALL include: upload, download, total, connection_count, last_active
12. THE Proxy_Handler SHALL validate protocol-specific settings before creating/updating proxy
13. WHEN proxy is disabled THEN the Proxy_Manager SHALL remove it from active Xray configuration
14. THE Proxy_Handler SHALL implement batch operations: enable_all, disable_all, delete_selected

### Requirement 22: Xray 集成完善

**User Story:** As an administrator, I want seamless Xray integration, so that proxy configurations are automatically applied.

#### Acceptance Criteria

1. THE Xray_Manager SHALL monitor Xray process status and report to health check
2. WHEN Xray process crashes THEN the Xray_Manager SHALL attempt automatic restart
3. THE Xray_Manager SHALL implement `GET /api/xray/status` endpoint to get Xray process status
4. THE Xray_Manager SHALL implement `POST /api/xray/restart` endpoint to restart Xray process
5. THE Xray_Manager SHALL implement `GET /api/xray/config` endpoint to get current Xray configuration
6. THE Xray_Manager SHALL implement `PUT /api/xray/config` endpoint to update Xray configuration
7. WHEN proxy is created/updated/deleted THEN the Xray_Manager SHALL regenerate and reload configuration
8. THE Xray_Manager SHALL validate configuration before applying to prevent Xray crash
9. THE Xray_Manager SHALL implement `GET /api/xray/version` endpoint to get Xray version info
10. THE Xray_Manager SHALL implement `POST /api/xray/update` endpoint to update Xray to latest version
11. THE Xray_Manager SHALL backup current configuration before applying changes
12. IF configuration reload fails THEN the Xray_Manager SHALL rollback to previous configuration
