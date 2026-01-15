# Design Document: Subscription System

## Overview

订阅链接系统为 V Panel 提供用户专属订阅链接功能，支持多种主流客户端格式。系统采用令牌认证机制，通过 RESTful API 提供订阅内容，并支持自动更新和访问控制。

### Key Design Decisions

1. **令牌机制**: 使用 32 字符的加密安全随机令牌，存储在数据库中与用户关联
2. **格式检测**: 优先使用 User-Agent 自动检测客户端类型，支持显式 format 参数覆盖
3. **内容生成**: 采用策略模式，每种客户端格式实现独立的生成器
4. **缓存策略**: 订阅内容支持条件请求（If-Modified-Since），减少不必要的数据传输

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (Vue.js)                        │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ SubscriptionPage│  │  QRCodeDisplay  │  │  LinkCopyButton │ │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘ │
└───────────┼─────────────────────┼─────────────────────┼─────────┘
            │                     │                     │
            ▼                     ▼                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                         API Layer (Gin)                         │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                   SubscriptionHandler                       ││
│  │  - GetLink()      - GetContent()     - Regenerate()        ││
│  │  - GetShortLink() - AdminList()      - AdminRevoke()       ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Service Layer                              │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                  SubscriptionService                        ││
│  │  - GenerateToken()    - ValidateToken()                     ││
│  │  - GenerateContent()  - DetectClientFormat()                ││
│  │  - RegenerateToken()  - GetSubscriptionInfo()               ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                   FormatGenerators                          ││
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       ││
│  │  │ V2rayN   │ │  Clash   │ │Shadowrkt │ │  Surge   │       ││
│  │  │Generator │ │Generator │ │Generator │ │Generator │       ││
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘       ││
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐                    ││
│  │  │QuantumX  │ │ Sing-box │ │ClashMeta │                    ││
│  │  │Generator │ │Generator │ │Generator │                    ││
│  │  └──────────┘ └──────────┘ └──────────┘                    ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Repository Layer                            │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │               SubscriptionRepository                        ││
│  │  - Create()    - GetByToken()    - GetByUserID()           ││
│  │  - Update()    - Delete()        - GetByShortCode()        ││
│  │  - UpdateAccessStats()           - ListAll()               ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Database (SQLite)                        │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    subscriptions table                      ││
│  │  id | user_id | token | short_code | created_at | ...      ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. SubscriptionHandler (API Layer)

```go
// SubscriptionHandler handles subscription-related HTTP requests
type SubscriptionHandler struct {
    service *SubscriptionService
    logger  logger.Logger
}

// Public endpoints (no auth required, token-based)
func (h *SubscriptionHandler) GetContent(c *gin.Context)      // GET /api/subscription/:token
func (h *SubscriptionHandler) GetShortContent(c *gin.Context) // GET /s/:short_code

// Protected endpoints (require authentication)
func (h *SubscriptionHandler) GetLink(c *gin.Context)         // GET /api/subscription/link
func (h *SubscriptionHandler) Regenerate(c *gin.Context)      // POST /api/subscription/regenerate
func (h *SubscriptionHandler) GetInfo(c *gin.Context)         // GET /api/subscription/info

// Admin endpoints (require admin role)
func (h *SubscriptionHandler) AdminList(c *gin.Context)       // GET /api/admin/subscriptions
func (h *SubscriptionHandler) AdminRevoke(c *gin.Context)     // DELETE /api/admin/subscriptions/:user_id
func (h *SubscriptionHandler) AdminResetStats(c *gin.Context) // POST /api/admin/subscriptions/:user_id/reset-stats
```

### 2. SubscriptionService (Business Logic)

```go
// SubscriptionService provides subscription business logic
type SubscriptionService struct {
    repo        SubscriptionRepository
    userRepo    UserRepository
    proxyRepo   ProxyRepository
    generators  map[ClientFormat]FormatGenerator
    logger      logger.Logger
}

// Token management
func (s *SubscriptionService) GenerateToken() (string, error)
func (s *SubscriptionService) ValidateToken(token string) (*Subscription, error)
func (s *SubscriptionService) RegenerateToken(userID uint) (*Subscription, error)

// Content generation
func (s *SubscriptionService) GenerateContent(userID uint, format ClientFormat, options *ContentOptions) ([]byte, error)
func (s *SubscriptionService) DetectClientFormat(userAgent string) ClientFormat

// Subscription management
func (s *SubscriptionService) GetOrCreateSubscription(userID uint) (*Subscription, error)
func (s *SubscriptionService) GetSubscriptionInfo(userID uint) (*SubscriptionInfo, error)
func (s *SubscriptionService) UpdateAccessStats(subscriptionID uint, ip string, userAgent string) error
```

### 3. FormatGenerator Interface

