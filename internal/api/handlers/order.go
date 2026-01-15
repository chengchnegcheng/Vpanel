// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"v/internal/commercial/order"
	"v/internal/logger"
)

// OrderHandler handles order-related requests.
type OrderHandler struct {
	orderService *order.Service
	logger       logger.Logger
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(orderService *order.Service, log logger.Logger) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		logger:       log,
	}
}

// OrderResponse represents an order in API responses.
type OrderResponse struct {
	ID             int64   `json:"id"`
	OrderNo        string  `json:"order_no"`
	UserID         int64   `json:"user_id"`
	PlanID         int64   `json:"plan_id"`
	CouponID       *int64  `json:"coupon_id,omitempty"`
	OriginalAmount int64   `json:"original_amount"`
	DiscountAmount int64   `json:"discount_amount"`
	BalanceUsed    int64   `json:"balance_used"`
	PayAmount      int64   `json:"pay_amount"`
	Status         string  `json:"status"`
	PaymentMethod  string  `json:"payment_method,omitempty"`
	PaymentNo      string  `json:"payment_no,omitempty"`
	PaidAt         *string `json:"paid_at,omitempty"`
	ExpiredAt      string  `json:"expired_at"`
	CreatedAt      string  `json:"created_at"`
}

// CreateOrderRequest represents a request to create an order.
type CreateOrderRequest struct {
	PlanID     int64  `json:"plan_id" binding:"required"`
	CouponCode string `json:"coupon_code"`
}


// CreateOrder creates a new order.
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	createReq := &order.CreateOrderRequest{
		UserID:     userID.(int64),
		PlanID:     req.PlanID,
		CouponCode: req.CouponCode,
	}

	o, err := h.orderService.Create(c.Request.Context(), createReq)
	if err != nil {
		h.logger.Error("Failed to create order", logger.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"order": h.toOrderResponse(o)})
}

// GetOrder returns an order by ID.
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	o, err := h.orderService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Check if user owns this order (unless admin)
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")
	if role != "admin" && o.UserID != userID.(int64) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": h.toOrderResponse(o)})
}

// ListUserOrders returns orders for the current user.
func (h *OrderHandler) ListUserOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orders, total, err := h.orderService.ListByUser(c.Request.Context(), userID.(int64), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list orders", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}

	response := make([]OrderResponse, len(orders))
	for i, o := range orders {
		response[i] = h.toOrderResponse(o)
	}

	c.JSON(http.StatusOK, gin.H{"orders": response, "total": total, "page": page, "page_size": pageSize})
}

// CancelOrder cancels a pending order.
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Verify ownership
	o, err := h.orderService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	userID, _ := c.Get("userID")
	role, _ := c.Get("role")
	if role != "admin" && o.UserID != userID.(int64) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.orderService.Cancel(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to cancel order", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled"})
}


// ListAllOrders returns all orders (admin only).
func (h *OrderHandler) ListAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	filter := order.OrderFilter{Status: status}
	orders, total, err := h.orderService.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list orders", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}

	response := make([]OrderResponse, len(orders))
	for i, o := range orders {
		response[i] = h.toOrderResponse(o)
	}

	c.JSON(http.StatusOK, gin.H{"orders": response, "total": total, "page": page, "page_size": pageSize})
}

// UpdateOrderStatus updates an order's status (admin only).
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.orderService.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		h.logger.Error("Failed to update order status", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated"})
}

func (h *OrderHandler) toOrderResponse(o *order.Order) OrderResponse {
	resp := OrderResponse{
		ID:             o.ID,
		OrderNo:        o.OrderNo,
		UserID:         o.UserID,
		PlanID:         o.PlanID,
		CouponID:       o.CouponID,
		OriginalAmount: o.OriginalAmount,
		DiscountAmount: o.DiscountAmount,
		BalanceUsed:    o.BalanceUsed,
		PayAmount:      o.PayAmount,
		Status:         o.Status,
		PaymentMethod:  o.PaymentMethod,
		PaymentNo:      o.PaymentNo,
		ExpiredAt:      o.ExpiredAt.Format("2006-01-02 15:04:05"),
		CreatedAt:      o.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if o.PaidAt != nil {
		paidAt := o.PaidAt.Format("2006-01-02 15:04:05")
		resp.PaidAt = &paidAt
	}
	return resp
}
