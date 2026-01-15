# Requirements Document

## Introduction

本功能为 V Panel 应用程序提供完整的订阅链接系统，允许用户生成专属订阅链接，支持多种主流客户端格式（Clash、V2rayN、Shadowrocket 等），并实现订阅内容的自动更新。该系统将使用户能够方便地将代理配置导入到各种客户端应用中。

## Glossary

- **Subscription_Service**: 订阅服务，负责生成和管理订阅链接
- **Subscription_Handler**: 订阅 API 处理器，负责处理订阅相关的 HTTP 请求
- **Subscription_Link**: 用户专属的订阅链接，包含唯一标识符
- **Subscription_Token**: 订阅令牌，用于验证订阅链接的有效性
- **Client_Format**: 客户端格式，如 Clash、V2rayN、Shadowrocket、Surge 等
- **Proxy_Config**: 代理配置，包含协议、服务器、端口等信息
- **System**: V Panel 应用程序系统
- **User**: 系统用户，拥有一个或多个代理配置

## Requirements

### Requirement 1: 订阅链接生成

**User Story:** As a user, I want to generate a unique subscription link, so that I can easily import my proxy configurations into various client applications.

#### Acceptance Criteria

1. WHEN a user requests a subscription link THEN the Subscription_Service SHALL generate a unique URL containing a secure token
2. THE Subscription_Link SHALL follow the format: `{base_url}/api/subscription/{token}`
3. THE Subscription_Token SHALL be a cryptographically secure random string of at least 32 characters
4. WHEN generating a subscription link THEN the Subscription_Service SHALL associate the token with the user's ID in the database
5. THE Subscription_Service SHALL support regenerating a new token to invalidate the old subscription link
6. WHEN a subscription link is regenerated THEN the Subscription_Service SHALL invalidate the previous token immediately
7. THE Subscription_Handler SHALL implement `GET /api/subscription/link` endpoint to retrieve the user's subscription link
8. THE Subscription_Handler SHALL implement `POST /api/subscription/regenerate` endpoint to regenerate the subscription token

### Requirement 2: 多客户端格式支持

**User Story:** As a user, I want my subscription to support multiple client formats, so that I can use my preferred VPN client application.

#### Acceptance Criteria

1. THE Subscription_Service SHALL support generating configurations in the following formats:
   - V2rayN/V2rayNG (base64 encoded vmess/vless/trojan/ss links)
   - Clash (YAML format)
   - Clash Meta (YAML format with extended features)
   - Shadowrocket (base64 encoded links)
   - Surge (Surge proxy list format)
   - Quantumult X (configuration format)
   - Sing-box (JSON format)
2. WHEN a client requests subscription content THEN the Subscription_Handler SHALL detect the client type from User-Agent header
3. THE Subscription_Handler SHALL support explicit format selection via query parameter `?format={format_name}`
4. WHEN format is not specified and User-Agent is unrecognized THEN the Subscription_Service SHALL default to V2rayN format
5. THE Subscription_Service SHALL generate valid configuration syntax for each supported client format
6. WHEN a proxy protocol is not supported by a client format THEN the Subscription_Service SHALL skip that proxy in the output

### Requirement 3: 订阅内容生成

**User Story:** As a user, I want my subscription to include all my active proxies, so that I can access all my configured servers.

#### Acceptance Criteria

1. WHEN generating subscription content THEN the Subscription_Service SHALL include only the user's enabled proxies
2. THE Subscription_Content SHALL include proxy name, server address, port, protocol type, and protocol-specific settings
3. WHEN a proxy has a custom remark/name THEN the Subscription_Service SHALL use it as the proxy name in the subscription
4. THE Subscription_Service SHALL generate unique proxy names to avoid conflicts in client applications
5. WHEN generating Clash format THEN the Subscription_Service SHALL include proxy groups for load balancing and fallback
6. THE Subscription_Content SHALL include the subscription update interval hint in the response headers
7. WHEN the user has no enabled proxies THEN the Subscription_Service SHALL return an empty but valid configuration

### Requirement 4: 订阅链接访问控制

**User Story:** As a system administrator, I want subscription links to be secure, so that unauthorized users cannot access proxy configurations.

#### Acceptance Criteria

1. WHEN an invalid or expired token is used THEN the Subscription_Handler SHALL return a 404 Not Found response
2. WHEN the user account is disabled THEN the Subscription_Handler SHALL return a 403 Forbidden response
3. WHEN the user's traffic limit is exceeded THEN the Subscription_Handler SHALL return a 403 Forbidden response with appropriate message
4. WHEN the user account has expired THEN the Subscription_Handler SHALL return a 403 Forbidden response
5. THE Subscription_Handler SHALL log all subscription access attempts with IP address and User-Agent
6. THE Subscription_Service SHALL support optional IP whitelist for subscription access
7. WHEN rate limiting is enabled THEN the Subscription_Handler SHALL limit requests to 60 per hour per token

