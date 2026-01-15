# Design Document: Commercial System

## Overview

商业化系统为 V Panel 提供完整的商业运营能力，包括套餐管理、订单处理、支付集成、余额系统、续费机制、优惠券、邀请推广和财务报表。系统采用模块化设计，支持灵活配置和扩展。

### Key Design Decisions

1. **支付网关抽象**: 使用策略模式抽象支付网关，便于添加新的支付方式
2. **余额系统**: 采用双写模式确保余额一致性，所有变更记录交易日志
3. **订单状态机**: 使用状态机模式管理订单生命周期
4. **佣金延迟结算**: 佣金在可配置的延迟期后才能提现，防止退款欺诈
5. **幂等性设计**: 支付回调和关键操作实现幂等性，防止重复处理

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Commercial System                                    │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                           API Layer                                      ││
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐     ││
│  │  │ PlanAPI  │ │ OrderAPI │ │PaymentAPI│ │BalanceAPI│ │InviteAPI │     ││
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘     ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                    │                                         │
│  ┌─────────────────────────────────┴───────────────────────────────────────┐│
│  │                         Service Layer                                    ││
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐           ││
│  │  │PlanService │ │OrderService│ │PaymentSvc  │ │BalanceSvc  │           ││
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘           ││
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐           ││
│  │  │CouponSvc   │ │InviteSvc   │ │CommissionSvc│ │InvoiceSvc │           ││
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘           ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                    │                                         │
│  ┌─────────────────────────────────┴───────────────────────────────────────┐│
│  │                       Payment Gateway Layer                              ││
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐           ││
│  │  │  Alipay    │ │ WeChatPay  │ │  PayPal    │ │  Crypto    │           ││
│  │  │  Gateway   │ │  Gateway   │ │  Gateway   │ │  Gateway   │           ││
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘           ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                    │                                         │
│  ┌─────────────────────────────────┴───────────────────────────────────────┐│
│  │                       Repository Layer                                   ││
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐     ││
│  │  │ PlanRepo │ │OrderRepo │ │BalanceRepo│ │CouponRepo│ │InviteRepo│     ││
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘     ││
│  └─────────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Plan Service

```go
// internal/commercial/plan/service.go
type PlanService struct {
    planRepo    repository.PlanRepository
    logger      logger.Logger
}

type Plan struct {
    ID              uint      `json:"id"`
    Name            string    `json:"name"`
    Description     string    `json:"description"`
    TrafficLimit    int64     `json:"traffic_limit"`    // bytes, 0 = unlimited
    Duration        int       `json:"duration"`         // days
    Price           int64     `json:"price"`            // cents
    PlanType        string    `json:"plan_type"`        // monthly, quarterly, yearly, traffic
    ResetCycle      string    `json:"reset_cycle"`      // monthly, on_purchase, never
    IPLimit         int       `json:"ip_limit"`         // 0 = unlimited
    SortOrder       int       `json:"sort_order"`
    IsActive        bool      `json:"is_active"`
    IsRecommended   bool      `json:"is_recommended"`
    GroupID         *uint     `json:"group_id"`
    PaymentMethods  []string  `json:"payment_methods"`  // JSON array
    Features        []string  `json:"features"`         // JSON array
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

func (s *PlanService) Create(plan *Plan) error
func (s *PlanService) Update(id uint, plan *Plan) error
func (s *PlanService) Delete(id uint) error
func (s *PlanService) GetByID(id uint) (*Plan, error)
func (s *PlanService) List(filter PlanFilter) ([]Plan, error)
func (s *PlanService) ListActive() ([]Plan, error)
func (s *PlanService) SetActive(id uint, active bool) error
func (s *PlanService) CalculateMonthlyPrice(plan *Plan) int64
```

### 2. Order Service

```go
// internal/commercial/order/service.go
type OrderService struct {
    orderRepo     repository.OrderRepository
    planService   *plan.PlanService
    balanceService *balance.BalanceService
    couponService *coupon.CouponService
    logger        logger.Logger
}

type Order struct {
    ID              uint      `json:"id"`
    OrderNo         string    `json:"order_no"`         // ORD-20260114-XXXX
    UserID          uint      `json:"user_id"`
    PlanID          uint      `json:"plan_id"`
    CouponID        *uint     `json:"coupon_id"`
    OriginalAmount  int64     `json:"original_amount"`  // cents
    DiscountAmount  int64     `json:"discount_amount"`  // cents
    BalanceUsed     int64     `json:"balance_used"`     // cents
    PayAmount       int64     `json:"pay_amount"`       // cents (actual payment)
    Status          string    `json:"status"`           // pending, paid, completed, cancelled, refunded
    PaymentMethod   string    `json:"payment_method"`
    PaymentNo       string    `json:"payment_no"`       // external payment ID
    PaidAt          *time.Time `json:"paid_at"`
    ExpiredAt       time.Time `json:"expired_at"`
    Notes           string    `json:"notes"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

type OrderStatus string
const (
    OrderStatusPending   OrderStatus = "pending"
    OrderStatusPaid      OrderStatus = "paid"
    OrderStatusCompleted OrderStatus = "completed"
    OrderStatusCancelled OrderStatus = "cancelled"
    OrderStatusRefunded  OrderStatus = "refunded"
)

