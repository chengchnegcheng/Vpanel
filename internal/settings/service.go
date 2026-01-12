// Package settings provides system settings management.
package settings

import (
	"context"
	"encoding/json"
	"sync"

	"v/internal/database/repository"
)

// SystemSettings represents all system settings.
type SystemSettings struct {
	SiteName            string `json:"site_name"`
	SiteDescription     string `json:"site_description"`
	AllowRegistration   bool   `json:"allow_registration"`
	DefaultTrafficLimit int64  `json:"default_traffic_limit"`
	DefaultExpiryDays   int    `json:"default_expiry_days"`

	// SMTP settings
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"-"` // Hidden in JSON responses

	// Telegram settings
	TelegramBotToken string `json:"-"` // Hidden in JSON responses
	TelegramChatID   string `json:"telegram_chat_id"`

	// Rate limiting
	RateLimitEnabled  bool `json:"rate_limit_enabled"`
	RateLimitRequests int  `json:"rate_limit_requests"`
	RateLimitWindow   int  `json:"rate_limit_window"`

	// Xray settings
	XrayConfigTemplate string `json:"xray_config_template"`
}

// DefaultSettings returns default system settings.
func DefaultSettings() *SystemSettings {
	return &SystemSettings{
		SiteName:            "V Panel",
		SiteDescription:     "Proxy Server Management Panel",
		AllowRegistration:   false,
		DefaultTrafficLimit: 0, // Unlimited
		DefaultExpiryDays:   30,
		SMTPPort:            587,
		RateLimitEnabled:    true,
		RateLimitRequests:   100,
		RateLimitWindow:     60, // seconds
	}
}

// Service provides settings management functionality.
type Service struct {
	repo    repository.SettingsRepository
	cache   *SystemSettings
	cacheMu sync.RWMutex
}

// NewService creates a new settings service.
func NewService(repo repository.SettingsRepository) *Service {
	return &Service{
		repo:  repo,
		cache: nil,
	}
}

// Get retrieves a single setting value.
func (s *Service) Get(ctx context.Context, key string) (string, error) {
	return s.repo.Get(ctx, key)
}

// GetAll retrieves all settings as a map.
func (s *Service) GetAll(ctx context.Context) (map[string]string, error) {
	return s.repo.GetAll(ctx)
}

// GetTyped retrieves a setting and unmarshals it into the target.
func (s *Service) GetTyped(ctx context.Context, key string, target interface{}) error {
	value, err := s.repo.Get(ctx, key)
	if err != nil {
		return err
	}
	if value == "" {
		return nil
	}
	return json.Unmarshal([]byte(value), target)
}

// Set updates a single setting.
func (s *Service) Set(ctx context.Context, key, value string) error {
	err := s.repo.Set(ctx, key, value)
	if err != nil {
		return err
	}
	// Invalidate cache
	s.cacheMu.Lock()
	s.cache = nil
	s.cacheMu.Unlock()
	return nil
}

// SetMultiple updates multiple settings.
func (s *Service) SetMultiple(ctx context.Context, settings map[string]string) error {
	err := s.repo.SetMultiple(ctx, settings)
	if err != nil {
		return err
	}
	// Invalidate cache
	s.cacheMu.Lock()
	s.cache = nil
	s.cacheMu.Unlock()
	return nil
}

// GetSystemSettings retrieves all system settings as a structured object.
func (s *Service) GetSystemSettings(ctx context.Context) (*SystemSettings, error) {
	// Check cache first
	s.cacheMu.RLock()
	if s.cache != nil {
		cached := *s.cache
		s.cacheMu.RUnlock()
		return &cached, nil
	}
	s.cacheMu.RUnlock()

	// Load from database
	allSettings, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	settings := DefaultSettings()

	// Map database values to struct
	if v, ok := allSettings["site_name"]; ok && v != "" {
		settings.SiteName = v
	}
	if v, ok := allSettings["site_description"]; ok && v != "" {
		settings.SiteDescription = v
	}
	if v, ok := allSettings["allow_registration"]; ok {
		settings.AllowRegistration = v == "true"
	}
	if v, ok := allSettings["default_traffic_limit"]; ok && v != "" {
		var limit int64
		if json.Unmarshal([]byte(v), &limit) == nil {
			settings.DefaultTrafficLimit = limit
		}
	}
	if v, ok := allSettings["default_expiry_days"]; ok && v != "" {
		var days int
		if json.Unmarshal([]byte(v), &days) == nil {
			settings.DefaultExpiryDays = days
		}
	}
	if v, ok := allSettings["smtp_host"]; ok {
		settings.SMTPHost = v
	}
	if v, ok := allSettings["smtp_port"]; ok && v != "" {
		var port int
		if json.Unmarshal([]byte(v), &port) == nil {
			settings.SMTPPort = port
		}
	}
	if v, ok := allSettings["smtp_user"]; ok {
		settings.SMTPUser = v
	}
	if v, ok := allSettings["smtp_password"]; ok {
		settings.SMTPPassword = v
	}
	if v, ok := allSettings["telegram_bot_token"]; ok {
		settings.TelegramBotToken = v
	}
	if v, ok := allSettings["telegram_chat_id"]; ok {
		settings.TelegramChatID = v
	}
	if v, ok := allSettings["rate_limit_enabled"]; ok {
		settings.RateLimitEnabled = v == "true"
	}
	if v, ok := allSettings["rate_limit_requests"]; ok && v != "" {
		var requests int
		if json.Unmarshal([]byte(v), &requests) == nil {
			settings.RateLimitRequests = requests
		}
	}
	if v, ok := allSettings["rate_limit_window"]; ok && v != "" {
		var window int
		if json.Unmarshal([]byte(v), &window) == nil {
			settings.RateLimitWindow = window
		}
	}
	if v, ok := allSettings["xray_config_template"]; ok {
		settings.XrayConfigTemplate = v
	}

	// Update cache
	s.cacheMu.Lock()
	s.cache = settings
	s.cacheMu.Unlock()

	return settings, nil
}

// UpdateSystemSettings updates system settings from a structured object.
func (s *Service) UpdateSystemSettings(ctx context.Context, settings *SystemSettings) error {
	updates := make(map[string]string)

	updates["site_name"] = settings.SiteName
	updates["site_description"] = settings.SiteDescription
	updates["allow_registration"] = boolToString(settings.AllowRegistration)

	if data, err := json.Marshal(settings.DefaultTrafficLimit); err == nil {
		updates["default_traffic_limit"] = string(data)
	}
	if data, err := json.Marshal(settings.DefaultExpiryDays); err == nil {
		updates["default_expiry_days"] = string(data)
	}

	updates["smtp_host"] = settings.SMTPHost
	if data, err := json.Marshal(settings.SMTPPort); err == nil {
		updates["smtp_port"] = string(data)
	}
	updates["smtp_user"] = settings.SMTPUser
	if settings.SMTPPassword != "" {
		updates["smtp_password"] = settings.SMTPPassword
	}

	if settings.TelegramBotToken != "" {
		updates["telegram_bot_token"] = settings.TelegramBotToken
	}
	updates["telegram_chat_id"] = settings.TelegramChatID

	updates["rate_limit_enabled"] = boolToString(settings.RateLimitEnabled)
	if data, err := json.Marshal(settings.RateLimitRequests); err == nil {
		updates["rate_limit_requests"] = string(data)
	}
	if data, err := json.Marshal(settings.RateLimitWindow); err == nil {
		updates["rate_limit_window"] = string(data)
	}

	updates["xray_config_template"] = settings.XrayConfigTemplate

	return s.SetMultiple(ctx, updates)
}

// Backup creates a backup of all settings.
func (s *Service) Backup(ctx context.Context) ([]byte, error) {
	return s.repo.Backup(ctx)
}

// Restore restores settings from a backup.
func (s *Service) Restore(ctx context.Context, data []byte) error {
	err := s.repo.Restore(ctx, data)
	if err != nil {
		return err
	}
	// Invalidate cache
	s.cacheMu.Lock()
	s.cache = nil
	s.cacheMu.Unlock()
	return nil
}

// InvalidateCache clears the settings cache.
func (s *Service) InvalidateCache() {
	s.cacheMu.Lock()
	s.cache = nil
	s.cacheMu.Unlock()
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