### Requirement 5: 订阅信息管理界面

**User Story:** As a user, I want a web interface to manage my subscription, so that I can easily copy and share my subscription link.

#### Acceptance Criteria

1. WHEN a user visits the subscription page THEN the System SHALL display the user's subscription link with a copy button
2. THE Subscription_Page SHALL display a QR code for the subscription link for easy mobile scanning
3. THE Subscription_Page SHALL show the last access time and access count for the subscription
4. WHEN the user clicks regenerate THEN the System SHALL confirm the action before invalidating the old link
5. THE Subscription_Page SHALL provide format-specific links for each supported client type
6. THE Subscription_Page SHALL display instructions for importing the subscription into popular clients
7. WHEN displaying the subscription link THEN the System SHALL mask the token by default with an option to reveal

### Requirement 6: 订阅自动更新支持

**User Story:** As a user, I want my client to automatically update the subscription, so that I always have the latest proxy configurations.

#### Acceptance Criteria

1. THE Subscription_Handler SHALL include `Subscription-Userinfo` header with traffic usage information
2. THE Subscription_Handler SHALL include `Profile-Update-Interval` header to suggest update frequency (default 24 hours)
3. THE Subscription_Handler SHALL include `Content-Disposition` header with appropriate filename
4. WHEN proxy configurations change THEN the Subscription_Service SHALL update the subscription content immediately
5. THE Subscription_Handler SHALL support `If-Modified-Since` header to enable conditional requests
6. WHEN subscription content has not changed THEN the Subscription_Handler SHALL return 304 Not Modified
7. THE Subscription_Handler SHALL include `Profile-Title` header with the subscription name

### Requirement 7: 管理员订阅管理

**User Story:** As an administrator, I want to manage all users' subscriptions, so that I can monitor and control subscription usage.

#### Acceptance Criteria

1. THE Admin_Panel SHALL display a list of all subscription tokens with associated user information
2. THE Admin_Handler SHALL implement `GET /api/admin/subscriptions` endpoint to list all subscriptions
3. THE Admin_Handler SHALL implement `DELETE /api/admin/subscriptions/:user_id` endpoint to revoke a user's subscription
4. THE Admin_Panel SHALL show subscription access statistics including total requests and unique IPs
5. WHEN an administrator revokes a subscription THEN the System SHALL invalidate the token immediately
6. THE Admin_Panel SHALL allow filtering subscriptions by user, access count, and last access time
7. THE Admin_Handler SHALL implement `POST /api/admin/subscriptions/:user_id/reset-stats` endpoint to reset access statistics

### Requirement 8: 订阅链接短链接支持

**User Story:** As a user, I want shorter subscription links, so that they are easier to share and type.

#### Acceptance Criteria

1. THE Subscription_Service SHALL support generating short subscription links
2. THE Short_Link SHALL use a 8-character alphanumeric code instead of the full token
3. THE Subscription_Handler SHALL implement `GET /s/{short_code}` endpoint for short link access
4. WHEN a short link is accessed THEN the Subscription_Handler SHALL redirect or serve the subscription content directly
5. THE Subscription_Service SHALL maintain a mapping between short codes and full tokens
6. THE Short_Link SHALL be optional and can be enabled/disabled per user

### Requirement 9: 订阅配置自定义

**User Story:** As a user, I want to customize my subscription output, so that I can tailor it to my specific needs.

#### Acceptance Criteria

1. THE Subscription_Service SHALL support filtering proxies by protocol type in the subscription output
2. THE Subscription_Handler SHALL support `?protocols=vmess,vless` query parameter to filter protocols
3. THE Subscription_Service SHALL support custom proxy naming templates
4. THE Subscription_Handler SHALL support `?rename={template}` query parameter for custom naming
5. THE Subscription_Service SHALL support excluding specific proxies from the subscription
6. WHEN generating Clash format THEN the Subscription_Service SHALL support custom proxy group configurations
7. THE Subscription_Handler SHALL support `?include={proxy_ids}` and `?exclude={proxy_ids}` query parameters

### Requirement 10: 订阅数据持久化

**User Story:** As a developer, I want subscription data to be properly stored, so that the system maintains subscription state across restarts.

#### Acceptance Criteria

1. THE Subscription_Model SHALL be stored in the database with fields: id, user_id, token, short_code, created_at, updated_at, last_access_at, access_count
2. THE Database SHALL create index on `subscriptions.token` for fast token lookup
3. THE Database SHALL create index on `subscriptions.short_code` for fast short link lookup
4. THE Database SHALL create unique constraint on `subscriptions.user_id` (one subscription per user)
5. WHEN a user is deleted THEN the Database SHALL cascade delete the associated subscription record
6. THE Subscription_Repository SHALL implement methods for CRUD operations on subscription records

