package ip

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Tracker provides IP tracking functionality.
type Tracker struct {
	db *gorm.DB
}

// NewTracker creates a new Tracker instance.
func NewTracker(db *gorm.DB) *Tracker {
	return &Tracker{db: db}
}

// AddActiveIP adds or updates an active IP for a user.
func (t *Tracker) AddActiveIP(ctx context.Context, userID uint, ip, userAgent, deviceType, country, city string) error {
	activeIP := ActiveIP{
		UserID:     userID,
		IP:         ip,
		UserAgent:  userAgent,
		DeviceType: deviceType,
		Country:    country,
		City:       city,
		LastActive: time.Now(),
	}

	// Upsert: update if exists, insert if not
	return t.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "ip"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_agent", "device_type", "country", "city", "last_active"}),
	}).Create(&activeIP).Error
}

// RemoveActiveIP removes an active IP for a user.
func (t *Tracker) RemoveActiveIP(ctx context.Context, userID uint, ip string) error {
	return t.db.WithContext(ctx).
		Where("user_id = ? AND ip = ?", userID, ip).
		Delete(&ActiveIP{}).Error
}

// GetActiveIPCount returns the count of active IPs for a user.
func (t *Tracker) GetActiveIPCount(ctx context.Context, userID uint) (int, error) {
	var count int64
	err := t.db.WithContext(ctx).
		Model(&ActiveIP{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return int(count), err
}

// GetActiveIPs returns all active IPs for a user.
func (t *Tracker) GetActiveIPs(ctx context.Context, userID uint) ([]ActiveIP, error) {
	var ips []ActiveIP
	err := t.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("last_active DESC").
		Find(&ips).Error
	return ips, err
}

// GetOnlineIPs returns online IP information for a user.
func (t *Tracker) GetOnlineIPs(ctx context.Context, userID uint) ([]OnlineIP, error) {
	activeIPs, err := t.GetActiveIPs(ctx, userID)
	if err != nil {
		return nil, err
	}

	onlineIPs := make([]OnlineIP, len(activeIPs))
	for i, ip := range activeIPs {
		onlineIPs[i] = OnlineIP{
			IP:         ip.IP,
			UserAgent:  ip.UserAgent,
			DeviceType: ip.DeviceType,
			Country:    ip.Country,
			City:       ip.City,
			LastActive: ip.LastActive,
			CreatedAt:  ip.CreatedAt,
		}
	}
	return onlineIPs, nil
}

// UpdateLastActive updates the last active timestamp for an IP.
func (t *Tracker) UpdateLastActive(ctx context.Context, userID uint, ip string) error {
	return t.db.WithContext(ctx).
		Model(&ActiveIP{}).
		Where("user_id = ? AND ip = ?", userID, ip).
		Update("last_active", time.Now()).Error
}

// IsIPActive checks if an IP is currently active for a user.
func (t *Tracker) IsIPActive(ctx context.Context, userID uint, ip string) (bool, error) {
	var count int64
	err := t.db.WithContext(ctx).
		Model(&ActiveIP{}).
		Where("user_id = ? AND ip = ?", userID, ip).
		Count(&count).Error
	return count > 0, err
}


// CleanupInactiveIPs removes IPs that have been inactive for longer than the timeout.
func (t *Tracker) CleanupInactiveIPs(ctx context.Context, timeout time.Duration) (int, error) {
	cutoff := time.Now().Add(-timeout)
	result := t.db.WithContext(ctx).
		Where("last_active < ?", cutoff).
		Delete(&ActiveIP{})
	return int(result.RowsAffected), result.Error
}

// CleanupInactiveIPsForUser removes inactive IPs for a specific user.
func (t *Tracker) CleanupInactiveIPsForUser(ctx context.Context, userID uint, timeout time.Duration) (int, error) {
	cutoff := time.Now().Add(-timeout)
	result := t.db.WithContext(ctx).
		Where("user_id = ? AND last_active < ?", userID, cutoff).
		Delete(&ActiveIP{})
	return int(result.RowsAffected), result.Error
}

// RemoveAllActiveIPs removes all active IPs for a user.
func (t *Tracker) RemoveAllActiveIPs(ctx context.Context, userID uint) error {
	return t.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&ActiveIP{}).Error
}

// RecordIPHistory records an IP access in the history.
func (t *Tracker) RecordIPHistory(ctx context.Context, record *IPHistory) error {
	return t.db.WithContext(ctx).Create(record).Error
}

// GetIPHistory returns IP history for a user with optional filters.
func (t *Tracker) GetIPHistory(ctx context.Context, userID uint, filter *IPHistoryFilter) ([]IPHistory, error) {
	var records []IPHistory
	query := t.db.WithContext(ctx).Where("user_id = ?", userID)

	if filter != nil {
		if filter.StartTime != nil {
			query = query.Where("created_at >= ?", *filter.StartTime)
		}
		if filter.EndTime != nil {
			query = query.Where("created_at <= ?", *filter.EndTime)
		}
		if filter.AccessType != nil {
			query = query.Where("access_type = ?", *filter.AccessType)
		}
		if filter.IP != "" {
			query = query.Where("ip = ?", filter.IP)
		}
		if filter.Country != "" {
			query = query.Where("country = ?", filter.Country)
		}
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	err := query.Order("created_at DESC").Find(&records).Error
	return records, err
}

// GetUniqueIPCount returns the count of unique IPs for a user within a time range.
func (t *Tracker) GetUniqueIPCount(ctx context.Context, userID uint, startTime, endTime time.Time) (int, error) {
	var count int64
	err := t.db.WithContext(ctx).
		Model(&IPHistory{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startTime, endTime).
		Distinct("ip").
		Count(&count).Error
	return int(count), err
}

// GetIPsByCountry returns IP counts grouped by country for a user.
func (t *Tracker) GetIPsByCountry(ctx context.Context, userID uint) (map[string]int, error) {
	type result struct {
		Country string
		Count   int
	}
	var results []result

	err := t.db.WithContext(ctx).
		Model(&IPHistory{}).
		Select("country, COUNT(DISTINCT ip) as count").
		Where("user_id = ?", userID).
		Group("country").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	countryMap := make(map[string]int)
	for _, r := range results {
		countryMap[r.Country] = r.Count
	}
	return countryMap, nil
}

// CleanupOldHistory removes IP history records older than the retention period.
func (t *Tracker) CleanupOldHistory(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := t.db.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&IPHistory{})
	return result.RowsAffected, result.Error
}

// MarkSuspicious marks an IP history record as suspicious.
func (t *Tracker) MarkSuspicious(ctx context.Context, id uint) error {
	return t.db.WithContext(ctx).
		Model(&IPHistory{}).
		Where("id = ?", id).
		Update("is_suspicious", true).Error
}

// GetRecentCountries returns the countries accessed by a user in the last N minutes.
func (t *Tracker) GetRecentCountries(ctx context.Context, userID uint, minutes int) ([]string, error) {
	var countries []string
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)

	err := t.db.WithContext(ctx).
		Model(&IPHistory{}).
		Where("user_id = ? AND created_at >= ?", userID, cutoff).
		Distinct("country").
		Pluck("country", &countries).Error

	return countries, err
}


// GetDB returns the database connection.
func (t *Tracker) GetDB() *gorm.DB {
	return t.db
}
