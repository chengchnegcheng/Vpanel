// Package refund provides refund management functionality.
package refund

import (
	"context"
	"errors"
	"fmt"
	"time"

	"v/internal/commercial/balance"
	"v/internal/commercial/commission"
	"v/internal/database/repository"
	"v/internal/logger"
)

// Common errors
var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderNotRefundable  = errors.New("order is not refundable")
	ErrInvalidRefundAmount = errors.New("invalid refund amount")
	ErrRefundExceedsAmount = errors.New("refund amount exceeds order amount")
	ErrAlreadyRefunded     = errors.New("order already refunded")
)

// Order status constants
const (
	StatusPaid      = "paid"
	StatusCompleted = "completed"
	StatusRefunded  = "refunded"
)

// RefundRequest represents a refund request.
type RefundRequest struct {
	OrderID int64  `json:"order_id"`
	Amount  int64  `json:"amount"` // 0 = full refund
	Reason  string `json:"reason"`
}

// RefundResult represents the result of a refund operation.
type RefundResult struct {
	OrderID          int64  `json:"order_id"`
	RefundAmount     int64  `json:"refund_amount"`
	BalanceRestored  int64  `json:"balance_restored"`
	CommissionCancel int64  `json:"commission_cancelled"`
	Status           string `json:"status"`
}

// Service provides refund management operations.
type Service struct {
	orderRepo         repository.OrderRepository
	balanceService    *balance.Service
	commissionService *commission.Service
	logger            logger.Logger
}

// NewService creates a new refund service.
func NewService(
	orderRepo repository.OrderRepository,
	balanceService *balance.Service,
	commissionService *commission.Service,
	log logger.Logger,
) *Service {
	return &Service{
		orderRepo:         orderRepo,
		balanceService:    balanceService,
		commissionService: commissionService,
		logger:            log,
	}
}

// ProcessRefund processes a refund for an order.
func (s *Service) ProcessRefund(ctx context.Context, req *RefundRequest) (*RefundResult, error) {
	// Get order
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	// Check if order is refundable
	if !s.isRefundable(order.Status) {
		return nil, ErrOrderNotRefundable
	}

	// Determine refund amount
	refundAmount := req.Amount
	if refundAmount == 0 {
		// Full refund
		refundAmount = order.PayAmount + order.BalanceUsed
	}

	// Validate refund amount
	if refundAmount < 0 {
		return nil, ErrInvalidRefundAmount
	}

	totalPaid := order.PayAmount + order.BalanceUsed
	if refundAmount > totalPaid {
		return nil, ErrRefundExceedsAmount
	}

	result := &RefundResult{
		OrderID:      req.OrderID,
		RefundAmount: refundAmount,
	}

	// Restore balance
	if s.balanceService != nil {
		// Calculate how much to restore to balance
		// Priority: restore balance used first, then payment amount
		balanceToRestore := refundAmount
		if balanceToRestore > totalPaid {
			balanceToRestore = totalPaid
		}

		if balanceToRestore > 0 {
			desc := fmt.Sprintf("Refund for order #%d", order.ID)
			if req.Reason != "" {
				desc = fmt.Sprintf("%s: %s", desc, req.Reason)
			}

			orderIDPtr := &order.ID
			if err := s.balanceService.Refund(ctx, order.UserID, balanceToRestore, orderIDPtr, desc); err != nil {
				s.logger.Error("Failed to restore balance for refund",
					logger.Err(err),
					logger.F("orderID", req.OrderID),
					logger.F("amount", balanceToRestore))
				// Continue with refund even if balance restoration fails
			} else {
				result.BalanceRestored = balanceToRestore
			}
		}
	}

	// Cancel commissions
	if s.commissionService != nil {
		if err := s.commissionService.CancelByOrder(ctx, req.OrderID); err != nil {
			s.logger.Error("Failed to cancel commissions for refund",
				logger.Err(err),
				logger.F("orderID", req.OrderID))
			// Continue with refund even if commission cancellation fails
		}
	}

	// Update order status
	if err := s.orderRepo.UpdateStatus(ctx, req.OrderID, StatusRefunded); err != nil {
		s.logger.Error("Failed to update order status to refunded",
			logger.Err(err),
			logger.F("orderID", req.OrderID))
		return nil, err
	}

	// Update order notes
	notes := fmt.Sprintf("Refunded %d cents at %s", refundAmount, time.Now().Format("2006-01-02 15:04:05"))
	if req.Reason != "" {
		notes = fmt.Sprintf("%s. Reason: %s", notes, req.Reason)
	}
	order.Notes = notes
	if err := s.orderRepo.Update(ctx, order); err != nil {
		s.logger.Warn("Failed to update order notes", logger.Err(err))
	}

	result.Status = StatusRefunded

	s.logger.Info("Refund processed",
		logger.F("orderID", req.OrderID),
		logger.F("refundAmount", refundAmount),
		logger.F("balanceRestored", result.BalanceRestored))

	return result, nil
}

// ProcessPartialRefund processes a partial refund for an order.
func (s *Service) ProcessPartialRefund(ctx context.Context, orderID int64, amount int64, reason string) (*RefundResult, error) {
	if amount <= 0 {
		return nil, ErrInvalidRefundAmount
	}

	return s.ProcessRefund(ctx, &RefundRequest{
		OrderID: orderID,
		Amount:  amount,
		Reason:  reason,
	})
}

// ProcessFullRefund processes a full refund for an order.
func (s *Service) ProcessFullRefund(ctx context.Context, orderID int64, reason string) (*RefundResult, error) {
	return s.ProcessRefund(ctx, &RefundRequest{
		OrderID: orderID,
		Amount:  0, // 0 means full refund
		Reason:  reason,
	})
}

// isRefundable checks if an order status allows refund.
func (s *Service) isRefundable(status string) bool {
	return status == StatusPaid || status == StatusCompleted
}

// CanRefund checks if an order can be refunded.
func (s *Service) CanRefund(ctx context.Context, orderID int64) (bool, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return false, ErrOrderNotFound
	}

	return s.isRefundable(order.Status), nil
}

// GetMaxRefundAmount returns the maximum refundable amount for an order.
func (s *Service) GetMaxRefundAmount(ctx context.Context, orderID int64) (int64, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return 0, ErrOrderNotFound
	}

	if !s.isRefundable(order.Status) {
		return 0, ErrOrderNotRefundable
	}

	return order.PayAmount + order.BalanceUsed, nil
}
