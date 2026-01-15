// Package plan provides plan management functionality.
package plan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"v/internal/database/repository"
	"v/internal/logger"
)

// Common errors
var (
	ErrPlanNotFound   = errors.New("plan not found")
	ErrPlanInactive   = errors.New("plan is not active")
	ErrInvalidPlan    = errors.New("invalid plan data")
	ErrGroupNotFound  = errors.New("plan group not found")
)

// Plan represents a commercial plan with all its attributes.
type Plan struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	TrafficLimit   int64    `json:"traffic_limit"`
	Duration       int      `json:"duration"`
	Price          int64    `json:"price"`
	PlanType       string   `json:"plan_type"`
	ResetCycle     string   `json:"reset_cycle"`
	IPLimit        int      `json:"ip_limit"`
	SortOrder      int      `json:"sort_order"`
	IsActive       bool     `json:"is_active"`
	IsRecommended  bool     `json:"is_recommended"`
	GroupID        *int64   `json:"group_id"`
	PaymentMethods []string `json:"payment_methods"`
	Features       []string `json:"features"`
	MonthlyPrice   int64    `json:"monthly_price"`
}

// PlanGroup represents a plan group.
type PlanGroup struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

// CreatePlanRequest represents a request to create a plan.
type CreatePlanRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	TrafficLimit   int64    `json:"traffic_limit"`
	Duration       int      `json:"duration"`
	Price          int64    `json:"price"`
	PlanType       string   `json:"plan_type"`
	ResetCycle     string   `json:"reset_cycle"`
	IPLimit        int      `json:"ip_limit"`
	SortOrder      int      `json:"sort_order"`
	IsActive       bool     `json:"is_active"`
	IsRecommended  bool     `json:"is_recommended"`
	GroupID        *int64   `json:"group_id"`
	PaymentMethods []string `json:"payment_methods"`
	Features       []string `json:"features"`
}

// UpdatePlanRequest represents a request to update a plan.
type UpdatePlanRequest struct {
	Name           *string   `json:"name"`
	Description    *string   `json:"description"`
	TrafficLimit   *int64    `json:"traffic_limit"`
	Duration       *int      `json:"duration"`
	Price          *int64    `json:"price"`
	PlanType       *string   `json:"plan_type"`
	ResetCycle     *string   `json:"reset_cycle"`
	IPLimit        *int      `json:"ip_limit"`
	SortOrder      *int      `json:"sort_order"`
	IsActive       *bool     `json:"is_active"`
	IsRecommended  *bool     `json:"is_recommended"`
	GroupID        *int64    `json:"group_id"`
	PaymentMethods *[]string `json:"payment_methods"`
	Features       *[]string `json:"features"`
}

// PlanFilter defines filter options for listing plans.
type PlanFilter struct {
	IsActive      *bool
	PlanType      string
	GroupID       *int64
	MinPrice      *int64
	MaxPrice      *int64
	IsRecommended *bool
}

// Service provides plan management operations.
type Service struct {
	planRepo repository.PlanRepository
	logger   logger.Logger
}

// NewService creates a new plan service.
func NewService(planRepo repository.PlanRepository, log logger.Logger) *Service {
	return &Service{
		planRepo: planRepo,
		logger:   log,
	}
}

// Create creates a new plan.
func (s *Service) Create(ctx context.Context, req *CreatePlanRequest) (*Plan, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidPlan)
	}
	if req.Duration <= 0 {
		return nil, fmt.Errorf("%w: duration must be positive", ErrInvalidPlan)
	}
	if req.Price < 0 {
		return nil, fmt.Errorf("%w: price cannot be negative", ErrInvalidPlan)
	}

	// Set defaults
	if req.PlanType == "" {
		req.PlanType = "monthly"
	}
	if req.ResetCycle == "" {
		req.ResetCycle = "monthly"
	}

	paymentMethodsJSON, _ := json.Marshal(req.PaymentMethods)
	featuresJSON, _ := json.Marshal(req.Features)

	repoPlan := &repository.CommercialPlan{
		Name:           req.Name,
		Description:    req.Description,
		TrafficLimit:   req.TrafficLimit,
		Duration:       req.Duration,
		Price:          req.Price,
		PlanType:       req.PlanType,
		ResetCycle:     req.ResetCycle,
		IPLimit:        req.IPLimit,
		SortOrder:      req.SortOrder,
		IsActive:       req.IsActive,
		IsRecommended:  req.IsRecommended,
		GroupID:        req.GroupID,
		PaymentMethods: string(paymentMethodsJSON),
		Features:       string(featuresJSON),
	}

	if err := s.planRepo.Create(ctx, repoPlan); err != nil {
		s.logger.Error("Failed to create plan", logger.Err(err))
		return nil, err
	}

	return s.toPlan(repoPlan), nil
}

