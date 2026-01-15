# Requirements Document

## Introduction

本功能为 V Panel 应用程序提供完整的 IP 限制系统，用于控制用户同时在线设备数量、限制订阅链接访问来源、以及实现基于 IP 的安全策略。该系统将帮助运营者防止账号共享滥用，提升服务安全性。

## Glossary

- **IP_Restriction**: IP 限制，基于 IP 地址的访问控制策略
- **Concurrent_IP**: 并发 IP，同一时间段内使用服务的不同 IP 地址数量
- **IP_Whitelist**: IP 白名单，允许访问的 IP 地址列表
- **IP_Blacklist**: IP 黑名单，禁止访问的 IP 地址列表
- **Online_Device**: 在线设备，当前正在使用代理服务的客户端
- **IP_Record**: IP 记录，用户访问时的 IP 地址历史
- **System**: V Panel 应用程序系统
- **User**: 使用代理服务的终端用户
- **Admin**: 系统管理员

## Requirements

### Requirement 1: 并发 IP 限制

**User Story:** As an admin, I want to limit the number of concurrent IPs per user, so that I can prevent account sharing abuse.

#### Acceptance Criteria

1. THE User_Model SHALL include max_concurrent_ips field to set maximum allowed concurrent IPs per user
2. THE Plan SHALL support setting default max_concurrent_ips for users subscribing to that plan
3. WHEN a user connects from a new IP THEN the System SHALL check against the concurrent IP limit
4. IF concurrent IP limit is exceeded THEN the System SHALL reject the new connection with appropriate error message
5. THE System SHALL track active IPs per user with last activity timestamp
6. WHEN an IP has been inactive for configurable duration (default 10 minutes) THEN the System SHALL remove it from active count
7. THE Admin_Panel SHALL allow overriding concurrent IP limit for individual users
8. THE System SHALL support unlimited concurrent IPs option (value: 0 or -1)


### Requirement 2: IP 活动追踪

**User Story:** As an admin, I want to track user IP activities, so that I can monitor usage patterns and detect abuse.

#### Acceptance Criteria

1. WHEN a user accesses subscription link THEN the System SHALL record the IP address with timestamp
2. WHEN a user connects to proxy THEN the System SHALL record the connection IP via Xray API
3. THE IP_Record SHALL include: user_id, ip_address, user_agent, access_type, timestamp
4. THE System SHALL aggregate IP records to show unique IPs per day/week/month
5. THE Admin_Panel SHALL display IP activity history for each user
6. THE System SHALL support IP geolocation lookup to show country/region
7. THE IP_History SHALL be retained for configurable period (default 30 days)
8. THE System SHALL detect and flag suspicious patterns (e.g., IPs from multiple countries in short time)

### Requirement 3: 在线设备管理

**User Story:** As a user, I want to see my online devices, so that I can manage my active sessions.

#### Acceptance Criteria

1. THE User_Portal SHALL display list of currently online devices/IPs
2. THE Device_List SHALL show: IP address, location (country/city), last active time, device type (if detectable)
3. THE User_Portal SHALL allow user to disconnect/kick specific devices
4. WHEN user kicks a device THEN the System SHALL add that IP to temporary block list
5. THE User_Portal SHALL display remaining available device slots
6. WHEN max devices reached THEN the User_Portal SHALL show which device to disconnect to add new one
7. THE System SHALL send notification when new device connects (optional, configurable)

### Requirement 4: IP 白名单

**User Story:** As an admin, I want to configure IP whitelists, so that trusted IPs can bypass restrictions.

#### Acceptance Criteria

1. THE Admin_Panel SHALL provide interface to manage global IP whitelist
2. THE IP_Whitelist SHALL support individual IPs and CIDR ranges
3. WHEN an IP is whitelisted THEN the System SHALL bypass concurrent IP checks for that IP
4. THE System SHALL support per-user IP whitelist configuration
5. THE Admin_Panel SHALL allow importing IP whitelist from file (one IP per line)
6. THE IP_Whitelist SHALL support adding description/label for each entry
7. WHEN whitelisted IP is used THEN the System SHALL still log the access for audit

### Requirement 5: IP 黑名单

**User Story:** As an admin, I want to block specific IPs, so that I can prevent malicious access.

#### Acceptance Criteria

