// Package balance provides balance management functionality.
package balance

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"v/internal/database/repository"
	"v/internal/logger"
)

// Common errors
var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrUserNotFound        = errors.New("user not found")
	ErrNegativeBalance     = errors.New("balance cannot be negative")
)

// Transaction type constants
const (
	TxTypeRecharge   = repository.BalanceTxTypeRecharge
	TxTypePurchase   = repository.BalanceTxTypePurchase
	TxTypeRefund     = repository.BalanceTxTypeRefund
	TxTypeCommission = repository.BalanceTxTypeCommission
	TxTypeAdjustment = repository.BalanceTxTypeAdjustment
)

// Transaction represents a balance transaction.
type Transaction struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Type        string `json:"type"`
	Amount      int64  `json:"amount"`
	Balance     int64  `json:"balance"`
	OrderID     *int64 `json:"order_id"`
	Description string `json:"description"`
	Operator    string `json:"operator"`
	CreatedAt   string `json:"created_at"`
}

// Service provides balance management operations.
type Service struct {
	balanceRepo repository.BalanceRepository
	logger      logger.Logger
	mu          sync.Mutex
}

// NewService creates a new balance service.
func NewService(balanceRepo repository.BalanceRepository, log logger.Logger) *Service {
	return &Service{
		balanceRepo: balanceRepo,
		logger:      log,
	}
}


// GetBalance retrieves the current balance for a user.
func (s *Service) GetBalance(ctx context.Context, userID int64) (int64, error) {
	balance, err := s.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get balance", logger.Err(err), logger.F("userID", userID))
		return 0, err
	}
	return balance, nil
}

// CanDeduct checks if a user has sufficient balance for a deduction.
func (s *Service) CanDeduct(ctx context.Context, userID int64, amount int64) bool {
	if amount <= 0 {
		return false
	}
	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return false
	}
	return balance >= amount
}

// Recharge adds funds to a user's balance.
func (s *Service) Recharge(ctx context.Context, userID int64, amount int64, orderID *int64, description string) error {
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidAmount)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current balance
	currentBalance, err := s.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get balance for recharge", logger.Err(err), logger.F("userID", userID))
		return err
	}

	newBalance := currentBalance + amount

	// Update balance
	if err := s.balanceRepo.IncrementBalance(ctx, userID, amount); err != nil {
		s.logger.Error("Failed to increment balance", logger.Err(err), logger.F("userID", userID))
		return err
	}

	// Record transaction
	tx := &repository.BalanceTransaction{
		UserID:      userID,
		Type:        TxTypeRecharge,
		Amount:      amount,
		Balance:     newBalance,
		OrderID:     orderID,
		Description: description,
		Operator:    "system",
	}

	if err := s.balanceRepo.CreateTransaction(ctx, tx); err != nil {
		s.logger.Error("Failed to create recharge transaction", logger.Err(err))
		return err
	}

	s.logger.Info("Balance recharged", logger.F("userID", userID), logger.F("amount", amount), logger.F("newBalance", newBalance))
	return nil
}

// Deduct subtracts funds from a user's balance.
func (s *Service) Deduct(ctx context.Context, userID int64, amount int64, orderID *int64, description string) error {
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidAmount)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current balance
	currentBalance, err := s.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get balance for deduction", logger.Err(err), logger.F("userID", userID))
		return err
	}

	// Check sufficient balance
	if currentBalance < amount {
		return ErrInsufficientBalance
	}

	newBalance := currentBalance - amount

	// Ensure non-negative balance
	if newBalance < 0 {
		return ErrNegativeBalance
	}

	// Update balance
	if err := s.balanceRepo.DecrementBalance(ctx, userID, amount); err != nil {
		s.logger.Error("Failed to decrement balance", logger.Err(err), logger.F("userID", userID))
		return err
	}

	// Record transaction (negative amount for deduction)
	tx := &repository.BalanceTransaction{
		UserID:      userID,
		Type:        TxTypePurchase,
		Amount:      -amount,
		Balance:     newBalance,
		OrderID:     orderID,
		Description: description,
		Operator:    "system",
	}

	if err := s.balanceRepo.CreateTransaction(ctx, tx); err != nil {
		s.logger.Error("Failed to create deduction transaction", logger.Err(err))
		return err
	}

	s.logger.Info("Balance deducted", logger.F("userID", userID), logger.F("amount", amount), logger.F("newBalance", newBalance))
	return nil
}

// Refund adds refunded funds back to a user's balance.
func (s *Service) Refund(ctx context.Context, userID int64, amount int64, orderID *int64, description string) error {
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidAmount)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current balance
	currentBalance, err := s.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get balance for refund", logger.Err(err), logger.F("userID", userID))
		return err
	}

	newBalance := currentBalance + amount

	// Update balance
	if err := s.balanceRepo.IncrementBalance(ctx, userID, amount); err != nil {
		s.logger.Error("Failed to increment balance for refund", logger.Err(err), logger.F("userID", userID))
		return err
	}

	// Record transaction
	tx := &repository.BalanceTransaction{
		UserID:      userID,
		Type:        TxTypeRefund,
		Amount:      amount,
		Balance:     newBalance,
		OrderID:     orderID,
		Description: description,
		Operator:    "system",
	}

	if err := s.balanceRepo.CreateTransaction(ctx, tx); err != nil {
		s.logger.Error("Failed to create refund transaction", logger.Err(err))
		return err
	}

	s.logger.Info("Balance refunded", logger.F("userID", userID), logger.F("amount", amount), logger.F("newBalance", newBalance))
	return nil
}