func (s *OrderService) Create(userID, planID uint, couponCode string) (*Order, error)
func (s *OrderService) GetByID(id uint) (*Order, error)
func (s *OrderService) GetByOrderNo(orderNo string) (*Order, error)
func (s *OrderService) ListByUser(userID uint, page, pageSize int) ([]Order, int64, error)
func (s *OrderService) Cancel(id uint) error
func (s *OrderService) MarkPaid(orderNo string, paymentNo string) error
func (s *OrderService) Complete(id uint) error
func (s *OrderService) ExpirePendingOrders() error  // cron job
func (s *OrderService) GenerateOrderNo() string
func (s *OrderService) ApplyCoupon(order *Order, couponCode string) error
func (s *OrderService) UseBalance(order *Order, amount int64) error
```

### 3. Payment Gateway Interface

```go
// internal/commercial/payment/gateway.go
type PaymentGateway interface {
    Name() string
    CreatePayment(order *Order) (*PaymentRequest, error)
    VerifyCallback(data []byte, signature string) (*PaymentResult, error)
    QueryPayment(paymentNo string) (*PaymentResult, error)
    Refund(paymentNo string, amount int64, reason string) (*RefundResult, error)
}

type PaymentRequest struct {
    PaymentURL  string            `json:"payment_url"`   // redirect URL
    QRCodeURL   string            `json:"qrcode_url"`    // QR code image URL
    QRCodeData  string            `json:"qrcode_data"`   // QR code raw data
    ExpireTime  time.Time         `json:"expire_time"`
    Extra       map[string]string `json:"extra"`
}

type PaymentResult struct {
    Success     bool      `json:"success"`
    OrderNo     string    `json:"order_no"`
    PaymentNo   string    `json:"payment_no"`
    Amount      int64     `json:"amount"`
    PaidAt      time.Time `json:"paid_at"`
    Error       string    `json:"error"`
}

// internal/commercial/payment/alipay.go
type AlipayGateway struct {
    appID       string
    privateKey  string
    publicKey   string
    notifyURL   string
    returnURL   string
}

func (g *AlipayGateway) Name() string { return "alipay" }
func (g *AlipayGateway) CreatePayment(order *Order) (*PaymentRequest, error)
func (g *AlipayGateway) VerifyCallback(data []byte, signature string) (*PaymentResult, error)
func (g *AlipayGateway) QueryPayment(paymentNo string) (*PaymentResult, error)
func (g *AlipayGateway) Refund(paymentNo string, amount int64, reason string) (*RefundResult, error)

// internal/commercial/payment/wechat.go
type WeChatGateway struct {
    appID       string
    mchID       string
    apiKey      string
    certPath    string
    notifyURL   string
}

// internal/commercial/payment/service.go
type PaymentService struct {
    gateways    map[string]PaymentGateway
    orderService *order.OrderService
    logger      logger.Logger
}

func (s *PaymentService) RegisterGateway(gateway PaymentGateway)
func (s *PaymentService) CreatePayment(orderNo string, method string) (*PaymentRequest, error)
func (s *PaymentService) HandleCallback(method string, data []byte, signature string) error
func (s *PaymentService) ProcessRefund(orderID uint, amount int64, reason string) error
```

### 4. Balance Service

```go
// internal/commercial/balance/service.go
type BalanceService struct {
    balanceRepo     repository.BalanceRepository
    transactionRepo repository.TransactionRepository
    logger          logger.Logger
}

type BalanceTransaction struct {
    ID          uint      `json:"id"`
    UserID      uint      `json:"user_id"`
    Type        string    `json:"type"`        // recharge, purchase, refund, commission, adjustment
    Amount      int64     `json:"amount"`      // cents, positive or negative
    Balance     int64     `json:"balance"`     // balance after transaction
    OrderID     *uint     `json:"order_id"`
    Description string    `json:"description"`
    Operator    string    `json:"operator"`    // system, admin username
    CreatedAt   time.Time `json:"created_at"`
}

func (s *BalanceService) GetBalance(userID uint) (int64, error)
func (s *BalanceService) Recharge(userID uint, amount int64, orderID uint) error
func (s *BalanceService) Deduct(userID uint, amount int64, orderID uint, desc string) error
func (s *BalanceService) Refund(userID uint, amount int64, orderID uint) error
func (s *BalanceService) AddCommission(userID uint, amount int64, desc string) error
func (s *BalanceService) Adjust(userID uint, amount int64, reason string, operator string) error
func (s *BalanceService) GetTransactions(userID uint, page, pageSize int) ([]BalanceTransaction, int64, error)
func (s *BalanceService) CanDeduct(userID uint, amount int64) bool
```

### 5. Coupon Service

```go
// internal/commercial/coupon/service.go
type CouponService struct {
    couponRepo repository.CouponRepository
    logger     logger.Logger
}

type Coupon struct {
    ID              uint       `json:"id"`
    Code            string     `json:"code"`
    Name            string     `json:"name"`
    Type            string     `json:"type"`            // fixed, percentage
    Value           int64      `json:"value"`           // cents or percentage * 100
    MinOrderAmount  int64      `json:"min_order_amount"` // cents
    MaxDiscount     int64      `json:"max_discount"`    // cents, for percentage type
    TotalLimit      int        `json:"total_limit"`     // 0 = unlimited
    PerUserLimit    int        `json:"per_user_limit"`  // 0 = unlimited
    UsedCount       int        `json:"used_count"`
    PlanIDs         []uint     `json:"plan_ids"`        // empty = all plans
    StartAt         time.Time  `json:"start_at"`
    ExpireAt        time.Time  `json:"expire_at"`
    IsActive        bool       `json:"is_active"`
    CreatedAt       time.Time  `json:"created_at"`
}