// GetByID retrieves a plan by ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*Plan, error) {
	repoPlan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrPlanNotFound
	}
	return s.toPlan(repoPlan), nil
}

// Update updates a plan.
func (s *Service) Update(ctx context.Context, id int64, req *UpdatePlanRequest) (*Plan, error) {
	repoPlan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrPlanNotFound
	}

	if req.Name != nil {
		repoPlan.Name = *req.Name
	}
	if req.Description != nil {
		repoPlan.Description = *req.Description
	}
	if req.TrafficLimit != nil {
		repoPlan.TrafficLimit = *req.TrafficLimit
	}
	if req.Duration != nil {
		repoPlan.Duration = *req.Duration
	}
	if req.Price != nil {
		repoPlan.Price = *req.Price
	}
	if req.PlanType != nil {
		repoPlan.PlanType = *req.PlanType
	}
	if req.ResetCycle != nil {
		repoPlan.ResetCycle = *req.ResetCycle
	}
	if req.IPLimit != nil {
		repoPlan.IPLimit = *req.IPLimit
	}
	if req.SortOrder != nil {
		repoPlan.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		repoPlan.IsActive = *req.IsActive
	}
	if req.IsRecommended != nil {
		repoPlan.IsRecommended = *req.IsRecommended
	}
	if req.GroupID != nil {
		repoPlan.GroupID = req.GroupID
	}
	if req.PaymentMethods != nil {
		paymentMethodsJSON, _ := json.Marshal(*req.PaymentMethods)
		repoPlan.PaymentMethods = string(paymentMethodsJSON)
	}
	if req.Features != nil {
		featuresJSON, _ := json.Marshal(*req.Features)
		repoPlan.Features = string(featuresJSON)
	}

	if err := s.planRepo.Update(ctx, repoPlan); err != nil {
		s.logger.Error("Failed to update plan", logger.Err(err), logger.F("id", id))
		return nil, err
	}

	return s.toPlan(repoPlan), nil
}

// Delete deletes a plan.
func (s *Service) Delete(ctx context.Context, id int64) error {
	_, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return ErrPlanNotFound
	}

	if err := s.planRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete plan", logger.Err(err), logger.F("id", id))
		return err
	}

	return nil
}

// List lists plans with filter and pagination.
func (s *Service) List(ctx context.Context, filter PlanFilter, page, pageSize int) ([]*Plan, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	repoFilter := repository.PlanFilter{
		IsActive:      filter.IsActive,
		PlanType:      filter.PlanType,
		GroupID:       filter.GroupID,
		MinPrice:      filter.MinPrice,
		MaxPrice:      filter.MaxPrice,
		IsRecommended: filter.IsRecommended,
	}

	offset := (page - 1) * pageSize
	repoPlans, total, err := s.planRepo.List(ctx, repoFilter, pageSize, offset)
	if err != nil {
		s.logger.Error("Failed to list plans", logger.Err(err))
		return nil, 0, err
	}

	plans := make([]*Plan, len(repoPlans))
	for i, rp := range repoPlans {
		plans[i] = s.toPlan(rp)
	}

	return plans, total, nil
}

// ListActive lists all active plans.
func (s *Service) ListActive(ctx context.Context) ([]*Plan, error) {
	repoPlans, err := s.planRepo.ListActive(ctx)
	if err != nil {
		s.logger.Error("Failed to list active plans", logger.Err(err))
		return nil, err
	}

	plans := make([]*Plan, len(repoPlans))
	for i, rp := range repoPlans {
		plans[i] = s.toPlan(rp)
	}

	return plans, nil
}

// SetActive sets the active status of a plan.
func (s *Service) SetActive(ctx context.Context, id int64, active bool) error {
	_, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return ErrPlanNotFound
	}

	if err := s.planRepo.SetActive(ctx, id, active); err != nil {
		s.logger.Error("Failed to set plan active status", logger.Err(err), logger.F("id", id), logger.F("active", active))
		return err
	}

	return nil
}

