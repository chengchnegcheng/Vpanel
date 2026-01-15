// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"v/internal/commercial/order"
	"v/internal/logger"
)

// ReportHandler handles report-related requests.
type ReportHandler struct {
	orderService *order.Service
	logger       logger.Logger
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(orderService *order.Service, log logger.Logger) *ReportHandler {
	return &ReportHandler{orderService: orderService, logger: log}
}

// GetRevenueReport returns revenue statistics (admin only).
func (h *ReportHandler) GetRevenueReport(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	
	var start, end time.Time
	var err error
	
	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
	} else {
		start = time.Now().AddDate(0, -1, 0) // Default: last month
	}
	
	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
	} else {
		end = time.Now()
	}
	
	revenue, err := h.orderService.GetRevenueByDateRange(c.Request.Context(), start, end)
	if err != nil {
		h.logger.Error("Failed to get revenue", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get revenue"})
		return
	}
	
	orderCount, err := h.orderService.GetOrderCountByDateRange(c.Request.Context(), start, end)
	if err != nil {
		h.logger.Error("Failed to get order count", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order count"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"revenue":     revenue,
		"order_count": orderCount,
		"start":       start.Format("2006-01-02"),
		"end":         end.Format("2006-01-02"),
	})
}

// GetOrderStats returns order statistics (admin only).
func (h *ReportHandler) GetOrderStats(c *gin.Context) {
	ctx := c.Request.Context()
	
	total, _ := h.orderService.GetOrderCount(ctx)
	pending, _ := h.orderService.GetOrderCountByStatus(ctx, "pending")
	paid, _ := h.orderService.GetOrderCountByStatus(ctx, "paid")
	completed, _ := h.orderService.GetOrderCountByStatus(ctx, "completed")
	cancelled, _ := h.orderService.GetOrderCountByStatus(ctx, "cancelled")
	refunded, _ := h.orderService.GetOrderCountByStatus(ctx, "refunded")
	
	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"pending":   pending,
		"paid":      paid,
		"completed": completed,
		"cancelled": cancelled,
		"refunded":  refunded,
	})
}