```go
// FormatGenerator defines the interface for subscription format generators
type FormatGenerator interface {
    // Generate creates subscription content for the specific format
    Generate(proxies []*Proxy, options *GeneratorOptions) ([]byte, error)
    
    // ContentType returns the MIME type for the generated content
    ContentType() string
    
    // FileExtension returns the file extension for downloads
    FileExtension() string
    
    // SupportsProtocol checks if the format supports a specific protocol
    SupportsProtocol(protocol string) bool
}

// Implementations
type V2rayNGenerator struct{}      // Base64 encoded links
type ClashGenerator struct{}       // YAML format
type ClashMetaGenerator struct{}   // Extended YAML format
type ShadowrocketGenerator struct{} // Base64 encoded links
type SurgeGenerator struct{}       // Surge proxy list
type QuantumultXGenerator struct{} // Quantumult X format
type SingboxGenerator struct{}     // JSON format
```

### 4. SubscriptionRepository (Data Access)

```go
// SubscriptionRepository defines data access methods for subscriptions
type SubscriptionRepository interface {
    Create(subscription *Subscription) error
    GetByID(id uint) (*Subscription, error)
    GetByToken(token string) (*Subscription, error)
    GetByShortCode(shortCode string) (*Subscription, error)
    GetByUserID(userID uint) (*Subscription, error)
    Update(subscription *Subscription) error
    Delete(id uint) error
    UpdateAccessStats(id uint, ip string, userAgent string) error
    ListAll(filter *SubscriptionFilter) ([]*Subscription, int64, error)
}
```

## Data Models

### Subscription Model

```go
// Subscription represents a user's subscription record
type Subscription struct {
    ID           uint      `gorm:"primaryKey"`
    UserID       uint      `gorm:"uniqueIndex;not null"`
    Token        string    `gorm:"uniqueIndex;size:64;not null"`
    ShortCode    string    `gorm:"uniqueIndex;size:16"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    LastAccessAt *time.Time
    AccessCount  int64     `gorm:"default:0"`
    LastIP       string    `gorm:"size:45"`
    LastUA       string    `gorm:"size:256"`
    
    // Relations
    User         *User     `gorm:"foreignKey:UserID"`
}

// SubscriptionInfo represents subscription information for display
type SubscriptionInfo struct {
    Link         string    `json:"link"`
    ShortLink    string    `json:"short_link,omitempty"`
    Token        string    `json:"token"`
    CreatedAt    time.Time `json:"created_at"`
    LastAccessAt *time.Time `json:"last_access_at,omitempty"`
    AccessCount  int64     `json:"access_count"`
    QRCodeData   string    `json:"qr_code_data"`
    Formats      []FormatInfo `json:"formats"`
}

// FormatInfo represents information about a supported format
type FormatInfo struct {
    Name        string `json:"name"`
    DisplayName string `json:"display_name"`
    Link        string `json:"link"`
    Icon        string `json:"icon,omitempty"`
}
```

### Client Format Enum

```go
// ClientFormat represents supported subscription client formats
type ClientFormat string

const (
    FormatV2rayN      ClientFormat = "v2rayn"
    FormatClash       ClientFormat = "clash"
    FormatClashMeta   ClientFormat = "clashmeta"
    FormatShadowrocket ClientFormat = "shadowrocket"
    FormatSurge       ClientFormat = "surge"
    FormatQuantumultX ClientFormat = "quantumultx"
    FormatSingbox     ClientFormat = "singbox"
    FormatAuto        ClientFormat = "auto"
)

// ContentOptions represents options for content generation
type ContentOptions struct {
    Protocols    []string // Filter by protocols
    Include      []uint   // Include specific proxy IDs
    Exclude      []uint   // Exclude specific proxy IDs
    RenameTemplate string // Custom naming template
}
```

### API Response Models

```go
// SubscriptionLinkResponse represents the response for getting subscription link
type SubscriptionLinkResponse struct {
    Link      string       `json:"link"`
    ShortLink string       `json:"short_link,omitempty"`
    Formats   []FormatInfo `json:"formats"`
    Stats     *AccessStats `json:"stats"`
}

// AccessStats represents subscription access statistics
type AccessStats struct {
    TotalAccess  int64      `json:"total_access"`
    LastAccessAt *time.Time `json:"last_access_at,omitempty"`
    LastIP       string     `json:"last_ip,omitempty"`
}

// AdminSubscriptionListResponse represents admin subscription list response
type AdminSubscriptionListResponse struct {
    Subscriptions []*AdminSubscriptionItem `json:"subscriptions"`
    Total         int64                    `json:"total"`
    Page          int                      `json:"page"`
    PageSize      int                      `json:"page_size"`
}

