# Requirements Document

## Introduction

本功能为 V Panel 提供完整的商业化系统，包括套餐管理、订单处理、充值续费和邀请推广功能。该系统使 V Panel 能够作为商业化代理服务运营，支持用户购买套餐、管理订单、充值账户余额以及通过邀请获得奖励。

## Glossary

- **Plan**: 套餐，定义流量、时长、价格等服务内容的产品
- **Order**: 订单，用户购买套餐或充值的交易记录
- **Payment**: 支付，用户完成订单的付款行为
- **Balance**: 余额，用户账户中可用于购买的金额
- **Recharge**: 充值，用户向账户添加余额的行为
- **Renewal**: 续费，用户延长现有服务期限的行为
- **Invite_Code**: 邀请码，用于邀请新用户注册的唯一代码
- **Commission**: 佣金，邀请人从被邀请人消费中获得的奖励
- **Coupon**: 优惠券，可用于订单折扣的凭证
- **System**: V Panel 应用程序系统
- **User**: 使用代理服务的终端用户
- **Admin**: 系统管理员

## Requirements

### Requirement 1: 套餐管理

**User Story:** As an admin, I want to create and manage service plans, so that users can purchase different levels of service.

#### Acceptance Criteria

1. THE Admin_Panel SHALL provide interface to create, edit, and delete plans
2. THE Plan SHALL include: name, description, traffic limit, duration (days), price, and status
3. THE System SHALL support multiple plan types: monthly, quarterly, yearly, one-time traffic
4. THE Plan SHALL support setting traffic reset cycle (monthly, on purchase date, never)
5. THE System SHALL support plan sorting and display order configuration
6. THE Admin_Panel SHALL allow enabling/disabling plans without deletion
7. THE Plan SHALL support setting maximum concurrent IP limit
8. THE System SHALL support plan groups/categories for organization
9. WHEN a plan is deleted THEN the System SHALL preserve existing user subscriptions
10. THE Plan SHALL support setting available payment methods

### Requirement 2: 套餐展示

**User Story:** As a user, I want to view available plans, so that I can choose the right service for my needs.

#### Acceptance Criteria

1. THE User_Portal SHALL display all active plans in an organized layout
2. THE Plan_Display SHALL show: name, price, traffic, duration, and features
3. THE System SHALL highlight recommended or popular plans
4. THE Plan_Display SHALL show price per month for comparison
5. THE User_Portal SHALL support filtering plans by type or price range
6. WHEN user is logged in THEN the System SHALL show personalized recommendations
7. THE Plan_Display SHALL clearly indicate any limitations or restrictions
8. THE System SHALL display plan comparison table for easy selection

### Requirement 3: 订单创建

**User Story:** As a user, I want to create orders to purchase plans, so that I can access the proxy service.

#### Acceptance Criteria

1. WHEN a user selects a plan THEN the System SHALL create a pending order
2. THE Order SHALL include: order ID, user ID, plan ID, amount, status, and timestamps
3. THE System SHALL generate unique order IDs with prefix (e.g., ORD-20260114-XXXX)
4. THE Order SHALL support applying coupon codes for discounts
5. WHEN coupon is applied THEN the System SHALL validate and calculate discounted price
6. THE System SHALL support using account balance for partial or full payment
7. THE Order SHALL have expiration time (default 30 minutes) for pending orders
8. WHEN order expires THEN the System SHALL automatically cancel it
9. THE System SHALL prevent duplicate orders for the same plan within short period
10. THE Order_Page SHALL display order summary before payment confirmation

### Requirement 4: 支付集成

**User Story:** As a user, I want to pay for orders using various payment methods, so that I can complete purchases conveniently.

#### Acceptance Criteria

1. THE System SHALL support Alipay payment integration
2. THE System SHALL support WeChat Pay integration
3. THE System SHALL support PayPal payment integration (optional)
4. THE System SHALL support cryptocurrency payment (USDT, BTC) via third-party gateway (optional)
5. WHEN payment is initiated THEN the System SHALL redirect to payment gateway or display QR code
6. THE System SHALL handle payment callbacks/webhooks to update order status
7. WHEN payment succeeds THEN the System SHALL immediately activate user's subscription
8. IF payment fails THEN the System SHALL allow retry without creating new order
9. THE System SHALL support payment timeout handling (15 minutes default)
10. THE Admin_Panel SHALL allow configuring payment gateway credentials

### Requirement 5: 订单管理

