# Requirements Document

## Introduction

本功能为 V Panel 应用程序提供完整的用户前台门户系统（User Portal），与现有的管理后台分离，为普通用户提供友好的界面来管理个人账户、查看节点、获取订阅、下载客户端等功能。用户前台将采用独立的路由和界面风格，支持响应式设计以适配移动端访问。

## Glossary

- **User_Portal**: 用户前台门户系统，面向普通用户的 Web 界面
- **Admin_Panel**: 管理后台，面向管理员的 Web 界面
- **User_Dashboard**: 用户仪表板，展示用户核心信息的页面
- **Node**: 代理节点，用户可用的代理服务器
- **Subscription**: 订阅，用户的代理配置订阅链接
- **Client_App**: 客户端应用，用于连接代理的软件
- **Announcement**: 系统公告，管理员发布的通知信息
- **Ticket**: 工单，用户提交的支持请求
- **System**: V Panel 应用程序系统
- **User**: 普通用户，拥有账户和代理配置的终端用户

## Requirements

### Requirement 1: 用户注册系统

**User Story:** As a new user, I want to register an account, so that I can access the proxy service.

#### Acceptance Criteria

1. WHEN a visitor accesses the registration page THEN the System SHALL display a registration form with username, email, and password fields
2. THE Registration_Form SHALL validate email format before submission
3. THE Registration_Form SHALL enforce password strength requirements (minimum 8 characters, at least one letter and one number)
4. WHEN registration is submitted THEN the System SHALL check for duplicate username and email
5. IF username or email already exists THEN the System SHALL display a specific error message
6. WHEN invite code is required THEN the Registration_Form SHALL include an invite code field
7. IF invite code is invalid or expired THEN the System SHALL reject the registration with appropriate message
8. WHEN registration succeeds THEN the System SHALL send a verification email to the user
9. THE System SHALL support optional CAPTCHA verification to prevent automated registrations
10. WHEN email verification is enabled THEN the System SHALL require email confirmation before allowing login

### Requirement 2: 用户登录系统

**User Story:** As a registered user, I want to log in to my account, so that I can access my proxy configurations.

#### Acceptance Criteria

1. WHEN a user accesses the login page THEN the System SHALL display a login form with username/email and password fields
2. WHEN login credentials are valid THEN the System SHALL authenticate the user and redirect to the dashboard
3. IF login credentials are invalid THEN the System SHALL display a generic error message without revealing which field is incorrect
4. THE System SHALL implement rate limiting with maximum 5 login attempts per 15 minutes per IP
5. WHEN rate limit is exceeded THEN the System SHALL display a lockout message with remaining time
6. THE Login_Page SHALL provide a "Forgot Password" link
7. WHEN "Remember Me" is checked THEN the System SHALL extend the session duration to 30 days
8. THE System SHALL support optional two-factor authentication (2FA) via TOTP
9. WHEN 2FA is enabled THEN the System SHALL prompt for verification code after password validation
10. THE System SHALL log all login attempts with timestamp, IP address, and success status

### Requirement 3: 密码重置功能

**User Story:** As a user who forgot my password, I want to reset it, so that I can regain access to my account.

#### Acceptance Criteria

1. WHEN a user requests password reset THEN the System SHALL send a reset link to the registered email
2. THE Reset_Link SHALL expire after 1 hour
3. THE Reset_Link SHALL be single-use and invalidated after successful password change
4. WHEN setting a new password THEN the System SHALL enforce the same password strength requirements as registration
5. WHEN password is successfully reset THEN the System SHALL invalidate all existing sessions for that user
6. THE System SHALL rate limit password reset requests to 3 per hour per email

### Requirement 4: 用户仪表板

**User Story:** As a logged-in user, I want to see my account overview, so that I can quickly understand my account status.

#### Acceptance Criteria

1. WHEN a user visits the dashboard THEN the System SHALL display user's basic information (username, email, account status)
2. THE Dashboard SHALL display traffic usage with visual progress bar (used/total, percentage)
3. THE Dashboard SHALL display account expiration date with countdown if expiring within 7 days
4. THE Dashboard SHALL display traffic reset date and countdown
5. WHEN traffic usage exceeds 80% THEN the Dashboard SHALL display a warning indicator
6. WHEN account is about to expire (within 7 days) THEN the Dashboard SHALL display an expiration warning
7. THE Dashboard SHALL display quick action buttons (copy subscription, download client, view nodes)
8. THE Dashboard SHALL display recent system announcements (latest 3)
9. THE Dashboard SHALL display online device count if IP limiting is enabled
10. WHEN account status is abnormal (disabled, expired, traffic exceeded) THEN the Dashboard SHALL display a prominent alert with explanation

