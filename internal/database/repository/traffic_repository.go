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


// GetTotalByProxy retrieves total upload and download for a proxy.
func (r *trafficRepository) GetTotalByProxy(ctx context.Context, proxyID int64) (upload, download int64, err error) {
	var result struct {
		Upload   int64
		Download int64
	}
	err = r.db.WithContext(ctx).
		Model(&Traffic{}).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Where("proxy_id = ?", proxyID).
		Scan(&result).Error
	if err != nil {
		return 0, 0, errors.NewDatabaseError("failed to get total traffic by proxy", err)
	}
	return result.Upload, result.Download, nil
}

// GetTotalTraffic retrieves total upload and download for all traffic.
func (r *trafficRepository) GetTotalTraffic(ctx context.Context) (upload, download int64, err error) {
	var result struct {
		Upload   int64
		Download int64
	}
	err = r.db.WithContext(ctx).
		Model(&Traffic{}).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Scan(&result).Error
	if err != nil {
		return 0, 0, errors.NewDatabaseError("failed to get total traffic", err)
	}
	return result.Upload, result.Download, nil
}

// GetTotalTrafficByPeriod retrieves total upload and download within a time period.
func (r *trafficRepository) GetTotalTrafficByPeriod(ctx context.Context, start, end time.Time) (upload, download int64, err error) {
	var result struct {
		Upload   int64
		Download int64
	}
	err = r.db.WithContext(ctx).
		Model(&Traffic{}).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Where("recorded_at BETWEEN ? AND ?", start, end).
		Scan(&result).Error
	if err != nil {
		return 0, 0, errors.NewDatabaseError("failed to get total traffic by period", err)
	}
	return result.Upload, result.Download, nil
}

// GetTrafficByProtocol retrieves traffic statistics grouped by protocol.
func (r *trafficRepository) GetTrafficByProtocol(ctx context.Context, start, end time.Time) ([]*ProtocolTrafficStats, error) {
	var results []*ProtocolTrafficStats
	err := r.db.WithContext(ctx).
		Table("traffic t").
		Select("p.protocol, COUNT(DISTINCT p.id) as count, COALESCE(SUM(t.upload), 0) as upload, COALESCE(SUM(t.download), 0) as download").
		Joins("JOIN proxies p ON t.proxy_id = p.id").
		Where("t.recorded_at BETWEEN ? AND ?", start, end).
		Group("p.protocol").
		Scan(&results).Error
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get traffic by protocol", err)
	}
	return results, nil
}

// GetTrafficByUser retrieves traffic statistics grouped by user.
func (r *trafficRepository) GetTrafficByUser(ctx context.Context, start, end time.Time, limit int) ([]*UserTrafficStats, error) {
	var results []*UserTrafficStats
	query := r.db.WithContext(ctx).
		Table("traffic t").
		Select("t.user_id, u.username, COALESCE(SUM(t.upload), 0) as upload, COALESCE(SUM(t.download), 0) as download, COUNT(DISTINCT t.proxy_id) as proxy_count").
		Joins("JOIN users u ON t.user_id = u.id").
		Where("t.recorded_at BETWEEN ? AND ?", start, end).
		Group("t.user_id, u.username").
		Order("(upload + download) DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get traffic by user", err)
	}
	return results, nil
}

// GetTrafficTimeline retrieves traffic data points for timeline charts.
func (r *trafficRepository) GetTrafficTimeline(ctx context.Context, start, end time.Time, interval string) ([]*TrafficTimelinePoint, error) {
	var results []*TrafficTimelinePoint

	// SQLite date formatting based on interval
	var dateFormat string
	switch interval {
	case "hour":
		dateFormat = "%Y-%m-%d %H:00:00"
	case "day":
		dateFormat = "%Y-%m-%d"
	case "month":
		dateFormat = "%Y-%m"
	default:
		dateFormat = "%Y-%m-%d %H:00:00"
	}

	selectClause := "strftime('" + dateFormat + "', recorded_at) as time, COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download"
	groupClause := "strftime('" + dateFormat + "', recorded_at)"

	err := r.db.WithContext(ctx).
		Table("traffic").
		Select(selectClause).
		Where("recorded_at BETWEEN ? AND ?", start, end).
		Group(groupClause).
		Order("time ASC").
		Scan(&results).Error
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get traffic timeline", err)
	}
	return results, nil
}
