// Package repository provides data access interfaces and implementations.
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Invoice represents an invoice in the database.
type Invoice struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	InvoiceNo string    `gorm:"uniqueIndex;size:64;not null"`
	OrderID   int64     `gorm:"index;not null"`
	UserID    int64     `gorm:"index;not null"`
	Amount    int64     `gorm:"not null"`
	Content   string    `gorm:"type:text;not null"`
	PDFPath   string    `gorm:"size:256"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	Order *Order `gorm:"foreignKey:OrderID"`
	User  *User  `gorm:"foreignKey:UserID"`
}

// TableName returns the table name for Invoice.
func (Invoice) TableName() string {
	return "invoices"
}

// InvoiceFilter defines filter options for listing invoices.
type InvoiceFilter struct {
	UserID    *int64
	StartDate *time.Time
	EndDate   *time.Time
}

// InvoiceRepository defines the interface for invoice data access.
type InvoiceRepository interface {
	Create(ctx context.Context, invoice *Invoice) error
	GetByID(ctx context.Context, id int64) (*Invoice, error)
	GetByInvoiceNo(ctx context.Context, invoiceNo string) (*Invoice, error)
	GetByOrderID(ctx context.Context, orderID int64) (*Invoice, error)
	Update(ctx context.Context, invoice *Invoice) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter InvoiceFilter, limit, offset int) ([]*Invoice, int64, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*Invoice, int64, error)
	Count(ctx context.Context) (int64, error)
	CountByUser(ctx context.Context, userID int64) (int64, error)
}

// invoiceRepository implements InvoiceRepository.
type invoiceRepository struct {
	db *gorm.DB
}

// NewInvoiceRepository creates a new invoice repository.
func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

// Create creates a new invoice.
func (r *invoiceRepository) Create(ctx context.Context, invoice *Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

// GetByID retrieves an invoice by ID.
func (r *invoiceRepository) GetByID(ctx context.Context, id int64) (*Invoice, error) {
	var invoice Invoice
	err := r.db.WithContext(ctx).Preload("Order").Preload("User").First(&invoice, id).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// GetByInvoiceNo retrieves an invoice by invoice number.
func (r *invoiceRepository) GetByInvoiceNo(ctx context.Context, invoiceNo string) (*Invoice, error) {
	var invoice Invoice
	err := r.db.WithContext(ctx).Preload("Order").Preload("User").
		Where("invoice_no = ?", invoiceNo).First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// GetByOrderID retrieves an invoice by order ID.
func (r *invoiceRepository) GetByOrderID(ctx context.Context, orderID int64) (*Invoice, error) {
	var invoice Invoice
	err := r.db.WithContext(ctx).Preload("Order").Preload("User").
		Where("order_id = ?", orderID).First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// Update updates an invoice.
func (r *invoiceRepository) Update(ctx context.Context, invoice *Invoice) error {
	return r.db.WithContext(ctx).Save(invoice).Error
}

// Delete deletes an invoice by ID.
func (r *invoiceRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Invoice{}, id).Error
}

// List lists invoices with filter and pagination.
func (r *invoiceRepository) List(ctx context.Context, filter InvoiceFilter, limit, offset int) ([]*Invoice, int64, error) {
	var invoices []*Invoice
	var total int64

	query := r.db.WithContext(ctx).Model(&Invoice{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Order").Preload("User").
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&invoices).Error
	return invoices, total, err
}

// ListByUser lists invoices for a specific user.
func (r *invoiceRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*Invoice, int64, error) {
	var invoices []*Invoice
	var total int64

	query := r.db.WithContext(ctx).Model(&Invoice{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Order").
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&invoices).Error
	return invoices, total, err
}

// Count returns the total number of invoices.
func (r *invoiceRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Invoice{}).Count(&count).Error
	return count, err
}

// CountByUser returns the number of invoices for a user.
func (r *invoiceRepository) CountByUser(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Invoice{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