### Requirement 5: 节点列表页面

**User Story:** As a user, I want to view available proxy nodes, so that I can choose the best server for my needs.

#### Acceptance Criteria

1. WHEN a user visits the nodes page THEN the System SHALL display a list of available nodes
2. THE Node_List SHALL display node name, location/region, protocol type, and status (online/offline)
3. THE Node_List SHALL support filtering by region and protocol type
4. THE Node_List SHALL support sorting by name, region, or latency
5. WHEN a node is offline THEN the System SHALL display it with a distinct visual indicator
6. THE System SHALL display node load percentage if available
7. THE Node_Page SHALL provide a "Test Latency" button to measure ping to each node
8. WHEN latency test is performed THEN the System SHALL display results in milliseconds with color coding (green < 100ms, yellow < 300ms, red > 300ms)
9. THE Node_List SHALL display the number of available nodes in the header
10. WHEN user clicks on a node THEN the System SHALL display node details including server address, port, and protocol settings

### Requirement 6: 订阅管理页面

**User Story:** As a user, I want to manage my subscription link, so that I can easily configure my proxy clients.

#### Acceptance Criteria

1. WHEN a user visits the subscription page THEN the System SHALL display the user's subscription link
2. THE Subscription_Page SHALL provide a one-click copy button for the subscription link
3. THE Subscription_Page SHALL display a QR code for the subscription link
4. THE Subscription_Page SHALL provide format-specific links for different clients (Clash, V2rayN, Shadowrocket, etc.)
5. WHEN user clicks "Reset Subscription" THEN the System SHALL confirm the action and generate a new subscription token
6. THE Subscription_Page SHALL display subscription access statistics (total access count, last access time)
7. THE Subscription_Page SHALL provide "One-Click Import" buttons that open client apps with the subscription URL
8. THE Subscription_Page SHALL display instructions for importing subscription into popular clients
9. WHEN subscription link is copied THEN the System SHALL display a success toast notification
10. THE Subscription_Page SHALL support generating temporary subscription links with expiration

### Requirement 7: 客户端下载页面

**User Story:** As a user, I want to download proxy client applications, so that I can connect to the proxy service.

#### Acceptance Criteria

1. WHEN a user visits the download page THEN the System SHALL display client applications grouped by platform
2. THE Download_Page SHALL display clients for Windows, macOS, Linux, Android, and iOS
3. FOR each client THE System SHALL display: name, version, download link, and brief description
4. THE Download_Page SHALL highlight recommended clients for each platform
5. THE Download_Page SHALL provide direct download links and links to official repositories
6. FOR iOS clients THE System SHALL note that they require App Store purchase
7. THE Download_Page SHALL display client compatibility information (supported protocols)
8. THE Download_Page SHALL provide links to setup tutorials for each client
9. THE Download_Page SHALL detect user's operating system and highlight relevant clients
10. WHEN a download link is clicked THEN the System SHALL track download statistics

### Requirement 8: 个人设置页面

**User Story:** As a user, I want to manage my account settings, so that I can update my information and preferences.

#### Acceptance Criteria

1. WHEN a user visits settings page THEN the System SHALL display sections for profile, security, and notifications
2. THE Profile_Section SHALL allow updating display name and avatar
3. THE Security_Section SHALL allow changing password with current password verification
4. THE Security_Section SHALL allow enabling/disabling two-factor authentication
5. WHEN enabling 2FA THEN the System SHALL display QR code for TOTP setup and backup codes
6. THE Security_Section SHALL display active sessions with option to revoke
7. THE Notification_Section SHALL allow configuring email notification preferences
8. THE Notification_Section SHALL allow binding Telegram account for notifications
9. THE Settings_Page SHALL allow switching between light and dark themes
10. THE Settings_Page SHALL allow selecting preferred language (if i18n is supported)

### Requirement 9: 公告通知中心

**User Story:** As a user, I want to view system announcements, so that I can stay informed about service updates.

#### Acceptance Criteria

1. WHEN a user visits the announcements page THEN the System SHALL display a list of announcements sorted by date
2. THE Announcement_List SHALL display title, publish date, and preview text for each announcement
3. WHEN user clicks an announcement THEN the System SHALL display the full content
4. THE System SHALL mark announcements as read/unread for each user
5. THE Dashboard SHALL display unread announcement count as a badge
6. THE System SHALL support pinned announcements that appear at the top
7. THE Announcement_Page SHALL support pagination for large announcement lists
8. WHEN a new announcement is published THEN the System SHALL notify users via configured channels
9. THE System SHALL support announcement categories (maintenance, update, promotion)
10. THE Announcement_Detail SHALL support markdown formatting