// AddCommission adds commission to a user's balance.
func (s *Service) AddCommission(ctx context.Context, userID int64, amount int64, description string) error {
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidAmount)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current balance
	currentBalance, err := s.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get balance for commission", logger.Err(err), logger.F("userID", userID))
		return err
	}

	newBalance := currentBalance + amount

	// Update balance
	if err := s.balanceRepo.IncrementBalance(ctx, userID, amount); err != nil {
		s.logger.Error("Failed to increment balance for commission", logger.Err(err), logger.F("userID", userID))
		return err
	}

	// Record transaction
	tx := &repository.BalanceTransaction{
		UserID:      userID,
		Type:        TxTypeCommission,
		Amount:      amount,
		Balance:     newBalance,
		Description: description,
		Operator:    "system",
	}

	if err := s.balanceRepo.CreateTransaction(ctx, tx); err != nil {
		s.logger.Error("Failed to create commission transaction", logger.Err(err))
		return err
	}

	s.logger.Info("Commission added", logger.F("userID", userID), logger.F("amount", amount), logger.F("newBalance", newBalance))
	return nil
}

// Adjust manually adjusts a user's balance (admin operation).
func (s *Service) Adjust(ctx context.Context, userID int64, amount int64, reason string, operator string) error {
	if amount == 0 {
		return fmt.Errorf("%w: amount cannot be zero", ErrInvalidAmount)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current balance
	currentBalance, err := s.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get balance for adjustment", logger.Err(err), logger.F("userID", userID))
		return err
	}

	newBalance := currentBalance + amount

	// Ensure non-negative balance
	if newBalance < 0 {
		return ErrNegativeBalance
	}

	// Update balance
	if amount > 0 {
		if err := s.balanceRepo.IncrementBalance(ctx, userID, amount); err != nil {
			s.logger.Error("Failed to increment balance for adjustment", logger.Err(err), logger.F("userID", userID))
			return err
		}
	} else {
		if err := s.balanceRepo.DecrementBalance(ctx, userID, -amount); err != nil {
			s.logger.Error("Failed to decrement balance for adjustment", logger.Err(err), logger.F("userID", userID))
			return err
		}
	}

	// Record transaction
	tx := &repository.BalanceTransaction{
		UserID:      userID,
		Type:        TxTypeAdjustment,
		Amount:      amount,
		Balance:     newBalance,
		Description: reason,
		Operator:    operator,
	}

	if err := s.balanceRepo.CreateTransaction(ctx, tx); err != nil {
		s.logger.Error("Failed to create adjustment transaction", logger.Err(err))
		return err
	}

	s.logger.Info("Balance adjusted", logger.F("userID", userID), logger.F("amount", amount), logger.F("newBalance", newBalance), logger.F("operator", operator))
	return nil
}

// GetTransactions retrieves transaction history for a user.
func (s *Service) GetTransactions(ctx context.Context, userID int64, page, pageSize int) ([]*Transaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	repoTxs, total, err := s.balanceRepo.ListByUser(ctx, userID, pageSize, offset)
	if err != nil {
		s.logger.Error("Failed to list transactions", logger.Err(err), logger.F("userID", userID))
		return nil, 0, err
	}

	txs := make([]*Transaction, len(repoTxs))
	for i, rt := range repoTxs {
		txs[i] = s.toTransaction(rt)
	}

	return txs, total, nil
}

// GetStatistics retrieves balance statistics for a user.
func (s *Service) GetStatistics(ctx context.Context, userID int64) (totalRecharge, totalSpent, totalCommission int64, err error) {
	totalRecharge, err = s.balanceRepo.GetTotalRecharge(ctx, userID)
	if err != nil {
		return 0, 0, 0, err
	}

	totalSpent, err = s.balanceRepo.GetTotalSpent(ctx, userID)
	if err != nil {
		return 0, 0, 0, err
	}

	totalCommission, err = s.balanceRepo.GetTotalCommission(ctx, userID)
	if err != nil {
		return 0, 0, 0, err
	}

	return totalRecharge, totalSpent, totalCommission, nil
}

// toTransaction converts a repository transaction to a service transaction.
func (s *Service) toTransaction(rt *repository.BalanceTransaction) *Transaction {
	return &Transaction{
		ID:          rt.ID,
		UserID:      rt.UserID,
		Type:        rt.Type,
		Amount:      rt.Amount,
		Balance:     rt.Balance,
		OrderID:     rt.OrderID,
		Description: rt.Description,
		Operator:    rt.Operator,
		CreatedAt:   rt.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
