# Design Document: User Portal

## Overview

用户前台门户（User Portal）是 V Panel 面向普通用户的 Web 界面，与管理后台（Admin Panel）完全分离。采用 Vue 3 + Element Plus 技术栈，支持响应式设计和移动端适配，提供用户友好的界面来管理账户、查看节点、获取订阅等功能。

### Key Design Decisions

1. **路由分离**: 用户前台使用 `/user` 路由前缀，与管理后台 `/admin` 完全分离
2. **独立布局**: 用户前台使用简洁友好的布局，区别于管理后台的专业风格
3. **API 复用**: 复用现有后端 API，新增用户端专用接口
4. **响应式优先**: 采用移动优先的响应式设计策略
5. **渐进式加载**: 使用路由懒加载和组件懒加载优化性能

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         User Portal (Vue 3)                             │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                        Router (/user/*)                           │ │
│  │  /login | /register | /dashboard | /nodes | /subscription | ...  │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                  │                                      │
│  ┌───────────────────────────────┴───────────────────────────────────┐ │
│  │                         Layouts                                    │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐               │ │
│  │  │ AuthLayout  │  │ UserLayout  │  │ MobileLayout│               │ │
│  │  │ (登录/注册) │  │ (主布局)    │  │ (移动端)   │               │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘               │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                  │                                      │
│  ┌───────────────────────────────┴───────────────────────────────────┐ │
│  │                          Views                                     │ │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐│ │
│  │  │Dashboard │ │  Nodes   │ │Subscript │ │ Download │ │ Settings ││ │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘│ │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐            │ │
│  │  │Announce  │ │ Tickets  │ │  Stats   │ │HelpCenter│            │ │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘            │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                  │                                      │
│  ┌───────────────────────────────┴───────────────────────────────────┐ │
│  │                       Pinia Stores                                 │ │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐         │ │
│  │  │  user  │ │ nodes  │ │ ticket │ │announce│ │ stats  │         │ │
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘         │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                  │                                      │
│  ┌───────────────────────────────┴───────────────────────────────────┐ │
│  │                         API Layer                                  │ │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐         │ │
│  │  │authApi │ │nodesApi│ │ticketApi│ │statsApi│ │helpApi │         │ │
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘         │ │
│  └───────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Backend API (Go/Gin)                            │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                    User Portal Handlers                           │ │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐    │ │
│  │  │PortalAuth  │ │PortalNode  │ │PortalTicket│ │PortalStats │    │ │
│  │  │  Handler   │ │  Handler   │ │  Handler   │ │  Handler   │    │ │
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘    │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                   │                                     │
│  ┌───────────────────────────────┴───────────────────────────────────┐ │
│  │                       Services (Existing + New)                   │ │
│  │  AuthService | ProxyService | TicketService | AnnouncementService │ │
│  └───────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Frontend Router Structure

```typescript
// router/user.ts - User Portal Routes
const userRoutes = [
  {
    path: '/user',
    component: UserLayout,
    children: [
      { path: '', redirect: '/user/dashboard' },
      { path: 'dashboard', component: () => import('@/views/user/Dashboard.vue') },
      { path: 'nodes', component: () => import('@/views/user/Nodes.vue') },
      { path: 'subscription', component: () => import('@/views/user/Subscription.vue') },
      { path: 'download', component: () => import('@/views/user/Download.vue') },
      { path: 'settings', component: () => import('@/views/user/Settings.vue') },
      { path: 'announcements', component: () => import('@/views/user/Announcements.vue') },
      { path: 'tickets', component: () => import('@/views/user/Tickets.vue') },
      { path: 'tickets/:id', component: () => import('@/views/user/TicketDetail.vue') },
      { path: 'stats', component: () => import('@/views/user/Stats.vue') },
      { path: 'help', component: () => import('@/views/user/HelpCenter.vue') },
      { path: 'help/:slug', component: () => import('@/views/user/HelpArticle.vue') },
    ],
    meta: { requiresAuth: true }
  },
  {
    path: '/user/login',
    component: AuthLayout,
    children: [
      { path: '', component: () => import('@/views/user/Login.vue') },
    ]
  },
  {
    path: '/user/register',
    component: AuthLayout,
    children: [
      { path: '', component: () => import('@/views/user/Register.vue') },
    ]
  },
  {
    path: '/user/forgot-password',
    component: AuthLayout,
    children: [
      { path: '', component: () => import('@/views/user/ForgotPassword.vue') },
    ]
  },
]
```

### 2. Backend API Handlers

```go
// internal/api/handlers/portal_auth.go
type PortalAuthHandler struct {
    authService    *auth.Service
    userRepo       repository.UserRepository
    emailService   *email.Service
    logger         logger.Logger
}

func (h *PortalAuthHandler) Register(c *gin.Context)        // POST /api/portal/auth/register
func (h *PortalAuthHandler) Login(c *gin.Context)           // POST /api/portal/auth/login
func (h *PortalAuthHandler) Logout(c *gin.Context)          // POST /api/portal/auth/logout
func (h *PortalAuthHandler) ForgotPassword(c *gin.Context)  // POST /api/portal/auth/forgot-password
func (h *PortalAuthHandler) ResetPassword(c *gin.Context)   // POST /api/portal/auth/reset-password
func (h *PortalAuthHandler) VerifyEmail(c *gin.Context)     // GET /api/portal/auth/verify-email
func (h *PortalAuthHandler) GetProfile(c *gin.Context)      // GET /api/portal/auth/profile
func (h *PortalAuthHandler) UpdateProfile(c *gin.Context)   // PUT /api/portal/auth/profile
func (h *PortalAuthHandler) ChangePassword(c *gin.Context)  // PUT /api/portal/auth/password
func (h *PortalAuthHandler) Enable2FA(c *gin.Context)       // POST /api/portal/auth/2fa/enable
func (h *PortalAuthHandler) Verify2FA(c *gin.Context)       // POST /api/portal/auth/2fa/verify
func (h *PortalAuthHandler) Disable2FA(c *gin.Context)      // POST /api/portal/auth/2fa/disable
```

```go
// internal/api/handlers/portal_node.go
type PortalNodeHandler struct {
    proxyRepo   repository.ProxyRepository
    userRepo    repository.UserRepository
    logger      logger.Logger
}

func (h *PortalNodeHandler) ListNodes(c *gin.Context)       // GET /api/portal/nodes
func (h *PortalNodeHandler) GetNode(c *gin.Context)         // GET /api/portal/nodes/:id
func (h *PortalNodeHandler) TestLatency(c *gin.Context)     // POST /api/portal/nodes/:id/ping
```

```go
// internal/api/handlers/portal_ticket.go
type PortalTicketHandler struct {
    ticketService *ticket.Service
    logger        logger.Logger
}

func (h *PortalTicketHandler) ListTickets(c *gin.Context)   // GET /api/portal/tickets
func (h *PortalTicketHandler) CreateTicket(c *gin.Context)  // POST /api/portal/tickets
func (h *PortalTicketHandler) GetTicket(c *gin.Context)     // GET /api/portal/tickets/:id
func (h *PortalTicketHandler) ReplyTicket(c *gin.Context)   // POST /api/portal/tickets/:id/reply
func (h *PortalTicketHandler) CloseTicket(c *gin.Context)   // POST /api/portal/tickets/:id/close
```

### 3. Pinia Stores

```typescript
// stores/userPortal.ts
interface UserPortalState {
  user: UserProfile | null
  isAuthenticated: boolean
  loading: boolean
  error: string | null
}

interface UserProfile {
  id: number
  username: string
  email: string
  status: 'active' | 'disabled' | 'expired'
  trafficUsed: number
  trafficLimit: number
  expiresAt: string | null
  trafficResetAt: string | null
  twoFactorEnabled: boolean
}

// stores/nodes.ts
interface NodesState {
  nodes: Node[]
  loading: boolean
  error: string | null
  latencyResults: Record<number, number>
}

interface Node {
  id: number
  name: string
  region: string
  protocol: string
  status: 'online' | 'offline' | 'maintenance'
  load: number
}

// stores/tickets.ts
interface TicketsState {
  tickets: Ticket[]
  currentTicket: TicketDetail | null
  loading: boolean
  error: string | null
}

interface Ticket {
  id: number
  subject: string
  status: 'open' | 'pending' | 'resolved' | 'closed'
  createdAt: string
  updatedAt: string
}
```

## Data Models

### New Database Models

```go
// Ticket model for support system
type Ticket struct {
    ID          uint      `gorm:"primaryKey"`
    UserID      uint      `gorm:"index;not null"`
    Subject     string    `gorm:"size:256;not null"`
    Status      string    `gorm:"size:32;default:'open'"` // open, pending, resolved, closed
    Priority    string    `gorm:"size:32;default:'normal'"` // low, normal, high, urgent
    CreatedAt   time.Time
    UpdatedAt   time.Time
    ClosedAt    *time.Time
    
    User        *User          `gorm:"foreignKey:UserID"`
    Messages    []TicketMessage `gorm:"foreignKey:TicketID"`
}

// TicketMessage model for ticket conversations
type TicketMessage struct {
    ID          uint      `gorm:"primaryKey"`
    TicketID    uint      `gorm:"index;not null"`
    UserID      uint      `gorm:"index"` // null for system/admin messages
    Content     string    `gorm:"type:text;not null"`
    IsAdmin     bool      `gorm:"default:false"`
    Attachments string    `gorm:"type:text"` // JSON array of attachment URLs
    CreatedAt   time.Time
    
    Ticket      *Ticket   `gorm:"foreignKey:TicketID"`
    User        *User     `gorm:"foreignKey:UserID"`
}

// Announcement model for system announcements
type Announcement struct {
    ID          uint      `gorm:"primaryKey"`
    Title       string    `gorm:"size:256;not null"`
    Content     string    `gorm:"type:text;not null"`
    Category    string    `gorm:"size:64;default:'general'"` // general, maintenance, update, promotion
    IsPinned    bool      `gorm:"default:false"`
    IsPublished bool      `gorm:"default:false"`
    PublishedAt *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// AnnouncementRead tracks read status per user
type AnnouncementRead struct {
    ID             uint      `gorm:"primaryKey"`
    UserID         uint      `gorm:"index;not null"`
    AnnouncementID uint      `gorm:"index;not null"`
    ReadAt         time.Time
    
    User         *User         `gorm:"foreignKey:UserID"`
    Announcement *Announcement `gorm:"foreignKey:AnnouncementID"`
}

// HelpArticle model for knowledge base
type HelpArticle struct {
    ID          uint      `gorm:"primaryKey"`
    Slug        string    `gorm:"uniqueIndex;size:128;not null"`
    Title       string    `gorm:"size:256;not null"`
    Content     string    `gorm:"type:text;not null"`
    Category    string    `gorm:"size:64;index"`
    Tags        string    `gorm:"size:512"` // JSON array
    ViewCount   int64     `gorm:"default:0"`
    HelpfulCount int64    `gorm:"default:0"`
    IsPublished bool      `gorm:"default:false"`
    IsFeatured  bool      `gorm:"default:false"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// PasswordResetToken for password reset functionality
type PasswordResetToken struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"index;not null"`
    Token     string    `gorm:"uniqueIndex;size:64;not null"`
    ExpiresAt time.Time `gorm:"not null"`
    UsedAt    *time.Time
    CreatedAt time.Time
    
    User      *User     `gorm:"foreignKey:UserID"`
}

// EmailVerificationToken for email verification
type EmailVerificationToken struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"index;not null"`
    Email     string    `gorm:"size:256;not null"`
    Token     string    `gorm:"uniqueIndex;size:64;not null"`
    ExpiresAt time.Time `gorm:"not null"`
    VerifiedAt *time.Time
    CreatedAt time.Time
    
    User      *User     `gorm:"foreignKey:UserID"`
}

// InviteCode for invitation system
type InviteCode struct {
    ID          uint      `gorm:"primaryKey"`
    Code        string    `gorm:"uniqueIndex;size:32;not null"`
    CreatedBy   uint      `gorm:"index"` // Admin or user who created
    UsedBy      *uint     `gorm:"index"` // User who used the code
    MaxUses     int       `gorm:"default:1"`
    UsedCount   int       `gorm:"default:0"`
    ExpiresAt   *time.Time
    CreatedAt   time.Time
    UsedAt      *time.Time
}

// TwoFactorSecret for 2FA
type TwoFactorSecret struct {
    ID          uint      `gorm:"primaryKey"`
    UserID      uint      `gorm:"uniqueIndex;not null"`
    Secret      string    `gorm:"size:64;not null"` // Encrypted TOTP secret
    BackupCodes string    `gorm:"type:text"` // JSON array of hashed backup codes
    EnabledAt   time.Time
    
    User        *User     `gorm:"foreignKey:UserID"`
}
```

### User Model Extensions

```go
// Add to existing User model
type User struct {
    // ... existing fields ...
    
    // New fields for user portal
    EmailVerified     bool       `gorm:"default:false"`
    EmailVerifiedAt   *time.Time
    TwoFactorEnabled  bool       `gorm:"default:false"`
    LastLoginAt       *time.Time
    LastLoginIP       string     `gorm:"size:45"`
    AvatarURL         string     `gorm:"size:512"`
    DisplayName       string     `gorm:"size:64"`
    TelegramID        *int64     `gorm:"index"`
    NotifyEmail       bool       `gorm:"default:true"`
    NotifyTelegram    bool       `gorm:"default:false"`
    Theme             string     `gorm:"size:16;default:'auto'"` // auto, light, dark
    Language          string     `gorm:"size:8;default:'zh-CN'"`
    InvitedBy         *uint      `gorm:"index"`
}
```

## API Endpoints Summary

### Portal Authentication

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/portal/auth/register` | User registration |
| POST | `/api/portal/auth/login` | User login |
| POST | `/api/portal/auth/logout` | User logout |
| POST | `/api/portal/auth/forgot-password` | Request password reset |
| POST | `/api/portal/auth/reset-password` | Reset password with token |
| GET | `/api/portal/auth/verify-email` | Verify email address |
| GET | `/api/portal/auth/profile` | Get user profile |
| PUT | `/api/portal/auth/profile` | Update user profile |
| PUT | `/api/portal/auth/password` | Change password |
| POST | `/api/portal/auth/2fa/enable` | Enable 2FA |
| POST | `/api/portal/auth/2fa/verify` | Verify 2FA code |
| POST | `/api/portal/auth/2fa/disable` | Disable 2FA |

### Portal Dashboard

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/portal/dashboard` | Get dashboard data |
| GET | `/api/portal/dashboard/traffic` | Get traffic summary |
| GET | `/api/portal/dashboard/announcements` | Get recent announcements |

### Portal Nodes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/portal/nodes` | List available nodes |
| GET | `/api/portal/nodes/:id` | Get node details |
| POST | `/api/portal/nodes/:id/ping` | Test node latency |

### Portal Tickets

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/portal/tickets` | List user's tickets |
| POST | `/api/portal/tickets` | Create new ticket |
| GET | `/api/portal/tickets/:id` | Get ticket details |
| POST | `/api/portal/tickets/:id/reply` | Reply to ticket |
| POST | `/api/portal/tickets/:id/close` | Close ticket |

### Portal Announcements

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/portal/announcements` | List announcements |
| GET | `/api/portal/announcements/:id` | Get announcement detail |
| POST | `/api/portal/announcements/:id/read` | Mark as read |

### Portal Statistics

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/portal/stats/traffic` | Get traffic statistics |
| GET | `/api/portal/stats/usage` | Get usage by node/protocol |
| GET | `/api/portal/stats/export` | Export statistics as CSV |

### Portal Help Center

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/portal/help/articles` | List help articles |
| GET | `/api/portal/help/articles/:slug` | Get article by slug |
| GET | `/api/portal/help/search` | Search articles |
| POST | `/api/portal/help/articles/:slug/helpful` | Rate article |



## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Email Format Validation
*For any* string input, the email validation function SHALL correctly identify valid RFC 5322 compliant email addresses and reject invalid ones.
**Validates: Requirements 1.2**

### Property 2: Password Strength Validation
*For any* password string, the validation SHALL accept only passwords with minimum 8 characters containing at least one letter and one number.
**Validates: Requirements 1.3**

### Property 3: Username/Email Uniqueness
*For any* registration attempt with an existing username or email, the system SHALL reject the registration with an appropriate error.
**Validates: Requirements 1.4, 1.5**

### Property 4: Login Rate Limiting
*For any* IP address, after 5 failed login attempts within 15 minutes, subsequent login attempts SHALL be blocked until the cooldown period expires.
**Validates: Requirements 2.4, 2.5**

### Property 5: Password Reset Token Expiration
*For any* password reset token older than 1 hour, the system SHALL reject the reset attempt.
**Validates: Requirements 3.2**

### Property 6: Password Reset Token Single-Use
*For any* password reset token that has been used once, subsequent reset attempts with the same token SHALL fail.
**Validates: Requirements 3.3**

### Property 7: Session Invalidation on Password Reset
*For any* successful password reset, all existing sessions for that user SHALL be invalidated.
**Validates: Requirements 3.5**

### Property 8: Node List Filtering Correctness
*For any* filter criteria (region, protocol), the returned node list SHALL contain only nodes matching all specified criteria.
**Validates: Requirements 5.3**

### Property 9: Node List Sorting Correctness
*For any* sort criteria (name, region, latency), the returned node list SHALL be correctly ordered according to the specified criteria.
**Validates: Requirements 5.4**

### Property 10: Announcement Read Status Tracking
*For any* announcement marked as read by a user, subsequent queries SHALL return that announcement with read status true for that user.
**Validates: Requirements 9.4**

### Property 11: Ticket ID Uniqueness
*For any* two tickets in the system, their ticket IDs SHALL be unique.
**Validates: Requirements 10.6**

### Property 12: Ticket Status Transitions
*For any* ticket, status transitions SHALL follow the valid state machine: open → pending ↔ resolved → closed.
**Validates: Requirements 10.3**

### Property 13: Traffic Statistics Consistency
*For any* time period, the sum of daily traffic values SHALL equal the total traffic for that period.
**Validates: Requirements 11.2, 11.3**

### Property 14: Help Article Search Relevance
*For any* search query, returned articles SHALL contain the search terms in their title, content, or tags.
**Validates: Requirements 12.3**

### Property 15: 2FA Token Validation
*For any* valid TOTP token, verification SHALL succeed only within the valid time window (±30 seconds).
**Validates: Requirements 2.8, 2.9**

### Property 16: Invite Code Usage Limit
*For any* invite code with max_uses limit, the code SHALL be rejected after reaching the usage limit.
**Validates: Requirements 1.6, 1.7**

## Error Handling

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_CREDENTIALS` | 401 | Invalid username/password |
| `ACCOUNT_DISABLED` | 403 | User account is disabled |
| `ACCOUNT_EXPIRED` | 403 | User account has expired |
| `EMAIL_NOT_VERIFIED` | 403 | Email not verified |
| `RATE_LIMITED` | 429 | Too many requests |
| `INVALID_TOKEN` | 400 | Invalid or expired token |
| `TOKEN_USED` | 400 | Token already used |
| `DUPLICATE_USERNAME` | 409 | Username already exists |
| `DUPLICATE_EMAIL` | 409 | Email already exists |
| `INVALID_INVITE_CODE` | 400 | Invalid or expired invite code |
| `INVALID_2FA_CODE` | 400 | Invalid 2FA verification code |
| `TICKET_CLOSED` | 400 | Cannot reply to closed ticket |
| `ATTACHMENT_TOO_LARGE` | 413 | File attachment exceeds size limit |

### Error Response Format

```json
{
  "code": "INVALID_CREDENTIALS",
  "message": "用户名或密码错误",
  "details": {}
}
```

## Testing Strategy

### Unit Tests

Unit tests will cover:
- Email and password validation functions
- Token generation and validation
- Rate limiting logic
- Node filtering and sorting
- Ticket status transitions
- Statistics aggregation

### Property-Based Tests

Property-based tests will use Go's `testing/quick` package or `gopter` library to verify:
- Input validation correctness (email, password)
- Rate limiting behavior
- Token expiration and single-use
- Data filtering and sorting
- State machine transitions

**Configuration**: Each property test will run minimum 100 iterations.

**Test Annotation Format**: Each test will be tagged with:
```go
// Feature: user-portal, Property N: Property description
// Validates: Requirements X.Y
```

### Integration Tests

Integration tests will cover:
- Full registration flow
- Login with 2FA
- Password reset flow
- Ticket creation and reply flow
- Announcement read tracking

### Frontend Tests

Frontend tests will cover:
- Component rendering
- Form validation
- API integration
- Responsive layout

### Test Files Structure

```
internal/portal/
├── auth/
│   ├── service.go
│   ├── service_test.go
│   └── service_property_test.go
├── ticket/
│   ├── service.go
│   ├── service_test.go
│   └── service_property_test.go
├── announcement/
│   ├── service.go
│   └── service_test.go
└── help/
    ├── service.go
    └── service_test.go

web/src/views/user/
├── __tests__/
│   ├── Login.spec.ts
│   ├── Register.spec.ts
│   ├── Dashboard.spec.ts
│   └── Nodes.spec.ts
```

## Frontend Component Structure

```
web/src/
├── views/user/
│   ├── Login.vue
│   ├── Register.vue
│   ├── ForgotPassword.vue
│   ├── ResetPassword.vue
│   ├── Dashboard.vue
│   ├── Nodes.vue
│   ├── Subscription.vue
│   ├── Download.vue
│   ├── Settings.vue
│   ├── Announcements.vue
│   ├── AnnouncementDetail.vue
│   ├── Tickets.vue
│   ├── TicketDetail.vue
│   ├── TicketCreate.vue
│   ├── Stats.vue
│   ├── HelpCenter.vue
│   └── HelpArticle.vue
├── layouts/
│   ├── UserLayout.vue
│   ├── AuthLayout.vue
│   └── MobileLayout.vue
├── components/user/
│   ├── UserNavbar.vue
│   ├── UserSidebar.vue
│   ├── UserFooter.vue
│   ├── TrafficCard.vue
│   ├── NodeCard.vue
│   ├── TicketStatusBadge.vue
│   ├── AnnouncementCard.vue
│   └── TrafficChart.vue
├── stores/
│   ├── userPortal.ts
│   ├── nodes.ts
│   ├── tickets.ts
│   ├── announcements.ts
│   └── stats.ts
└── api/
    ├── portalAuth.ts
    ├── portalNodes.ts
    ├── portalTickets.ts
    ├── portalAnnouncements.ts
    ├── portalStats.ts
    └── portalHelp.ts
```

## Database Schema Additions

```sql
-- Tickets table
CREATE TABLE tickets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    subject VARCHAR(256) NOT NULL,
    status VARCHAR(32) DEFAULT 'open',
    priority VARCHAR(32) DEFAULT 'normal',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_tickets_user_id ON tickets(user_id);
CREATE INDEX idx_tickets_status ON tickets(status);

-- Ticket messages table
CREATE TABLE ticket_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ticket_id INTEGER NOT NULL,
    user_id INTEGER,
    content TEXT NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    attachments TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_ticket_messages_ticket_id ON ticket_messages(ticket_id);

-- Announcements table
CREATE TABLE announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(256) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(64) DEFAULT 'general',
    is_pinned BOOLEAN DEFAULT FALSE,
    is_published BOOLEAN DEFAULT FALSE,
    published_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_announcements_published ON announcements(is_published, published_at);

-- Announcement read status table
CREATE TABLE announcement_reads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    announcement_id INTEGER NOT NULL,
    read_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (announcement_id) REFERENCES announcements(id) ON DELETE CASCADE,
    UNIQUE(user_id, announcement_id)
);

-- Help articles table
CREATE TABLE help_articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug VARCHAR(128) NOT NULL UNIQUE,
    title VARCHAR(256) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(64),
    tags VARCHAR(512),
    view_count INTEGER DEFAULT 0,
    helpful_count INTEGER DEFAULT 0,
    is_published BOOLEAN DEFAULT FALSE,
    is_featured BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_help_articles_slug ON help_articles(slug);
CREATE INDEX idx_help_articles_category ON help_articles(category);

-- Password reset tokens table
CREATE TABLE password_reset_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    used_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);

-- Email verification tokens table
CREATE TABLE email_verification_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    email VARCHAR(256) NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    verified_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_email_verification_tokens_token ON email_verification_tokens(token);

-- Invite codes table
CREATE TABLE invite_codes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code VARCHAR(32) NOT NULL UNIQUE,
    created_by INTEGER,
    used_by INTEGER,
    max_uses INTEGER DEFAULT 1,
    used_count INTEGER DEFAULT 0,
    expires_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at DATETIME,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (used_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_invite_codes_code ON invite_codes(code);

-- Two-factor secrets table
CREATE TABLE two_factor_secrets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    secret VARCHAR(64) NOT NULL,
    backup_codes TEXT,
    enabled_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- User table extensions (add columns)
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN email_verified_at DATETIME;
ALTER TABLE users ADD COLUMN two_factor_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN last_login_at DATETIME;
ALTER TABLE users ADD COLUMN last_login_ip VARCHAR(45);
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(512);
ALTER TABLE users ADD COLUMN display_name VARCHAR(64);
ALTER TABLE users ADD COLUMN telegram_id BIGINT;
ALTER TABLE users ADD COLUMN notify_email BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN notify_telegram BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN theme VARCHAR(16) DEFAULT 'auto';
ALTER TABLE users ADD COLUMN language VARCHAR(8) DEFAULT 'zh-CN';
ALTER TABLE users ADD COLUMN invited_by INTEGER;
```
