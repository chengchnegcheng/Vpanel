// Package ticket provides ticket management services for the user portal.
package ticket

import (
	"context"
	"strings"
	"time"

	"v/internal/database/repository"
	"v/pkg/errors"
)

// Service provides ticket operations for the user portal.
type Service struct {
	ticketRepo repository.TicketRepository
	userRepo   repository.UserRepository
}

// NewService creates a new ticket service.
func NewService(ticketRepo repository.TicketRepository, userRepo repository.UserRepository) *Service {
	return &Service{
		ticketRepo: ticketRepo,
		userRepo:   userRepo,
	}
}

// CreateTicketRequest represents a request to create a ticket.
type CreateTicketRequest struct {
	Subject  string `json:"subject"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Priority string `json:"priority"`
}

// Validate validates the create ticket request.
func (r *CreateTicketRequest) Validate() error {
	r.Subject = strings.TrimSpace(r.Subject)
	if r.Subject == "" {
		return errors.NewValidationError("subject", "工单主题不能为空")
	}
	if len(r.Subject) > 200 {
		return errors.NewValidationError("subject", "工单主题不能超过200个字符")
	}

	r.Content = strings.TrimSpace(r.Content)
	if r.Content == "" {
		return errors.NewValidationError("content", "工单内容不能为空")
	}
	if len(r.Content) > 10000 {
		return errors.NewValidationError("content", "工单内容不能超过10000个字符")
	}

	// Validate category
	validCategories := map[string]bool{
		"technical": true,
		"billing":   true,
		"account":   true,
		"other":     true,
	}
	if r.Category == "" {
		r.Category = "other"
	}
	if !validCategories[r.Category] {
		return errors.NewValidationError("category", "无效的工单分类")
	}

	// Validate priority
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}
	if r.Priority == "" {
		r.Priority = "medium"
	}
	if !validPriorities[r.Priority] {
		return errors.NewValidationError("priority", "无效的优先级")
	}

	return nil
}


// TicketResult represents a ticket with its messages.
type TicketResult struct {
	*repository.Ticket
	Messages []*repository.TicketMessage `json:"messages,omitempty"`
}

// CreateTicket creates a new ticket.
func (s *Service) CreateTicket(ctx context.Context, userID int64, req *CreateTicketRequest) (*repository.Ticket, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	ticket := &repository.Ticket{
		UserID:   userID,
		Subject:  req.Subject,
		Category: req.Category,
		Priority: req.Priority,
		Status:   repository.TicketStatusOpen,
	}

	if err := s.ticketRepo.Create(ctx, ticket); err != nil {
		return nil, err
	}

	// Add the initial message
	userIDPtr := userID
	message := &repository.TicketMessage{
		TicketID: ticket.ID,
		UserID:   &userIDPtr,
		Content:  req.Content,
		IsAdmin:  false,
	}

	if err := s.ticketRepo.AddMessage(ctx, message); err != nil {
		return nil, err
	}

	return ticket, nil
}

// GetTicket retrieves a ticket by ID for a user.
func (s *Service) GetTicket(ctx context.Context, userID int64, ticketID int64) (*TicketResult, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if ticket.UserID != userID {
		return nil, errors.NewForbiddenError("无权访问此工单")
	}

	// Get messages
	messages, err := s.ticketRepo.GetMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	return &TicketResult{
		Ticket:   ticket,
		Messages: messages,
	}, nil
}

// ListTickets lists tickets for a user.
func (s *Service) ListTickets(ctx context.Context, userID int64, status string, limit, offset int) ([]*repository.Ticket, int64, error) {
	filter := &repository.TicketFilter{
		UserID: &userID,
		Limit:  limit,
		Offset: offset,
	}

	if status != "" {
		filter.Status = &status
	}

	return s.ticketRepo.ListByUser(ctx, userID, filter)
}

// ReplyTicketRequest represents a request to reply to a ticket.
type ReplyTicketRequest struct {
	Content string `json:"content"`
}

// Validate validates the reply request.
func (r *ReplyTicketRequest) Validate() error {
	r.Content = strings.TrimSpace(r.Content)
	if r.Content == "" {
		return errors.NewValidationError("content", "回复内容不能为空")
	}
	if len(r.Content) > 10000 {
		return errors.NewValidationError("content", "回复内容不能超过10000个字符")
	}
	return nil
}

// ReplyTicket adds a reply to a ticket.
func (s *Service) ReplyTicket(ctx context.Context, userID int64, ticketID int64, req *ReplyTicketRequest) (*repository.TicketMessage, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get ticket and verify ownership
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.UserID != userID {
		return nil, errors.NewForbiddenError("无权访问此工单")
	}

	// Check if ticket is closed
	if ticket.Status == repository.TicketStatusClosed {
		return nil, errors.NewValidationError("status", "工单已关闭，无法回复")
	}

	// Add message
	userIDPtr := userID
	message := &repository.TicketMessage{
		TicketID: ticketID,
		UserID:   &userIDPtr,
		Content:  req.Content,
		IsAdmin:  false,
	}

	if err := s.ticketRepo.AddMessage(ctx, message); err != nil {
		return nil, err
	}

	// Update ticket status to waiting if it was answered
	if ticket.Status == repository.TicketStatusAnswered {
		ticket.Status = repository.TicketStatusWaiting
		if err := s.ticketRepo.Update(ctx, ticket); err != nil {
			return nil, err
		}
	}

	return message, nil
}

// CloseTicket closes a ticket.
func (s *Service) CloseTicket(ctx context.Context, userID int64, ticketID int64) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.UserID != userID {
		return errors.NewForbiddenError("无权访问此工单")
	}

	if ticket.Status == repository.TicketStatusClosed {
		return errors.NewValidationError("status", "工单已关闭")
	}

	now := time.Now()
	ticket.Status = repository.TicketStatusClosed
	ticket.ClosedAt = &now

	return s.ticketRepo.Update(ctx, ticket)
}

// ReopenTicket reopens a closed ticket.
func (s *Service) ReopenTicket(ctx context.Context, userID int64, ticketID int64) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.UserID != userID {
		return errors.NewForbiddenError("无权访问此工单")
	}

	if ticket.Status != repository.TicketStatusClosed {
		return errors.NewValidationError("status", "工单未关闭")
	}

	ticket.Status = repository.TicketStatusOpen
	ticket.ClosedAt = nil

	return s.ticketRepo.Update(ctx, ticket)
}

// TicketStatusTransition defines valid status transitions.
var TicketStatusTransition = map[string][]string{
	repository.TicketStatusOpen:     {repository.TicketStatusWaiting, repository.TicketStatusAnswered, repository.TicketStatusClosed},
	repository.TicketStatusWaiting:  {repository.TicketStatusAnswered, repository.TicketStatusClosed},
	repository.TicketStatusAnswered: {repository.TicketStatusWaiting, repository.TicketStatusClosed},
	repository.TicketStatusClosed:   {repository.TicketStatusOpen},
}

// CanTransition checks if a status transition is valid.
func CanTransition(from, to string) bool {
	validTransitions, ok := TicketStatusTransition[from]
	if !ok {
		return false
	}
	for _, valid := range validTransitions {
		if valid == to {
			return true
		}
	}
	return false
}