**User Story:** As a user, I want to view and manage my orders, so that I can track my purchases.

#### Acceptance Criteria

1. THE User_Portal SHALL display order history with pagination
2. THE Order_List SHALL show: order ID, plan name, amount, status, and date
3. THE Order_Detail SHALL show complete order information including payment details
4. THE System SHALL support order statuses: pending, paid, completed, cancelled, refunded
5. THE User_Portal SHALL allow cancelling pending orders
6. THE Admin_Panel SHALL display all orders with filtering and search
7. THE Admin_Panel SHALL allow manual order status updates
8. THE System SHALL send order confirmation email after successful payment
9. THE Order_Detail SHALL provide invoice/receipt download option
10. THE System SHALL support order notes for admin reference

### Requirement 6: 账户余额系统

**User Story:** As a user, I want to maintain account balance, so that I can make purchases quickly.

#### Acceptance Criteria

1. THE User_Portal SHALL display current account balance prominently
2. THE System SHALL support balance top-up via payment methods
3. THE Recharge_Page SHALL offer preset amounts and custom amount input
4. WHEN balance is used for purchase THEN the System SHALL deduct immediately
5. THE System SHALL maintain balance transaction history
6. THE Balance_History SHALL show: type, amount, description, and timestamp
7. THE System SHALL support minimum recharge amount configuration
8. THE Admin_Panel SHALL allow manual balance adjustments with reason
9. THE System SHALL prevent negative balance (no credit)
10. WHEN balance is low THEN the System SHALL notify user (configurable threshold)

### Requirement 7: 续费系统

**User Story:** As a user, I want to renew my subscription, so that I can continue using the service.

#### Acceptance Criteria

1. THE User_Portal SHALL display renewal options before subscription expires
2. THE System SHALL support renewing with same plan or upgrading to different plan
3. WHEN renewing same plan THEN the System SHALL extend expiration date
4. WHEN upgrading plan THEN the System SHALL calculate prorated price
5. THE System SHALL send renewal reminder emails (7 days, 3 days, 1 day before expiry)
6. THE User_Portal SHALL show countdown to expiration on dashboard
7. THE System SHALL support auto-renewal option using account balance
8. WHEN auto-renewal is enabled THEN the System SHALL attempt renewal 1 day before expiry
9. IF auto-renewal fails THEN the System SHALL notify user and disable auto-renewal
10. THE System SHALL support renewal discount for loyal users (configurable)

### Requirement 8: 优惠券系统

**User Story:** As an admin, I want to create coupons, so that I can offer promotions to users.

#### Acceptance Criteria

1. THE Admin_Panel SHALL provide interface to create and manage coupons
2. THE Coupon SHALL support: fixed amount discount and percentage discount
3. THE Coupon SHALL have: code, discount value, usage limit, and expiration date
4. THE System SHALL support coupon restrictions: minimum order amount, specific plans only
5. THE System SHALL track coupon usage count and remaining uses
6. THE Coupon SHALL support single-use per user or unlimited uses per user
7. WHEN coupon is applied THEN the System SHALL validate all restrictions
8. IF coupon is invalid THEN the System SHALL display specific error message
9. THE Admin_Panel SHALL show coupon usage statistics
10. THE System SHALL support generating batch coupon codes

### Requirement 9: 邀请系统

**User Story:** As a user, I want to invite friends, so that I can earn rewards.

#### Acceptance Criteria

1. THE System SHALL generate unique invite code for each user
2. THE User_Portal SHALL display user's invite code and invite link
3. THE Invite_Link SHALL include referral tracking parameter
4. WHEN new user registers with invite code THEN the System SHALL record the referral
5. THE User_Portal SHALL display list of invited users (without sensitive info)
6. THE System SHALL support invite code customization (premium feature, optional)
7. THE Invite_Page SHALL provide easy sharing options (copy link, QR code)
8. THE System SHALL track invite statistics: total invites, successful registrations, conversions
9. THE Admin_Panel SHALL display referral network and statistics
10. THE System SHALL prevent self-referral and circular referrals

### Requirement 10: 邀请奖励

**User Story:** As a user, I want to receive rewards for inviting friends, so that I am motivated to promote the service.

#### Acceptance Criteria