// AdminSubscriptionItem represents a subscription item in admin list
type AdminSubscriptionItem struct {
    ID           uint       `json:"id"`
    UserID       uint       `json:"user_id"`
    Username     string     `json:"username"`
    Token        string     `json:"token"`
    ShortCode    string     `json:"short_code,omitempty"`
    CreatedAt    time.Time  `json:"created_at"`
    LastAccessAt *time.Time `json:"last_access_at,omitempty"`
    AccessCount  int64      `json:"access_count"`
    LastIP       string     `json:"last_ip,omitempty"`
}
```

## Format Generation Details

### V2rayN Format (Base64)

```
vmess://base64(json_config)
vless://uuid@server:port?params#name
trojan://password@server:port?params#name
ss://base64(method:password)@server:port#name
```

### Clash Format (YAML)

```yaml
port: 7890
socks-port: 7891
allow-lan: false
mode: Rule
log-level: info

proxies:
  - name: "Proxy Name"
    type: vmess
    server: server.example.com
    port: 443
    uuid: uuid-here
    alterId: 0
    cipher: auto
    tls: true

proxy-groups:
  - name: "Proxy"
    type: select
    proxies:
      - "Proxy Name"

rules:
  - MATCH,Proxy
```

### Sing-box Format (JSON)

```json
{
  "outbounds": [
    {
      "type": "vmess",
      "tag": "proxy-name",
      "server": "server.example.com",
      "server_port": 443,
      "uuid": "uuid-here",
      "security": "auto",
      "alter_id": 0,
      "tls": {
        "enabled": true
      }
    }
  ]
}
```



## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Token Uniqueness
*For any* two subscription tokens generated by the system, they SHALL never be equal.
**Validates: Requirements 1.1**

### Property 2: Token Length Constraint
*For any* generated subscription token, its length SHALL be at least 32 characters.
**Validates: Requirements 1.3**

### Property 3: Token-User Association Round Trip
*For any* user who generates a subscription token, querying the subscription by that token SHALL return the same user ID.
**Validates: Requirements 1.4**

### Property 4: Token Regeneration Invalidation
*For any* subscription, after regenerating the token, the old token SHALL be invalid (return not found) and the new token SHALL be valid.
**Validates: Requirements 1.5, 1.6**

### Property 5: Format Detection Consistency
*For any* known User-Agent string pattern, the detected client format SHALL be consistent across multiple calls.
**Validates: Requirements 2.2**

### Property 6: Format Override Priority
*For any* request with an explicit format parameter, the format parameter SHALL override User-Agent detection.
**Validates: Requirements 2.3**

### Property 7: Clash Configuration Round Trip
*For any* valid set of proxies, generating Clash YAML and parsing it back SHALL produce equivalent proxy configurations.
**Validates: Requirements 2.5**

### Property 8: Sing-box Configuration Round Trip
*For any* valid set of proxies, generating Sing-box JSON and parsing it back SHALL produce equivalent proxy configurations.
**Validates: Requirements 2.5**

### Property 9: Enabled Proxies Only
*For any* subscription content generation, the output SHALL contain only proxies that are enabled (no disabled proxies).
**Validates: Requirements 3.1**

### Property 10: Required Fields Presence
*For any* proxy in the subscription content, it SHALL contain name, server, port, and protocol type.
**Validates: Requirements 3.2**

### Property 11: Unique Proxy Names
*For any* subscription content, all proxy names within the content SHALL be unique.
**Validates: Requirements 3.4**

### Property 12: Invalid Token Returns 404
*For any* token that does not exist in the database, accessing the subscription SHALL return HTTP 404.
**Validates: Requirements 4.1**

### Property 13: Disabled User Access Denied
*For any* user whose account is disabled, accessing their subscription SHALL return HTTP 403.
**Validates: Requirements 4.2**

### Property 14: Traffic Exceeded Access Denied
*For any* user whose traffic usage exceeds their limit, accessing their subscription SHALL return HTTP 403.
**Validates: Requirements 4.3**

### Property 15: Expired User Access Denied
*For any* user whose account has expired, accessing their subscription SHALL return HTTP 403.
**Validates: Requirements 4.4**

### Property 16: Response Headers Presence
*For any* successful subscription content response, it SHALL include Subscription-Userinfo, Profile-Update-Interval, and Content-Disposition headers.
**Validates: Requirements 6.1, 6.2, 6.3, 6.7**

### Property 17: Content Reflects Current State
*For any* proxy configuration change, the subscription content SHALL immediately reflect the updated state.
**Validates: Requirements 6.4**

### Property 18: Short Code Length
*For any* generated short code, its length SHALL be exactly 8 characters.
**Validates: Requirements 8.2**

### Property 19: Short Code Mapping Consistency
*For any* short code, looking up the subscription by short code SHALL return the same subscription as looking up by the full token.
**Validates: Requirements 8.5**

### Property 20: Protocol Filter Correctness
*For any* subscription request with protocol filter, the output SHALL contain only proxies matching the specified protocols.
**Validates: Requirements 9.2**

### Property 21: User Subscription Uniqueness
*For any* user, there SHALL be at most one subscription record in the database.
**Validates: Requirements 10.4**

### Property 22: Cascade Delete
*For any* user deletion, the associated subscription record SHALL also be deleted.
**Validates: Requirements 10.5**

## Error Handling

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `SUBSCRIPTION_NOT_FOUND` | 404 | Token or short code not found |
| `USER_DISABLED` | 403 | User account is disabled |
| `USER_EXPIRED` | 403 | User account has expired |
| `TRAFFIC_EXCEEDED` | 403 | User traffic limit exceeded |
| `RATE_LIMITED` | 429 | Too many requests |
| `INVALID_FORMAT` | 400 | Invalid format parameter |
| `GENERATION_FAILED` | 500 | Failed to generate subscription content |

### Error Response Format

```json
{
  "code": "SUBSCRIPTION_NOT_FOUND",
  "message": "The subscription link is invalid or has been revoked",
  "details": {}
}
```

### Error Handling Strategy

1. **Token Validation Errors**: Return 404 to avoid leaking information about token existence
2. **Access Control Errors**: Return 403 with specific error code for client handling
3. **Rate Limiting**: Return 429 with Retry-After header
4. **Generation Errors**: Log full error, return 500 with generic message

## Testing Strategy

### Unit Tests

Unit tests will cover:
- Token generation and validation logic
- Format detection from User-Agent strings
- Individual format generators (V2rayN, Clash, Sing-box, etc.)
- Access control checks (disabled user, expired user, traffic exceeded)
- Short code generation and lookup

### Property-Based Tests

Property-based tests will use Go's `testing/quick` package or `gopter` library to verify:
- Token uniqueness across many generations
- Format generator output validity (round-trip parsing)
- Access control invariants
- Filter correctness

**Configuration**: Each property test will run minimum 100 iterations.

**Test Annotation Format**: Each test will be tagged with:
```go
// Feature: subscription-system, Property N: Property description
// Validates: Requirements X.Y
```

### Integration Tests

Integration tests will cover:
- Full API endpoint flows
- Database operations (CRUD)
- Subscription content generation with real proxy data
- Access control middleware integration

### Test Files Structure

```
internal/subscription/
├── service.go
├── service_test.go          # Unit tests
├── service_property_test.go # Property-based tests
├── handler.go
├── handler_test.go          # Handler unit tests
├── repository.go
├── repository_test.go       # Repository integration tests
└── generators/
    ├── v2rayn.go
    ├── v2rayn_test.go
    ├── clash.go
    ├── clash_test.go
    ├── singbox.go
    └── singbox_test.go