type CouponUsage struct {
    ID        uint      `json:"id"`
    CouponID  uint      `json:"coupon_id"`
    UserID    uint      `json:"user_id"`
    OrderID   uint      `json:"order_id"`
    Discount  int64     `json:"discount"`
    UsedAt    time.Time `json:"used_at"`
}

func (s *CouponService) Create(coupon *Coupon) error
func (s *CouponService) GetByCode(code string) (*Coupon, error)
func (s *CouponService) Validate(code string, userID uint, planID uint, amount int64) (*Coupon, int64, error)
func (s *CouponService) Use(couponID, userID, orderID uint, discount int64) error
func (s *CouponService) CalculateDiscount(coupon *Coupon, amount int64) int64
func (s *CouponService) GenerateBatchCodes(prefix string, count int) ([]string, error)
func (s *CouponService) List(filter CouponFilter) ([]Coupon, int64, error)
```

### 6. Invite Service

```go
// internal/commercial/invite/service.go
type InviteService struct {
    inviteRepo      repository.InviteRepository
    commissionService *commission.CommissionService
    logger          logger.Logger
}

type InviteCode struct {
    ID          uint      `json:"id"`
    UserID      uint      `json:"user_id"`
    Code        string    `json:"code"`
    InviteCount int       `json:"invite_count"`
    CreatedAt   time.Time `json:"created_at"`
}

type Referral struct {
    ID          uint      `json:"id"`
    InviterID   uint      `json:"inviter_id"`
    InviteeID   uint      `json:"invitee_id"`
    InviteCode  string    `json:"invite_code"`
    Status      string    `json:"status"`      // registered, converted
    ConvertedAt *time.Time `json:"converted_at"`
    CreatedAt   time.Time `json:"created_at"`
}

func (s *InviteService) GetOrCreateCode(userID uint) (*InviteCode, error)
func (s *InviteService) GetByCode(code string) (*InviteCode, error)
func (s *InviteService) RecordReferral(inviteCode string, inviteeID uint) error
func (s *InviteService) MarkConverted(inviteeID uint) error
func (s *InviteService) GetReferrals(userID uint, page, pageSize int) ([]Referral, int64, error)
func (s *InviteService) GetStats(userID uint) (*InviteStats, error)
func (s *InviteService) GenerateInviteLink(code string) string
```

### 7. Commission Service

```go
// internal/commercial/commission/service.go
type CommissionService struct {
    commissionRepo  repository.CommissionRepository
    balanceService  *balance.BalanceService
    config          *CommissionConfig
    logger          logger.Logger
}

type CommissionConfig struct {
    Enabled         bool    `json:"enabled"`
    Rate            float64 `json:"rate"`              // e.g., 0.1 = 10%
    FixedBonus      int64   `json:"fixed_bonus"`       // cents, one-time bonus
    TrafficBonus    int64   `json:"traffic_bonus"`     // bytes
    SettlementDelay int     `json:"settlement_delay"`  // days
    MinWithdraw     int64   `json:"min_withdraw"`      // cents
    MultiLevel      bool    `json:"multi_level"`
    MaxLevel        int     `json:"max_level"`         // max referral depth
}

type Commission struct {
    ID          uint       `json:"id"`
    UserID      uint       `json:"user_id"`       // inviter
    FromUserID  uint       `json:"from_user_id"`  // invitee
    OrderID     uint       `json:"order_id"`
    Amount      int64      `json:"amount"`        // cents
    Rate        float64    `json:"rate"`
    Level       int        `json:"level"`         // referral level
    Status      string     `json:"status"`        // pending, confirmed, cancelled
    ConfirmAt   *time.Time `json:"confirm_at"`
    CreatedAt   time.Time  `json:"created_at"`
}

func (s *CommissionService) Calculate(order *Order) ([]Commission, error)
func (s *CommissionService) Create(commissions []Commission) error
func (s *CommissionService) Confirm(id uint) error
func (s *CommissionService) Cancel(id uint) error
func (s *CommissionService) ConfirmPendingCommissions() error  // cron job
func (s *CommissionService) GetPending(userID uint) ([]Commission, int64, error)
func (s *CommissionService) GetConfirmed(userID uint) ([]Commission, int64, error)
func (s *CommissionService) GetTotalEarnings(userID uint) (int64, error)
```

### 8. Invoice Service

```go
// internal/commercial/invoice/service.go
type InvoiceService struct {
    invoiceRepo repository.InvoiceRepository
    config      *InvoiceConfig
    logger      logger.Logger
}

type InvoiceConfig struct {
    CompanyName    string `json:"company_name"`
    CompanyAddress string `json:"company_address"`
    TaxID          string `json:"tax_id"`
    NumberFormat   string `json:"number_format"`  // e.g., "INV-{YYYY}{MM}-{SEQ}"
}

type Invoice struct {
    ID          uint      `json:"id"`
    InvoiceNo   string    `json:"invoice_no"`
    OrderID     uint      `json:"order_id"`
    UserID      uint      `json:"user_id"`
    Amount      int64     `json:"amount"`
    Content     string    `json:"content"`      // JSON with line items
    PDFPath     string    `json:"pdf_path"`
    CreatedAt   time.Time `json:"created_at"`
}