### Requirement 10: 工单支持系统

**User Story:** As a user, I want to submit support tickets, so that I can get help with issues.

#### Acceptance Criteria

1. WHEN a user visits the ticket page THEN the System SHALL display a list of user's tickets
2. THE Ticket_List SHALL display ticket ID, subject, status, and last update time
3. THE System SHALL support ticket statuses: open, pending, resolved, closed
4. WHEN user creates a new ticket THEN the System SHALL require subject and description
5. THE Ticket_Form SHALL support file attachments (images, logs) with size limit
6. WHEN a ticket is created THEN the System SHALL assign a unique ticket ID
7. THE Ticket_Detail SHALL display conversation thread between user and support
8. WHEN user replies to a ticket THEN the System SHALL update ticket status to "pending"
9. WHEN support replies THEN the System SHALL notify user via email/Telegram
10. THE System SHALL allow users to close their own tickets
11. THE Ticket_Page SHALL display estimated response time based on current queue

### Requirement 11: 使用统计页面

**User Story:** As a user, I want to view my usage statistics, so that I can monitor my traffic consumption.

#### Acceptance Criteria

1. WHEN a user visits the statistics page THEN the System SHALL display traffic usage charts
2. THE Statistics_Page SHALL display daily traffic usage for the past 30 days
3. THE Statistics_Page SHALL display traffic breakdown by upload and download
4. THE Statistics_Page SHALL display traffic usage by node/protocol
5. THE Charts SHALL support switching between daily, weekly, and monthly views
6. THE Statistics_Page SHALL display peak usage times
7. THE Statistics_Page SHALL display total traffic used in current billing cycle
8. THE Statistics_Page SHALL display traffic remaining until reset
9. WHEN hovering over chart data points THEN the System SHALL display detailed values
10. THE Statistics_Page SHALL allow exporting usage data as CSV

### Requirement 12: 知识库帮助中心

**User Story:** As a user, I want to access help documentation, so that I can learn how to use the service.

#### Acceptance Criteria

1. WHEN a user visits the help center THEN the System SHALL display categorized help articles
2. THE Help_Center SHALL include categories: Getting Started, Client Setup, Troubleshooting, FAQ
3. THE Help_Center SHALL provide a search function to find relevant articles
4. THE Article_Page SHALL support markdown formatting with images and code blocks
5. THE Help_Center SHALL display popular/featured articles on the main page
6. THE Article_Page SHALL display related articles at the bottom
7. THE Help_Center SHALL support video tutorials embedded from YouTube/Bilibili
8. THE Article_Page SHALL display last updated date
9. THE Help_Center SHALL allow users to rate article helpfulness
10. THE Help_Center SHALL be accessible without login for basic articles

### Requirement 13: 用户前台路由和布局

**User Story:** As a developer, I want the user portal to have a separate routing structure, so that it is independent from the admin panel.

#### Acceptance Criteria

1. THE User_Portal SHALL use a separate route prefix `/user` or subdomain
2. THE User_Portal SHALL have its own layout component distinct from admin panel
3. THE Layout SHALL include a responsive navigation menu
4. THE Layout SHALL include a user dropdown with quick links (settings, logout)
5. THE Layout SHALL support mobile-responsive design with hamburger menu
6. THE Layout SHALL include a footer with links to help, terms, and contact
7. WHEN user is not authenticated THEN the System SHALL redirect to login page
8. THE Navigation SHALL highlight the current active page
9. THE Layout SHALL support dark/light theme switching
10. THE User_Portal SHALL lazy-load page components for performance

### Requirement 14: 移动端适配

**User Story:** As a mobile user, I want to access the portal on my phone, so that I can manage my account anywhere.

#### Acceptance Criteria

1. THE User_Portal SHALL be fully responsive and usable on mobile devices
2. THE Mobile_Layout SHALL use a bottom navigation bar for main sections
3. THE Subscription_Page SHALL display QR code at a scannable size on mobile
4. THE Node_List SHALL use a card layout on mobile for better touch interaction
5. THE Forms SHALL use appropriate mobile input types (email, tel, etc.)
6. THE Buttons SHALL have minimum touch target size of 44x44 pixels
7. THE Mobile_Layout SHALL support pull-to-refresh on list pages
8. THE Charts SHALL be touch-friendly with pinch-to-zoom support
9. THE Mobile_Layout SHALL minimize horizontal scrolling
10. THE System SHALL detect mobile devices and suggest native app download