// CalculateMonthlyPrice calculates the monthly price for a plan.
// Formula: (price / duration) * 30
func (s *Service) CalculateMonthlyPrice(plan *Plan) int64 {
	if plan.Duration <= 0 {
		return 0
	}
	return (plan.Price * 30) / int64(plan.Duration)
}

// CreateGroup creates a new plan group.
func (s *Service) CreateGroup(ctx context.Context, name string, sortOrder int) (*PlanGroup, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidPlan)
	}

	repoGroup := &repository.PlanGroup{
		Name:      name,
		SortOrder: sortOrder,
	}

	if err := s.planRepo.CreateGroup(ctx, repoGroup); err != nil {
		s.logger.Error("Failed to create plan group", logger.Err(err))
		return nil, err
	}

	return &PlanGroup{
		ID:        repoGroup.ID,
		Name:      repoGroup.Name,
		SortOrder: repoGroup.SortOrder,
	}, nil
}

// GetGroupByID retrieves a plan group by ID.
func (s *Service) GetGroupByID(ctx context.Context, id int64) (*PlanGroup, error) {
	repoGroup, err := s.planRepo.GetGroupByID(ctx, id)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	return &PlanGroup{
		ID:        repoGroup.ID,
		Name:      repoGroup.Name,
		SortOrder: repoGroup.SortOrder,
	}, nil
}

// UpdateGroup updates a plan group.
func (s *Service) UpdateGroup(ctx context.Context, id int64, name string, sortOrder int) (*PlanGroup, error) {
	repoGroup, err := s.planRepo.GetGroupByID(ctx, id)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	repoGroup.Name = name
	repoGroup.SortOrder = sortOrder

	if err := s.planRepo.UpdateGroup(ctx, repoGroup); err != nil {
		s.logger.Error("Failed to update plan group", logger.Err(err), logger.F("id", id))
		return nil, err
	}

	return &PlanGroup{
		ID:        repoGroup.ID,
		Name:      repoGroup.Name,
		SortOrder: repoGroup.SortOrder,
	}, nil
}

// DeleteGroup deletes a plan group.
func (s *Service) DeleteGroup(ctx context.Context, id int64) error {
	_, err := s.planRepo.GetGroupByID(ctx, id)
	if err != nil {
		return ErrGroupNotFound
	}

	if err := s.planRepo.DeleteGroup(ctx, id); err != nil {
		s.logger.Error("Failed to delete plan group", logger.Err(err), logger.F("id", id))
		return err
	}

	return nil
}

// ListGroups lists all plan groups.
func (s *Service) ListGroups(ctx context.Context) ([]*PlanGroup, error) {
	repoGroups, err := s.planRepo.ListGroups(ctx)
	if err != nil {
		s.logger.Error("Failed to list plan groups", logger.Err(err))
		return nil, err
	}

	groups := make([]*PlanGroup, len(repoGroups))
	for i, rg := range repoGroups {
		groups[i] = &PlanGroup{
			ID:        rg.ID,
			Name:      rg.Name,
			SortOrder: rg.SortOrder,
		}
	}

	return groups, nil
}

// toPlan converts a repository plan to a service plan.
func (s *Service) toPlan(rp *repository.CommercialPlan) *Plan {
	var paymentMethods []string
	var features []string

	if rp.PaymentMethods != "" {
		_ = json.Unmarshal([]byte(rp.PaymentMethods), &paymentMethods)
	}
	if rp.Features != "" {
		_ = json.Unmarshal([]byte(rp.Features), &features)
	}

	plan := &Plan{
		ID:             rp.ID,
		Name:           rp.Name,
		Description:    rp.Description,
		TrafficLimit:   rp.TrafficLimit,
		Duration:       rp.Duration,
		Price:          rp.Price,
		PlanType:       rp.PlanType,
		ResetCycle:     rp.ResetCycle,
		IPLimit:        rp.IPLimit,
		SortOrder:      rp.SortOrder,
		IsActive:       rp.IsActive,
		IsRecommended:  rp.IsRecommended,
		GroupID:        rp.GroupID,
		PaymentMethods: paymentMethods,
		Features:       features,
	}

	// Calculate monthly price
	plan.MonthlyPrice = s.CalculateMonthlyPrice(plan)

	return plan
}