1. THE System SHALL support commission-based rewards (percentage of invitee's purchases)
2. THE System SHALL support fixed bonus rewards (one-time bonus per successful invite)
3. THE System SHALL support traffic bonus rewards (extra traffic for inviter)
4. THE Admin_Panel SHALL allow configuring reward types and amounts
5. WHEN invitee makes first purchase THEN the System SHALL credit reward to inviter
6. THE Commission SHALL be credited to inviter's balance after configurable delay
7. THE User_Portal SHALL display pending and confirmed rewards
8. THE System SHALL support multi-level referral (optional, configurable depth)
9. THE Admin_Panel SHALL allow setting minimum withdrawal amount for commissions
10. THE System SHALL generate reward reports for tax/accounting purposes

### Requirement 11: 发票管理

**User Story:** As a user, I want to download invoices, so that I can keep records of my purchases.

#### Acceptance Criteria

1. THE System SHALL generate invoice for each completed order
2. THE Invoice SHALL include: invoice number, order details, payment info, and timestamps
3. THE User_Portal SHALL allow downloading invoice as PDF
4. THE Invoice SHALL include configurable business information (name, address, tax ID)
5. THE Admin_Panel SHALL allow customizing invoice template
6. THE System SHALL support invoice numbering format configuration
7. THE User_Portal SHALL display invoice history with download links
8. THE System SHALL support batch invoice download for date range
9. THE Invoice SHALL be generated in user's preferred language
10. THE System SHALL store invoices for at least 2 years

### Requirement 12: 财务报表

**User Story:** As an admin, I want to view financial reports, so that I can track business performance.

#### Acceptance Criteria

1. THE Admin_Panel SHALL display revenue dashboard with key metrics
2. THE Dashboard SHALL show: total revenue, order count, average order value
3. THE System SHALL provide daily, weekly, monthly, and yearly revenue reports
4. THE Reports SHALL include breakdown by plan, payment method, and user segment
5. THE Admin_Panel SHALL display revenue charts and trends
6. THE System SHALL track refund amounts and refund rate
7. THE Reports SHALL support export to CSV/Excel format
8. THE Admin_Panel SHALL show commission payouts and pending commissions
9. THE System SHALL provide user lifetime value (LTV) analysis
10. THE Reports SHALL support custom date range filtering

### Requirement 13: 退款处理

**User Story:** As an admin, I want to process refunds, so that I can handle customer complaints.

#### Acceptance Criteria

1. THE Admin_Panel SHALL allow processing refunds for paid orders
2. THE Refund SHALL support full refund and partial refund
3. WHEN refund is processed THEN the System SHALL update order status to refunded
4. THE System SHALL support refund to original payment method or account balance
5. WHEN refunding to balance THEN the System SHALL credit immediately
6. THE System SHALL require refund reason for audit purposes
7. THE Refund SHALL deduct any commission already paid to referrer
8. THE Admin_Panel SHALL display refund history and statistics
9. THE System SHALL send refund confirmation email to user
10. THE System SHALL support refund request from user (requires admin approval)

### Requirement 14: 安全和合规

**User Story:** As an admin, I want the payment system to be secure, so that user data is protected.

#### Acceptance Criteria

1. THE System SHALL encrypt all payment-related data at rest and in transit
2. THE System SHALL not store full credit card numbers (PCI compliance)
3. THE System SHALL implement payment signature verification for callbacks
4. THE System SHALL log all payment transactions for audit
5. THE System SHALL implement rate limiting on payment endpoints
6. THE Admin_Panel SHALL require additional authentication for financial operations
7. THE System SHALL support IP whitelist for payment callbacks
8. THE System SHALL implement idempotency for payment processing
9. THE System SHALL handle currency conversion if supporting multiple currencies
10. THE System SHALL comply with relevant financial regulations (configurable per region)


### Requirement 15: 套餐试用

**User Story:** As a user, I want to try the service before purchasing, so that I can evaluate if it meets my needs.

#### Acceptance Criteria

1. THE Admin_Panel SHALL allow configuring trial plan settings
2. THE Trial_Plan SHALL include: duration (days), traffic limit, and feature restrictions
3. THE System SHALL allow each user only one trial period (tracked by email/device)
4. WHEN user registers THEN the System SHALL optionally auto-activate trial
5. THE User_Portal SHALL display trial status with remaining time and traffic
6. WHEN trial expires THEN the System SHALL prompt user to purchase a plan
7. THE Trial SHALL support requiring email verification before activation
8. THE Admin_Panel SHALL allow manually granting trial to specific users
9. THE System SHALL track trial conversion rate (trial to paid)
10. THE Trial_User SHALL have limited access to premium features (configurable)

### Requirement 16: 套餐升降级

**User Story:** As a user, I want to upgrade or downgrade my plan, so that I can adjust my service level as needed.

#### Acceptance Criteria

1. THE User_Portal SHALL display upgrade options on current subscription page
2. WHEN upgrading plan THEN the System SHALL calculate prorated price based on remaining days
3. THE Proration_Formula SHALL be: (new_price - old_price) * (remaining_days / total_days)
4. WHEN downgrading plan THEN the System SHALL apply change at next billing cycle
5. THE System SHALL support immediate downgrade with prorated refund to balance (optional)
6. THE Upgrade_Page SHALL clearly show price difference and new features
7. WHEN upgrading THEN the System SHALL preserve remaining traffic if new plan has more
8. THE System SHALL send confirmation email after plan change
9. THE Admin_Panel SHALL display plan change history for each user
10. THE System SHALL support upgrade-only restrictions (no downgrade for certain plans)

### Requirement 17: 支付失败处理

**User Story:** As a user, I want failed payments to be handled gracefully, so that I don't lose my order.

#### Acceptance Criteria

1. WHEN payment fails THEN the System SHALL preserve the order in pending status
2. THE System SHALL support automatic payment retry (configurable, default 3 times)
3. THE Retry_Interval SHALL be configurable (default: 1 hour, 4 hours, 24 hours)
4. WHEN all retries fail THEN the System SHALL notify user and cancel order
5. THE User_Portal SHALL display payment failure reason if available
6. THE System SHALL support switching payment method for failed orders
7. WHEN payment gateway is unavailable THEN the System SHALL queue payment for later
8. THE Admin_Panel SHALL display failed payment statistics
9. THE System SHALL log all payment attempts with failure reasons
10. WHEN auto-renewal payment fails THEN the System SHALL notify user immediately

### Requirement 18: 多币种支持

**User Story:** As an international user, I want to pay in my local currency, so that I can avoid currency conversion fees.

#### Acceptance Criteria

1. THE Admin_Panel SHALL allow configuring supported currencies
2. THE Plan SHALL support setting prices in multiple currencies
3. THE System SHALL auto-detect user's preferred currency based on location
4. THE User_Portal SHALL allow manually selecting currency
5. WHEN displaying prices THEN the System SHALL show in user's selected currency
6. THE System SHALL use reliable exchange rate API for conversion (if needed)
7. THE Exchange_Rate SHALL be cached and updated periodically (configurable)
8. THE Order SHALL record the currency and exchange rate at time of purchase
9. THE Financial_Reports SHALL support filtering and grouping by currency
10. THE System SHALL handle currency symbol and formatting per locale

### Requirement 19: 订阅暂停

**User Story:** As a user, I want to pause my subscription, so that I don't waste service time when I'm not using it.

#### Acceptance Criteria

1. THE User_Portal SHALL allow pausing active subscription
2. THE Pause SHALL freeze expiration countdown and traffic reset
3. THE System SHALL limit pause duration (configurable, default max 30 days)
4. THE System SHALL limit pause frequency (configurable, default once per billing cycle)
5. WHEN subscription is paused THEN the System SHALL disable proxy access
6. THE User_Portal SHALL allow resuming subscription at any time
7. WHEN resuming THEN the System SHALL restore remaining days and traffic
8. THE Admin_Panel SHALL allow configuring which plans support pause feature
9. THE System SHALL send notification before auto-resume (if max pause reached)
10. THE Admin_Panel SHALL display pause statistics and abuse patterns

### Requirement 20: 礼品卡系统

**User Story:** As a user, I want to purchase and redeem gift cards, so that I can gift service to others.

#### Acceptance Criteria

1. THE Admin_Panel SHALL allow creating gift card batches
2. THE Gift_Card SHALL include: code, value, expiration date, and status
3. THE User_Portal SHALL allow purchasing gift cards as products
4. THE Gift_Card SHALL be delivered via email with redemption instructions
5. THE User_Portal SHALL provide gift card redemption page
6. WHEN gift card is redeemed THEN the System SHALL credit value to user's balance
7. THE Gift_Card SHALL support partial redemption (optional)
8. THE System SHALL track gift card usage: purchaser, recipient, redemption date
9. THE Admin_Panel SHALL display gift card inventory and sales statistics
10. THE Gift_Card_Code SHALL be unique and cryptographically secure