```

## API Endpoints Summary

### Public Endpoints (Token-based auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/subscription/:token` | Get subscription content |
| GET | `/s/:short_code` | Get subscription content via short link |

### Protected Endpoints (JWT auth required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/subscription/link` | Get user's subscription link |
| GET | `/api/subscription/info` | Get subscription info with stats |
| POST | `/api/subscription/regenerate` | Regenerate subscription token |

### Admin Endpoints (Admin role required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/subscriptions` | List all subscriptions |
| DELETE | `/api/admin/subscriptions/:user_id` | Revoke user's subscription |
| POST | `/api/admin/subscriptions/:user_id/reset-stats` | Reset access statistics |

## Database Schema

```sql
CREATE TABLE subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    token VARCHAR(64) NOT NULL UNIQUE,
    short_code VARCHAR(16) UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_access_at DATETIME,
    access_count INTEGER NOT NULL DEFAULT 0,
    last_ip VARCHAR(45),
    last_ua VARCHAR(256),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_subscriptions_token ON subscriptions(token);
CREATE INDEX idx_subscriptions_short_code ON subscriptions(short_code);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
```

## Frontend Components

### SubscriptionPage.vue

Main subscription management page with:
- Subscription link display with copy button
- QR code generation
- Format-specific links
- Access statistics
- Regenerate button with confirmation

### Components

```
web/src/views/Subscription.vue       # Main subscription page
web/src/components/
├── subscription/
│   ├── SubscriptionLink.vue         # Link display with copy
│   ├── SubscriptionQRCode.vue       # QR code display
│   ├── SubscriptionFormats.vue      # Format-specific links
│   └── SubscriptionStats.vue        # Access statistics
```

### Pinia Store

```typescript
// stores/subscription.ts
interface SubscriptionState {
  link: string | null
  shortLink: string | null
  formats: FormatInfo[]
  stats: AccessStats | null
  loading: boolean
  error: string | null
}

const useSubscriptionStore = defineStore('subscription', {
  state: (): SubscriptionState => ({...}),
  actions: {
    async fetchLink(): Promise<void>
    async regenerate(): Promise<void>
    async fetchInfo(): Promise<void>
  }
})
```
