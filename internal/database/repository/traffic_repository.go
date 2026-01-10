// Package repository provides data access implementations.
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"v/pkg/errors"
)

// trafficRepository implements TrafficRepository.
type trafficRepository struct {
	db *gorm.DB
}

// NewTrafficRepository creates a new traffic repository.
func NewTrafficRepository(db *gorm.DB) TrafficRepository {
	return &trafficRepository{db: db}
}

// Create creates a new traffic record.
func (r *trafficRepository) Create(ctx context.Context, traffic *Traffic) error {
	result := r.db.WithContext(ctx).Create(traffic)
	if result.Error != nil {
		return errors.NewDatabaseError("failed to create traffic record", result.Error)
	}
	return nil
}

// GetByUserID retrieves traffic records by user ID.
func (r *trafficRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*Traffic, error) {
	var records []*Traffic
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("recorded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records)
	if result.Error != nil {
		return nil, errors.NewDatabaseError("failed to get traffic by user", result.Error)
	}
	return records, nil
}

// GetByProxyID retrieves traffic records by proxy ID.
func (r *trafficRepository) GetByProxyID(ctx context.Context, proxyID int64, limit, offset int) ([]*Traffic, error) {
	var records []*Traffic
	result := r.db.WithContext(ctx).
		Where("proxy_id = ?", proxyID).
		Order("recorded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records)
	if result.Error != nil {
		return nil, errors.NewDatabaseError("failed to get traffic by proxy", result.Error)
	}
	return records, nil
}

// GetByDateRange retrieves traffic records within a date range.
func (r *trafficRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]*Traffic, error) {
	var records []*Traffic
	result := r.db.WithContext(ctx).
		Where("recorded_at BETWEEN ? AND ?", start, end).
		Order("recorded_at DESC").
		Find(&records)
	if result.Error != nil {
		return nil, errors.NewDatabaseError("failed to get traffic by date range", result.Error)
	}
	return records, nil
}

// GetTotalByUser retrieves total upload and download for a user.
func (r *trafficRepository) GetTotalByUser(ctx context.Context, userID int64) (upload, download int64, err error) {
	var result struct {
		Upload   int64
		Download int64
	}
	err = r.db.WithContext(ctx).
		Model(&Traffic{}).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Where("user_id = ?", userID).
		Scan(&result).Error
	if err != nil {
		return 0, 0, errors.NewDatabaseError("failed to get total traffic", err)
	}
	return result.Upload, result.Download, nil
}