func (s *InvoiceService) Generate(orderID uint) (*Invoice, error)
func (s *InvoiceService) GetByID(id uint) (*Invoice, error)
func (s *InvoiceService) GetByOrder(orderID uint) (*Invoice, error)
func (s *InvoiceService) ListByUser(userID uint, page, pageSize int) ([]Invoice, int64, error)
func (s *InvoiceService) GeneratePDF(invoice *Invoice) ([]byte, error)
func (s *InvoiceService) GenerateInvoiceNo() string
```

## Data Models

### Database Schema

```sql
-- Plans table
CREATE TABLE plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    traffic_limit BIGINT DEFAULT 0,
    duration INTEGER NOT NULL,
    price BIGINT NOT NULL,
    plan_type VARCHAR(32) NOT NULL DEFAULT 'monthly',
    reset_cycle VARCHAR(32) DEFAULT 'monthly',
    ip_limit INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    is_recommended BOOLEAN DEFAULT FALSE,
    group_id INTEGER,
    payment_methods TEXT,
    features TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_plans_active ON plans(is_active, sort_order);

-- Orders table
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    plan_id INTEGER NOT NULL,
    coupon_id INTEGER,
    original_amount BIGINT NOT NULL,
    discount_amount BIGINT DEFAULT 0,
    balance_used BIGINT DEFAULT 0,
    pay_amount BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    payment_method VARCHAR(32),
    payment_no VARCHAR(128),
    paid_at DATETIME,
    expired_at DATETIME NOT NULL,
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (plan_id) REFERENCES plans(id),
    FOREIGN KEY (coupon_id) REFERENCES coupons(id)
);

CREATE INDEX idx_orders_user ON orders(user_id, created_at DESC);
CREATE INDEX idx_orders_status ON orders(status, expired_at);
CREATE INDEX idx_orders_order_no ON orders(order_no);

-- Balance transactions table
CREATE TABLE balance_transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type VARCHAR(32) NOT NULL,
    amount BIGINT NOT NULL,
    balance BIGINT NOT NULL,
    order_id INTEGER,
    description VARCHAR(256),
    operator VARCHAR(64),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE INDEX idx_balance_tx_user ON balance_transactions(user_id, created_at DESC);

-- Coupons table
CREATE TABLE coupons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code VARCHAR(32) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    type VARCHAR(16) NOT NULL,
    value BIGINT NOT NULL,
    min_order_amount BIGINT DEFAULT 0,
    max_discount BIGINT DEFAULT 0,
    total_limit INTEGER DEFAULT 0,
    per_user_limit INTEGER DEFAULT 1,
    used_count INTEGER DEFAULT 0,
    plan_ids TEXT,
    start_at DATETIME NOT NULL,
    expire_at DATETIME NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_coupons_code ON coupons(code);
CREATE INDEX idx_coupons_active ON coupons(is_active, start_at, expire_at);

-- Coupon usage table
CREATE TABLE coupon_usages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    coupon_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    order_id INTEGER NOT NULL,
    discount BIGINT NOT NULL,
    used_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (coupon_id) REFERENCES coupons(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE INDEX idx_coupon_usage ON coupon_usages(coupon_id, user_id);

-- Invite codes table
CREATE TABLE invite_codes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    code VARCHAR(16) NOT NULL UNIQUE,
    invite_count INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_invite_codes_code ON invite_codes(code);

-- Referrals table
CREATE TABLE referrals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    inviter_id INTEGER NOT NULL,
    invitee_id INTEGER NOT NULL UNIQUE,
    invite_code VARCHAR(16) NOT NULL,
    status VARCHAR(32) DEFAULT 'registered',
    converted_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (inviter_id) REFERENCES users(id),
    FOREIGN KEY (invitee_id) REFERENCES users(id)
);

CREATE INDEX idx_referrals_inviter ON referrals(inviter_id);

-- Commissions table
CREATE TABLE commissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    from_user_id INTEGER NOT NULL,
    order_id INTEGER NOT NULL,
    amount BIGINT NOT NULL,
    rate REAL NOT NULL,
    level INTEGER DEFAULT 1,
    status VARCHAR(32) DEFAULT 'pending',
    confirm_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (from_user_id) REFERENCES users(id),
    FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE INDEX idx_commissions_user ON commissions(user_id, status);

-- Invoices table
CREATE TABLE invoices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_no VARCHAR(64) NOT NULL UNIQUE,
    order_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    amount BIGINT NOT NULL,
    content TEXT NOT NULL,
    pdf_path VARCHAR(256),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_invoices_user ON invoices(user_id, created_at DESC);

-- User balance extension (add to users table)
ALTER TABLE users ADD COLUMN balance BIGINT DEFAULT 0;
ALTER TABLE users ADD COLUMN auto_renewal BOOLEAN DEFAULT FALSE;
```

## API Endpoints

### Plan APIs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/plans` | List active plans (public) |
| GET | `/api/plans/:id` | Get plan details |
| POST | `/api/admin/plans` | Create plan (admin) |
| PUT | `/api/admin/plans/:id` | Update plan (admin) |
| DELETE | `/api/admin/plans/:id` | Delete plan (admin) |
| PUT | `/api/admin/plans/:id/status` | Toggle plan status (admin) |

### Order APIs

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/orders` | Create order |
| GET | `/api/orders` | List user's orders |
| GET | `/api/orders/:id` | Get order details |
| POST | `/api/orders/:id/cancel` | Cancel pending order |
| GET | `/api/admin/orders` | List all orders (admin) |
| PUT | `/api/admin/orders/:id` | Update order (admin) |

### Payment APIs

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/payments/create` | Create payment |
| POST | `/api/payments/callback/:method` | Payment callback |
| GET | `/api/payments/status/:orderNo` | Check payment status |
| POST | `/api/admin/refunds` | Process refund (admin) |

