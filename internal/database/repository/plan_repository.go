// Package repository provides data access interfaces and implementations.
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// CommercialPlan represents a commercial plan in the database.
type CommercialPlan struct {
	ID             int64     `gorm:"primaryKey;autoIncrement"`
	Name           string    `gorm:"size:128;not null"`
	Description    string    `gorm:"type:text"`
	TrafficLimit   int64     `gorm:"default:0"`
	Duration       int       `gorm:"not null"`
	Price          int64     `gorm:"not null"`
	PlanType       string    `gorm:"size:32;default:monthly"`
	ResetCycle     string    `gorm:"size:32;default:monthly"`
	IPLimit        int       `gorm:"default:0"`
	SortOrder      int       `gorm:"default:0"`
	IsActive       bool      `gorm:"default:true"`
	IsRecommended  bool      `gorm:"default:false"`
	GroupID        *int64    `gorm:"index"`
	PaymentMethods string    `gorm:"type:text"`
	Features       string    `gorm:"type:text"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

// TableName returns the table name for CommercialPlan.
func (CommercialPlan) TableName() string {
	return "commercial_plans"
}

// PlanGroup represents a plan group in the database.
type PlanGroup struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:64;not null"`
	SortOrder int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName returns the table name for PlanGroup.
func (PlanGroup) TableName() string {
	return "plan_groups"
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

// PlanRepository defines the interface for plan data access.
type PlanRepository interface {
	Create(ctx context.Context, plan *CommercialPlan) error
	GetByID(ctx context.Context, id int64) (*CommercialPlan, error)
	Update(ctx context.Context, plan *CommercialPlan) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter PlanFilter, limit, offset int) ([]*CommercialPlan, int64, error)
	ListActive(ctx context.Context) ([]*CommercialPlan, error)
	SetActive(ctx context.Context, id int64, active bool) error
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
	// Group operations
	CreateGroup(ctx context.Context, group *PlanGroup) error
	GetGroupByID(ctx context.Context, id int64) (*PlanGroup, error)
	UpdateGroup(ctx context.Context, group *PlanGroup) error
	DeleteGroup(ctx context.Context, id int64) error
	ListGroups(ctx context.Context) ([]*PlanGroup, error)
}

// planRepository implements PlanRepository.
type planRepository struct {
	db *gorm.DB
}

// NewPlanRepository creates a new plan repository.
func NewPlanRepository(db *gorm.DB) PlanRepository {
	return &planRepository{db: db}
}

// Create creates a new plan.
func (r *planRepository) Create(ctx context.Context, plan *CommercialPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

// GetByID retrieves a plan by ID.
func (r *planRepository) GetByID(ctx context.Context, id int64) (*CommercialPlan, error) {
	var plan CommercialPlan
	err := r.db.WithContext(ctx).First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// Update updates a plan.
func (r *planRepository) Update(ctx context.Context, plan *CommercialPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

// Delete deletes a plan by ID.
func (r *planRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&CommercialPlan{}, id).Error
}

// List lists plans with filter and pagination.
func (r *planRepository) List(ctx context.Context, filter PlanFilter, limit, offset int) ([]*CommercialPlan, int64, error) {
	var plans []*CommercialPlan
	var total int64

	query := r.db.WithContext(ctx).Model(&CommercialPlan{})

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.PlanType != "" {
		query = query.Where("plan_type = ?", filter.PlanType)
	}
	if filter.GroupID != nil {
		query = query.Where("group_id = ?", *filter.GroupID)
	}
	if filter.MinPrice != nil {
		query = query.Where("price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		query = query.Where("price <= ?", *filter.MaxPrice)
	}
	if filter.IsRecommended != nil {
		query = query.Where("is_recommended = ?", *filter.IsRecommended)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("sort_order ASC, id ASC").Limit(limit).Offset(offset).Find(&plans).Error
	return plans, total, err
}

// ListActive lists all active plans.
func (r *planRepository) ListActive(ctx context.Context) ([]*CommercialPlan, error) {
	var plans []*CommercialPlan
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&plans).Error
	return plans, err
}

// SetActive sets the active status of a plan.
func (r *planRepository) SetActive(ctx context.Context, id int64, active bool) error {
	return r.db.WithContext(ctx).
		Model(&CommercialPlan{}).
		Where("id = ?", id).
		Update("is_active", active).Error
}

// Count returns the total number of plans.
func (r *planRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CommercialPlan{}).Count(&count).Error
	return count, err
}

// CountActive returns the number of active plans.
func (r *planRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CommercialPlan{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// CreateGroup creates a new plan group.
func (r *planRepository) CreateGroup(ctx context.Context, group *PlanGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetGroupByID retrieves a plan group by ID.
func (r *planRepository) GetGroupByID(ctx context.Context, id int64) (*PlanGroup, error) {
	var group PlanGroup
	err := r.db.WithContext(ctx).First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateGroup updates a plan group.
func (r *planRepository) UpdateGroup(ctx context.Context, group *PlanGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteGroup deletes a plan group by ID.
func (r *planRepository) DeleteGroup(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&PlanGroup{}, id).Error
}

// ListGroups lists all plan groups.
func (r *planRepository) ListGroups(ctx context.Context) ([]*PlanGroup, error) {
	var groups []*PlanGroup
	err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&groups).Error
	return groups, err
}
