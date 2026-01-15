// Package stats provides statistics services for the user portal.
package stats

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"v/internal/database/repository"
)

// Service provides statistics operations for the user portal.
type Service struct {
	trafficRepo repository.TrafficRepository
	userRepo    repository.UserRepository
}

// NewService creates a new stats service.
func NewService(trafficRepo repository.TrafficRepository, userRepo repository.UserRepository) *Service {
	return &Service{
		trafficRepo: trafficRepo,
		userRepo:    userRepo,
	}
}

// TrafficSummary represents a summary of traffic usage.
type TrafficSummary struct {
	Upload      int64   `json:"upload"`
	Download    int64   `json:"download"`
	Total       int64   `json:"total"`
	UploadStr   string  `json:"upload_str"`
	DownloadStr string  `json:"download_str"`
	TotalStr    string  `json:"total_str"`
}

// DailyTraffic represents traffic for a single day.
type DailyTraffic struct {
	Date     string `json:"date"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
	Total    int64  `json:"total"`
}

// TrafficStats represents traffic statistics for a period.
type TrafficStats struct {
	Summary   *TrafficSummary  `json:"summary"`
	Daily     []*DailyTraffic  `json:"daily"`
	StartDate string           `json:"start_date"`
	EndDate   string           `json:"end_date"`
	Period    string           `json:"period"` // day, week, month
}

// GetTrafficSummary retrieves total traffic for a user.
func (s *Service) GetTrafficSummary(ctx context.Context, userID int64) (*TrafficSummary, error) {
	upload, download, err := s.trafficRepo.GetTotalByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	total := upload + download
	return &TrafficSummary{
		Upload:      upload,
		Download:    download,
		Total:       total,
		UploadStr:   FormatBytes(upload),
		DownloadStr: FormatBytes(download),
		TotalStr:    FormatBytes(total),
	}, nil
}

// GetDailyTraffic retrieves daily traffic for a user within a date range.
func (s *Service) GetDailyTraffic(ctx context.Context, userID int64, days int) ([]*DailyTraffic, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}

	end := time.Now()
	start := end.AddDate(0, 0, -days)

	// Get traffic timeline
	timeline, err := s.trafficRepo.GetTrafficTimeline(ctx, start, end, "day")
	if err != nil {
		return nil, err
	}

	// Convert to DailyTraffic
	daily := make([]*DailyTraffic, len(timeline))
	for i, point := range timeline {
		daily[i] = &DailyTraffic{
			Date:     point.Time.Format("2006-01-02"),
			Upload:   point.Upload,
			Download: point.Download,
			Total:    point.Upload + point.Download,
		}
	}

	return daily, nil
}

// GetTrafficStats retrieves traffic statistics for a period.
func (s *Service) GetTrafficStats(ctx context.Context, userID int64, period string) (*TrafficStats, error) {
	var days int
	switch period {
	case "week":
		days = 7
	case "month":
		days = 30
	case "year":
		days = 365
	default:
		period = "month"
		days = 30
	}

	end := time.Now()
	start := end.AddDate(0, 0, -days)

	// Get summary
	upload, download, err := s.trafficRepo.GetTotalTrafficByPeriod(ctx, start, end)
	if err != nil {
		return nil, err
	}

	total := upload + download
	summary := &TrafficSummary{
		Upload:      upload,
		Download:    download,
		Total:       total,
		UploadStr:   FormatBytes(upload),
		DownloadStr: FormatBytes(download),
		TotalStr:    FormatBytes(total),
	}

	// Get daily data
	daily, err := s.GetDailyTraffic(ctx, userID, days)
	if err != nil {
		return nil, err
	}

	return &TrafficStats{
		Summary:   summary,
		Daily:     daily,
		StartDate: start.Format("2006-01-02"),
		EndDate:   end.Format("2006-01-02"),
		Period:    period,
	}, nil
}

// ExportTrafficCSV exports traffic data as CSV.
func (s *Service) ExportTrafficCSV(ctx context.Context, userID int64, days int) ([]byte, error) {
	daily, err := s.GetDailyTraffic(ctx, userID, days)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	if err := writer.Write([]string{"Date", "Upload (bytes)", "Download (bytes)", "Total (bytes)", "Upload", "Download", "Total"}); err != nil {
		return nil, err
	}

	// Write data
	for _, d := range daily {
		record := []string{
			d.Date,
			fmt.Sprintf("%d", d.Upload),
			fmt.Sprintf("%d", d.Download),
			fmt.Sprintf("%d", d.Total),
			FormatBytes(d.Upload),
			FormatBytes(d.Download),
			FormatBytes(d.Total),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// AggregateDaily aggregates a list of daily traffic records.
func AggregateDaily(daily []*DailyTraffic) *TrafficSummary {
	var upload, download int64
	for _, d := range daily {
		upload += d.Upload
		download += d.Download
	}
	total := upload + download
	return &TrafficSummary{
		Upload:      upload,
		Download:    download,
		Total:       total,
		UploadStr:   FormatBytes(upload),
		DownloadStr: FormatBytes(download),
		TotalStr:    FormatBytes(total),
	}
}

// FormatBytes formats bytes into human-readable string.
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ValidatePeriod validates and normalizes a period string.
func ValidatePeriod(period string) string {
	switch period {
	case "day", "week", "month", "year":
		return period
	default:
		return "month"
	}
}

// GetPeriodDays returns the number of days for a period.
func GetPeriodDays(period string) int {
	switch period {
	case "day":
		return 1
	case "week":
		return 7
	case "month":
		return 30
	case "year":
		return 365
	default:
		return 30
	}
}