### Balance APIs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/balance` | Get user balance |
| GET | `/api/balance/transactions` | Get transaction history |
| POST | `/api/balance/recharge` | Create recharge order |
| POST | `/api/admin/balance/adjust` | Adjust balance (admin) |

### Coupon APIs

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/coupons/validate` | Validate coupon code |
| GET | `/api/admin/coupons` | List coupons (admin) |
| POST | `/api/admin/coupons` | Create coupon (admin) |
| PUT | `/api/admin/coupons/:id` | Update coupon (admin) |
| DELETE | `/api/admin/coupons/:id` | Delete coupon (admin) |
| POST | `/api/admin/coupons/batch` | Generate batch codes (admin) |

### Invite APIs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/invite/code` | Get user's invite code |
| GET | `/api/invite/referrals` | List referrals |
| GET | `/api/invite/stats` | Get invite statistics |
| GET | `/api/invite/commissions` | Get commission history |
| GET | `/api/admin/invite/stats` | Global invite stats (admin) |

### Invoice APIs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/invoices` | List user's invoices |
| GET | `/api/invoices/:id/download` | Download invoice PDF |
| GET | `/api/admin/invoices` | List all invoices (admin) |

### Report APIs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/reports/revenue` | Revenue report (admin) |
| GET | `/api/admin/reports/orders` | Order statistics (admin) |
| GET | `/api/admin/reports/commissions` | Commission report (admin) |
| GET | `/api/admin/reports/export` | Export report (admin) |

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Order ID Uniqueness
*For any* two orders in the system, their order numbers SHALL be unique.
**Validates: Requirements 3.3**

### Property 2: Coupon Discount Calculation
*For any* valid coupon and order amount, the calculated discount SHALL not exceed the order amount, and for percentage coupons, SHALL not exceed the max_discount limit.
**Validates: Requirements 3.5, 8.2**

### Property 3: Balance Non-Negative Invariant
*For any* balance operation, the resulting user balance SHALL never be negative.
**Validates: Requirements 6.9**

### Property 4: Balance Transaction Consistency
*For any* sequence of balance transactions for a user, the final balance SHALL equal the sum of all transaction amounts.
**Validates: Requirements 6.4, 6.5**

### Property 5: Order Status Transitions
*For any* order, status transitions SHALL follow the valid state machine: pending → paid → completed, pending → cancelled, paid → refunded.
**Validates: Requirements 5.4**

### Property 6: Coupon Usage Limit
*For any* coupon with usage limits, the used_count SHALL never exceed total_limit, and per-user usage SHALL never exceed per_user_limit.
**Validates: Requirements 8.5, 8.6**

### Property 7: Order Expiration
*For any* pending order past its expired_at time, the system SHALL mark it as cancelled.
**Validates: Requirements 3.7, 3.8**

### Property 8: Invite Code Uniqueness
*For any* two invite codes in the system, their codes SHALL be unique.
**Validates: Requirements 9.1**

### Property 9: Commission Calculation
*For any* order with a referrer, the commission amount SHALL equal order amount multiplied by commission rate.
**Validates: Requirements 10.1**

### Property 10: Payment Callback Idempotency
*For any* payment callback processed multiple times with the same payment_no, the order status and balance SHALL only be updated once.
**Validates: Requirements 14.8**

### Property 11: Plan Price Per Month Calculation
*For any* plan, the monthly price SHALL equal (price / duration) * 30, rounded appropriately.
**Validates: Requirements 2.4**

### Property 12: Refund Balance Restoration
*For any* refund to balance, the user's balance SHALL increase by exactly the refund amount.
**Validates: Requirements 13.4, 13.5**

### Property 13: Subscription Activation on Payment
*For any* successful payment, the user's subscription SHALL be activated with correct expiration date based on plan duration.
**Validates: Requirements 4.7**

### Property 14: Coupon Validation Rules
*For any* coupon validation, the system SHALL reject coupons that are: expired, inactive, below minimum order amount, or exceeding usage limits.
**Validates: Requirements 8.4, 8.7**

## Error Handling

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `PLAN_NOT_FOUND` | 404 | Plan does not exist |
| `PLAN_INACTIVE` | 400 | Plan is not active |
| `ORDER_NOT_FOUND` | 404 | Order does not exist |
| `ORDER_EXPIRED` | 400 | Order has expired |
| `ORDER_ALREADY_PAID` | 400 | Order is already paid |
| `ORDER_CANNOT_CANCEL` | 400 | Order cannot be cancelled |
| `INSUFFICIENT_BALANCE` | 400 | Insufficient account balance |
| `COUPON_NOT_FOUND` | 404 | Coupon does not exist |
| `COUPON_EXPIRED` | 400 | Coupon has expired |
| `COUPON_INACTIVE` | 400 | Coupon is not active |
| `COUPON_LIMIT_REACHED` | 400 | Coupon usage limit reached |
| `COUPON_MIN_AMOUNT` | 400 | Order amount below minimum |
| `COUPON_PLAN_MISMATCH` | 400 | Coupon not valid for this plan |
| `PAYMENT_FAILED` | 400 | Payment processing failed |
| `PAYMENT_TIMEOUT` | 400 | Payment timed out |
| `INVALID_SIGNATURE` | 401 | Invalid payment callback signature |
| `INVITE_CODE_INVALID` | 400 | Invalid invite code |
| `SELF_REFERRAL` | 400 | Cannot use own invite code |
| `REFUND_FAILED` | 400 | Refund processing failed |

