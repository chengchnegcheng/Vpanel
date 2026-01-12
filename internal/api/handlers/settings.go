// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"v/internal/api/middleware"
	"v/internal/logger"
	"v/internal/settings"
	"v/pkg/errors"
)

// SettingsHandler handles settings-related requests.
type SettingsHandler struct {
	logger          logger.Logger
	settingsService *settings.Service
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(log logger.Logger, settingsService *settings.Service) *SettingsHandler {
	return &SettingsHandler{
		logger:          log,
		settingsService: settingsService,
	}
}

// GetSettings returns all system settings.
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	ctx := c.Request.Context()

	systemSettings, err := h.settingsService.GetSystemSettings(ctx)
	if err != nil {
		h.logger.Error("Failed to get settings", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("get settings", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    systemSettings,
	})
}

// UpdateSettingsRequest represents an update settings request.
type UpdateSettingsRequest struct {
	SiteName            *string `json:"site_name"`
	SiteDescription     *string `json:"site_description"`
	AllowRegistration   *bool   `json:"allow_registration"`
	DefaultTrafficLimit *int64  `json:"default_traffic_limit"`
	DefaultExpiryDays   *int    `json:"default_expiry_days"`

	// SMTP settings
	SMTPHost     *string `json:"smtp_host"`
	SMTPPort     *int    `json:"smtp_port"`
	SMTPUser     *string `json:"smtp_user"`
	SMTPPassword *string `json:"smtp_password"`

	// Telegram settings
	TelegramBotToken *string `json:"telegram_bot_token"`
	TelegramChatID   *string `json:"telegram_chat_id"`

	// Rate limiting
	RateLimitEnabled  *bool `json:"rate_limit_enabled"`
	RateLimitRequests *int  `json:"rate_limit_requests"`
	RateLimitWindow   *int  `json:"rate_limit_window"`

	// Xray settings
	XrayConfigTemplate *string `json:"xray_config_template"`
}

// UpdateSettings updates system settings.
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	ctx := c.Request.Context()

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid request", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	// Get current settings
	currentSettings, err := h.settingsService.GetSystemSettings(ctx)
	if err != nil {
		h.logger.Error("Failed to get current settings", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("get settings", err))
		return
	}

	// Apply updates (only update fields that are provided)
	if req.SiteName != nil {
		currentSettings.SiteName = *req.SiteName
	}
	if req.SiteDescription != nil {
		currentSettings.SiteDescription = *req.SiteDescription
	}
	if req.AllowRegistration != nil {
		currentSettings.AllowRegistration = *req.AllowRegistration
	}
	if req.DefaultTrafficLimit != nil {
		currentSettings.DefaultTrafficLimit = *req.DefaultTrafficLimit
	}
	if req.DefaultExpiryDays != nil {
		currentSettings.DefaultExpiryDays = *req.DefaultExpiryDays
	}
	if req.SMTPHost != nil {
		currentSettings.SMTPHost = *req.SMTPHost
	}
	if req.SMTPPort != nil {
		currentSettings.SMTPPort = *req.SMTPPort
	}
	if req.SMTPUser != nil {
		currentSettings.SMTPUser = *req.SMTPUser
	}
	if req.SMTPPassword != nil {
		currentSettings.SMTPPassword = *req.SMTPPassword
	}
	if req.TelegramBotToken != nil {
		currentSettings.TelegramBotToken = *req.TelegramBotToken
	}
	if req.TelegramChatID != nil {
		currentSettings.TelegramChatID = *req.TelegramChatID
	}
	if req.RateLimitEnabled != nil {
		currentSettings.RateLimitEnabled = *req.RateLimitEnabled
	}
	if req.RateLimitRequests != nil {
		currentSettings.RateLimitRequests = *req.RateLimitRequests
	}
	if req.RateLimitWindow != nil {
		currentSettings.RateLimitWindow = *req.RateLimitWindow
	}
	if req.XrayConfigTemplate != nil {
		currentSettings.XrayConfigTemplate = *req.XrayConfigTemplate
	}

	// Save updated settings
	if err := h.settingsService.UpdateSystemSettings(ctx, currentSettings); err != nil {
		h.logger.Error("Failed to update settings", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("update settings", err))
		return
	}

	h.logger.Info("Settings updated")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "settings updated",
		"data":    currentSettings,
	})
}

// BackupSettings creates a backup of all settings.
func (h *SettingsHandler) BackupSettings(c *gin.Context) {
	ctx := c.Request.Context()

	backup, err := h.settingsService.Backup(ctx)
	if err != nil {
		h.logger.Error("Failed to backup settings", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("backup settings", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "backup created",
		"data": gin.H{
			"backup": string(backup),
		},
	})
}

// RestoreSettingsRequest represents a restore settings request.
type RestoreSettingsRequest struct {
	Backup string `json:"backup" binding:"required"`
}

// RestoreSettings restores settings from a backup.
func (h *SettingsHandler) RestoreSettings(c *gin.Context) {
	ctx := c.Request.Context()

	var req RestoreSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid request", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	if err := h.settingsService.Restore(ctx, []byte(req.Backup)); err != nil {
		h.logger.Error("Failed to restore settings", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("restore settings", err))
		return
	}

	h.logger.Info("Settings restored from backup")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "settings restored",
	})
}