1. THE Admin_Panel SHALL provide interface to manage global IP blacklist
2. THE IP_Blacklist SHALL support individual IPs and CIDR ranges
3. WHEN a blacklisted IP attempts access THEN the System SHALL reject with 403 Forbidden
4. THE System SHALL support automatic blacklisting based on rules (e.g., too many failed attempts)
5. THE IP_Blacklist SHALL support temporary blocks with expiration time
6. THE Admin_Panel SHALL show reason and timestamp for each blacklist entry
7. THE System SHALL support per-user IP blacklist (block specific IPs for specific user)
8. WHEN IP is blacklisted THEN the System SHALL log the block event


### Requirement 6: 订阅链接 IP 限制

**User Story:** As an admin, I want to restrict subscription link access by IP, so that subscription links cannot be shared publicly.

#### Acceptance Criteria

1. THE Subscription_Service SHALL support IP-based access restriction
2. THE Admin_Panel SHALL allow configuring max unique IPs that can access a subscription link
3. WHEN subscription link IP limit is exceeded THEN the System SHALL return 403 with message to contact support
4. THE System SHALL track unique IPs that have accessed each subscription link
5. THE User_Portal SHALL display IPs that have accessed user's subscription link
6. THE User_Portal SHALL allow user to reset subscription IP access list (with regenerate token)
7. THE Admin_Panel SHALL allow clearing subscription IP access list for specific user

### Requirement 7: 地理位置限制

**User Story:** As an admin, I want to restrict access by geographic location, so that I can comply with regional policies.

#### Acceptance Criteria

1. THE Admin_Panel SHALL allow configuring allowed/blocked countries
2. WHEN country restriction is enabled THEN the System SHALL check IP geolocation before allowing access
3. IF IP is from blocked country THEN the System SHALL reject access with appropriate message
4. THE System SHALL use reliable IP geolocation database (e.g., MaxMind GeoLite2)
5. THE Admin_Panel SHALL support per-plan country restrictions
6. THE System SHALL cache geolocation lookups to improve performance
7. THE Admin_Panel SHALL display access statistics by country

### Requirement 8: IP 限制配置界面

**User Story:** As an admin, I want a comprehensive interface to manage IP restrictions, so that I can easily configure policies.

#### Acceptance Criteria

1. THE Admin_Panel SHALL provide dedicated IP restriction settings page
2. THE Settings_Page SHALL allow configuring global concurrent IP limit default
3. THE Settings_Page SHALL allow configuring IP inactive timeout duration
4. THE Settings_Page SHALL allow enabling/disabling IP restriction features
5. THE Settings_Page SHALL display current IP restriction statistics
6. THE Admin_Panel SHALL provide bulk operations for IP whitelist/blacklist
7. THE Settings_Page SHALL allow configuring automatic blacklist rules
8. THE Admin_Panel SHALL support exporting IP restriction configuration

### Requirement 9: IP 限制 API

**User Story:** As a developer, I want IP restriction APIs, so that I can integrate with external systems.

#### Acceptance Criteria

1. THE API_Handler SHALL implement `GET /api/admin/ip-restrictions/stats` endpoint for statistics
2. THE API_Handler SHALL implement `GET /api/admin/users/:id/online-ips` endpoint to get user's online IPs
3. THE API_Handler SHALL implement `POST /api/admin/users/:id/kick-ip` endpoint to disconnect specific IP
4. THE API_Handler SHALL implement CRUD endpoints for whitelist management
5. THE API_Handler SHALL implement CRUD endpoints for blacklist management
6. THE API_Handler SHALL implement `GET /api/user/devices` endpoint for user to view own devices
7. THE API_Handler SHALL implement `POST /api/user/devices/:ip/kick` endpoint for user to kick own device

### Requirement 10: IP 限制通知

**User Story:** As a user, I want to be notified about IP-related events, so that I can be aware of my account security.

#### Acceptance Criteria

1. WHEN new device connects THEN the System SHALL optionally send notification to user
2. WHEN IP limit is reached THEN the System SHALL notify user with suggestion to disconnect old devices
3. WHEN suspicious IP activity is detected THEN the System SHALL alert admin
4. THE Notification SHALL include: IP address, location, timestamp, and action taken
5. THE User_Settings SHALL allow configuring IP notification preferences
6. THE Admin_Panel SHALL display IP-related alerts in notification center
7. THE System SHALL support notification via email and Telegram (if configured)