### Error Response Format

```json
{
  "code": "INSUFFICIENT_BALANCE",
  "message": "账户余额不足",
  "details": {
    "required": 10000,
    "available": 5000
  }
}
```

## Testing Strategy

### Unit Tests

Unit tests will cover:
- Order number generation and uniqueness
- Coupon discount calculation logic
- Balance transaction operations
- Commission calculation
- Invoice number generation
- Price per month calculation
- Status transition validation

### Property-Based Tests

Property-based tests will use Go's `testing/quick` package or `gopter` library to verify:
- Order ID uniqueness across generated orders
- Balance non-negative invariant
- Coupon discount bounds
- Status transition validity
- Payment callback idempotency

**Configuration**: Each property test will run minimum 100 iterations.

**Test Annotation Format**: Each test will be tagged with:
```go
// Feature: commercial-system, Property N: Property description
// Validates: Requirements X.Y
```

### Integration Tests

Integration tests will cover:
- Complete order flow (create → pay → complete)
- Coupon application flow
- Balance recharge and deduction
- Referral and commission flow
- Refund processing

### Test Files Structure

```
internal/commercial/
├── plan/
│   ├── service.go
│   ├── service_test.go
│   └── service_property_test.go
├── order/
│   ├── service.go
│   ├── service_test.go
│   └── service_property_test.go
├── balance/
│   ├── service.go
│   ├── service_test.go
│   └── service_property_test.go
├── coupon/
│   ├── service.go
│   ├── service_test.go
│   └── service_property_test.go
├── invite/
│   ├── service.go
│   └── service_test.go
├── commission/
│   ├── service.go
│   ├── service_test.go
│   └── service_property_test.go
├── payment/
│   ├── gateway.go
│   ├── alipay.go
│   ├── wechat.go
│   └── service_test.go
└── invoice/
    ├── service.go
    └── service_test.go
```

## Frontend Components

### Vue Components Structure

```
web/src/
├── views/
│   ├── Plans.vue              # Plan listing page
│   ├── PlanDetail.vue         # Plan detail and purchase
│   ├── Orders.vue             # Order history
│   ├── OrderDetail.vue        # Order detail
│   ├── Payment.vue            # Payment page
│   ├── Balance.vue            # Balance and recharge
│   ├── Invite.vue             # Invite and referrals
│   └── Invoices.vue           # Invoice history
├── views/admin/
│   ├── AdminPlans.vue         # Plan management
│   ├── AdminOrders.vue        # Order management
│   ├── AdminCoupons.vue       # Coupon management
│   ├── AdminInvites.vue       # Invite statistics
│   ├── AdminReports.vue       # Financial reports
│   └── AdminRefunds.vue       # Refund management
├── components/commercial/
│   ├── PlanCard.vue           # Plan display card
│   ├── OrderCard.vue          # Order summary card
│   ├── PaymentMethods.vue     # Payment method selector
│   ├── CouponInput.vue        # Coupon code input
│   ├── BalanceCard.vue        # Balance display
│   ├── InviteCard.vue         # Invite code display
│   ├── CommissionList.vue     # Commission history
│   └── RevenueChart.vue       # Revenue chart
├── stores/
│   ├── plan.ts                # Plan store
│   ├── order.ts               # Order store
│   ├── balance.ts             # Balance store
│   ├── coupon.ts              # Coupon store
│   └── invite.ts              # Invite store
└── api/modules/
    ├── plans.ts               # Plan API
    ├── orders.ts              # Order API
    ├── payments.ts            # Payment API
    ├── balance.ts             # Balance API
    ├── coupons.ts             # Coupon API
    ├── invites.ts             # Invite API
    └── invoices.ts            # Invoice API
```


## Additional Components (Requirements 15-20)

### 9. Trial Service

```go
// internal/commercial/trial/service.go
type TrialService struct {
    trialRepo   repository.TrialRepository
    userService *user.UserService
    config      *TrialConfig
    logger      logger.Logger
}

type TrialConfig struct {
    Enabled           bool  `json:"enabled"`
    Duration          int   `json:"duration"`           // days
    TrafficLimit      int64 `json:"traffic_limit"`      // bytes
    RequireEmailVerify bool `json:"require_email_verify"`
    AutoActivate      bool  `json:"auto_activate"`      // on registration
    FeatureRestrictions []string `json:"feature_restrictions"`
}

type Trial struct {
    ID          uint       `json:"id"`
    UserID      uint       `json:"user_id" gorm:"uniqueIndex"`
    Status      string     `json:"status"`      // active, expired, converted
    StartAt     time.Time  `json:"start_at"`
    ExpireAt    time.Time  `json:"expire_at"`
    TrafficUsed int64      `json:"traffic_used"`
    ConvertedAt *time.Time `json:"converted_at"`
    CreatedAt   time.Time  `json:"created_at"`
}

func (s *TrialService) ActivateTrial(userID uint) (*Trial, error)
func (s *TrialService) GetTrial(userID uint) (*Trial, error)
func (s *TrialService) HasUsedTrial(userID uint) bool
func (s *TrialService) ExpireTrials() error  // cron job
func (s *TrialService) MarkConverted(userID uint) error
func (s *TrialService) GetConversionRate() float64
```

### 10. Plan Change Service

```go
// internal/commercial/planchange/service.go
type PlanChangeService struct {
    planService    *plan.PlanService
    orderService   *order.OrderService
    balanceService *balance.BalanceService
    logger         logger.Logger
}

type PlanChangeRequest struct {
    UserID      uint `json:"user_id"`
    CurrentPlan uint `json:"current_plan"`
    NewPlan     uint `json:"new_plan"`
    Immediate   bool `json:"immediate"`  // immediate or next cycle
}

type PlanChangeResult struct {
    PriceDifference int64     `json:"price_difference"`  // positive = pay more, negative = refund
    RemainingDays   int       `json:"remaining_days"`
    NewExpireAt     time.Time `json:"new_expire_at"`
    IsUpgrade       bool      `json:"is_upgrade"`
}

func (s *PlanChangeService) CalculateChange(req *PlanChangeRequest) (*PlanChangeResult, error)
func (s *PlanChangeService) ExecuteUpgrade(req *PlanChangeRequest) (*Order, error)
func (s *PlanChangeService) ScheduleDowngrade(req *PlanChangeRequest) error
func (s *PlanChangeService) GetPendingDowngrade(userID uint) (*PendingDowngrade, error)
func (s *PlanChangeService) CancelPendingDowngrade(userID uint) error
```

### 11. Payment Retry Service

```go
// internal/commercial/payment/retry.go
type PaymentRetryService struct {
    orderService   *order.OrderService
    paymentService *PaymentService
    config         *RetryConfig
    logger         logger.Logger
}

type RetryConfig struct {
    Enabled       bool  `json:"enabled"`
    MaxRetries    int   `json:"max_retries"`     // default 3
    RetryIntervals []int `json:"retry_intervals"` // minutes: [60, 240, 1440]
}

type PaymentRetry struct {
    ID          uint      `json:"id"`
    OrderID     uint      `json:"order_id"`
    AttemptNo   int       `json:"attempt_no"`
    Status      string    `json:"status"`      // pending, success, failed
    Error       string    `json:"error"`
    ScheduledAt time.Time `json:"scheduled_at"`
    ExecutedAt  *time.Time `json:"executed_at"`
    CreatedAt   time.Time `json:"created_at"`
}

func (s *PaymentRetryService) ScheduleRetry(orderID uint) error
func (s *PaymentRetryService) ExecutePendingRetries() error  // cron job
func (s *PaymentRetryService) CancelRetries(orderID uint) error
func (s *PaymentRetryService) GetRetryHistory(orderID uint) ([]PaymentRetry, error)
```

### 12. Currency Service

```go
// internal/commercial/currency/service.go
type CurrencyService struct {
    exchangeRepo repository.ExchangeRateRepository
    config       *CurrencyConfig
    logger       logger.Logger
}

type CurrencyConfig struct {
    BaseCurrency      string   `json:"base_currency"`      // e.g., "CNY"
    SupportedCurrencies []string `json:"supported_currencies"`
    ExchangeRateAPI   string   `json:"exchange_rate_api"`
    CacheTTL          int      `json:"cache_ttl"`          // minutes
}

type ExchangeRate struct {
    FromCurrency string    `json:"from_currency"`
    ToCurrency   string    `json:"to_currency"`
    Rate         float64   `json:"rate"`
    UpdatedAt    time.Time `json:"updated_at"`
}

func (s *CurrencyService) GetRate(from, to string) (float64, error)
func (s *CurrencyService) Convert(amount int64, from, to string) (int64, error)
func (s *CurrencyService) UpdateRates() error  // cron job
func (s *CurrencyService) FormatPrice(amount int64, currency string) string
func (s *CurrencyService) DetectCurrency(ip string) string
```

### 13. Subscription Pause Service

```go
// internal/commercial/pause/service.go
type PauseService struct {
    pauseRepo   repository.PauseRepository
    userService *user.UserService
    config      *PauseConfig
    logger      logger.Logger
}

type PauseConfig struct {
    Enabled         bool `json:"enabled"`
    MaxDuration     int  `json:"max_duration"`      // days
    MaxPerCycle     int  `json:"max_per_cycle"`     // times per billing cycle
    AllowedPlanIDs  []uint `json:"allowed_plan_ids"` // empty = all plans
}

type SubscriptionPause struct {
    ID          uint       `json:"id"`
    UserID      uint       `json:"user_id"`
    PausedAt    time.Time  `json:"paused_at"`
    ResumedAt   *time.Time `json:"resumed_at"`
    RemainingDays int      `json:"remaining_days"`
    RemainingTraffic int64 `json:"remaining_traffic"`
    AutoResumeAt time.Time `json:"auto_resume_at"`
    CreatedAt   time.Time  `json:"created_at"`
}

func (s *PauseService) Pause(userID uint) (*SubscriptionPause, error)
func (s *PauseService) Resume(userID uint) error
func (s *PauseService) GetActivePause(userID uint) (*SubscriptionPause, error)
func (s *PauseService) CanPause(userID uint) (bool, string)
func (s *PauseService) AutoResumePaused() error  // cron job
func (s *PauseService) GetPauseHistory(userID uint) ([]SubscriptionPause, error)
```

### 14. Gift Card Service

```go
// internal/commercial/giftcard/service.go
type GiftCardService struct {
    giftCardRepo repository.GiftCardRepository
    balanceService *balance.BalanceService
    logger       logger.Logger
}

type GiftCard struct {
    ID          uint       `json:"id"`
    Code        string     `json:"code" gorm:"uniqueIndex"`
    Value       int64      `json:"value"`       // cents
    Status      string     `json:"status"`      // active, redeemed, expired
    PurchaserID *uint      `json:"purchaser_id"`
    RedeemerID  *uint      `json:"redeemer_id"`
    OrderID     *uint      `json:"order_id"`    // purchase order
    ExpireAt    time.Time  `json:"expire_at"`
    RedeemedAt  *time.Time `json:"redeemed_at"`
    CreatedAt   time.Time  `json:"created_at"`
}

func (s *GiftCardService) CreateBatch(count int, value int64, expireDays int) ([]GiftCard, error)
func (s *GiftCardService) Purchase(userID uint, value int64) (*GiftCard, *Order, error)
func (s *GiftCardService) Redeem(code string, userID uint) error
func (s *GiftCardService) GetByCode(code string) (*GiftCard, error)
func (s *GiftCardService) ListByUser(userID uint) ([]GiftCard, error)
func (s *GiftCardService) ExpireGiftCards() error  // cron job
func (s *GiftCardService) GenerateCode() string
```

## Additional Correctness Properties

### Property 15: Trial Uniqueness
*For any* user, they SHALL have at most one trial record, and once used, cannot activate another trial.
**Validates: Requirements 15.3**

### Property 16: Plan Change Proration
*For any* plan upgrade, the prorated price SHALL equal (new_price - old_price) * (remaining_days / total_days).
**Validates: Requirements 16.3**

### Property 17: Payment Retry Limit
*For any* order, the number of payment retry attempts SHALL not exceed the configured max_retries.
**Validates: Requirements 17.2**

### Property 18: Currency Conversion Consistency
*For any* amount converted from currency A to B and back to A, the result SHALL be within acceptable rounding tolerance of the original amount.
**Validates: Requirements 18.6**

### Property 19: Pause Duration Limit
*For any* subscription pause, the duration SHALL not exceed the configured max_duration.
**Validates: Requirements 19.3**

### Property 20: Gift Card Redemption
*For any* gift card redemption, the user's balance SHALL increase by exactly the gift card value, and the gift card status SHALL change to redeemed.
**Validates: Requirements 20.6**

## Additional Database Schema

```sql
-- Trials table
CREATE TABLE trials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    status VARCHAR(32) DEFAULT 'active',
    start_at DATETIME NOT NULL,
    expire_at DATETIME NOT NULL,
    traffic_used BIGINT DEFAULT 0,
    converted_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Pending downgrades table
CREATE TABLE pending_downgrades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    current_plan_id INTEGER NOT NULL,
    new_plan_id INTEGER NOT NULL,
    effective_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Payment retries table
CREATE TABLE payment_retries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    attempt_no INTEGER NOT NULL,
    status VARCHAR(32) DEFAULT 'pending',
    error TEXT,
    scheduled_at DATETIME NOT NULL,
    executed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE INDEX idx_payment_retries_scheduled ON payment_retries(status, scheduled_at);

-- Exchange rates table
CREATE TABLE exchange_rates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    rate REAL NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(from_currency, to_currency)
);

-- Subscription pauses table
CREATE TABLE subscription_pauses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    paused_at DATETIME NOT NULL,
    resumed_at DATETIME,
    remaining_days INTEGER NOT NULL,
    remaining_traffic BIGINT NOT NULL,
    auto_resume_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_pauses_user ON subscription_pauses(user_id, paused_at DESC);
CREATE INDEX idx_pauses_auto_resume ON subscription_pauses(resumed_at, auto_resume_at);

-- Gift cards table
CREATE TABLE gift_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code VARCHAR(32) NOT NULL UNIQUE,
    value BIGINT NOT NULL,
    status VARCHAR(32) DEFAULT 'active',
    purchaser_id INTEGER,
    redeemer_id INTEGER,
    order_id INTEGER,
    expire_at DATETIME NOT NULL,
    redeemed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (purchaser_id) REFERENCES users(id),
    FOREIGN KEY (redeemer_id) REFERENCES users(id),
    FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE INDEX idx_gift_cards_code ON gift_cards(code);
CREATE INDEX idx_gift_cards_status ON gift_cards(status, expire_at);
```

## Additional API Endpoints

### Trial APIs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/trial` | Get user's trial status |
| POST | `/api/trial/activate` | Activate trial |
| GET | `/api/admin/trials` | List all trials (admin) |
| POST | `/api/admin/trials/grant` | Grant trial to user (admin) |

### Plan Change APIs

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/plan-change/calculate` | Calculate plan change |
| POST | `/api/plan-change/upgrade` | Execute upgrade |
| POST | `/api/plan-change/downgrade` | Schedule downgrade |
| DELETE | `/api/plan-change/downgrade` | Cancel pending downgrade |

### Subscription Pause APIs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/subscription/pause` | Get pause status |
| POST | `/api/subscription/pause` | Pause subscription |
| POST | `/api/subscription/resume` | Resume subscription |

### Gift Card APIs

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/gift-cards/redeem` | Redeem gift card |
| GET | `/api/gift-cards` | List user's gift cards |
| POST | `/api/gift-cards/purchase` | Purchase gift card |
| GET | `/api/admin/gift-cards` | List all gift cards (admin) |
| POST | `/api/admin/gift-cards/batch` | Create batch (admin) |
