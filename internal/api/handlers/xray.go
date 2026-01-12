package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"v/internal/logger"
	"v/internal/xray"
)

// XrayHandler handles Xray-related API requests.
type XrayHandler struct {
	manager xray.Manager
	logger  logger.Logger
}

// NewXrayHandler creates a new Xray handler.
func NewXrayHandler(manager xray.Manager, log logger.Logger) *XrayHandler {
	return &XrayHandler{
		manager: manager,
		logger:  log,
	}
}

// GetStatus returns the current Xray status.
// GET /api/xray/status
func (h *XrayHandler) GetStatus(c *gin.Context) {
	status, err := h.manager.GetStatus(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get xray status", logger.F("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Xray status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Restart restarts the Xray process.
// POST /api/xray/restart
func (h *XrayHandler) Restart(c *gin.Context) {
	if err := h.manager.Restart(c.Request.Context()); err != nil {
		h.logger.Error("failed to restart xray", logger.F("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restart Xray"})
		return
	}

	h.logger.Info("xray restarted by user")
	c.JSON(http.StatusOK, gin.H{"message": "Xray restarted successfully"})
}

// GetConfig returns the current Xray configuration.
// GET /api/xray/config
func (h *XrayHandler) GetConfig(c *gin.Context) {
	config, err := h.manager.GetConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get xray config", logger.F("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Xray config"})
		return
	}

	c.Data(http.StatusOK, "application/json", config)
}

// UpdateConfigRequest represents a config update request.
type UpdateConfigRequest struct {
	Config json.RawMessage `json:"config" binding:"required"`
}

// UpdateConfig updates the Xray configuration.
// PUT /api/xray/config
func (h *XrayHandler) UpdateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Backup current config first
	backupPath, err := h.manager.BackupConfig(c.Request.Context())
	if err != nil {
		h.logger.Warn("failed to backup config before update", logger.F("error", err))
	}

	// Update config
	if err := h.manager.UpdateConfig(c.Request.Context(), req.Config); err != nil {
		h.logger.Error("failed to update xray config", logger.F("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Reload config
	if err := h.manager.ReloadConfig(c.Request.Context()); err != nil {
		h.logger.Error("failed to reload xray config", logger.F("error", err))
		// Try to restore backup
		if backupPath != "" {
			if restoreErr := h.manager.RestoreConfig(c.Request.Context(), backupPath); restoreErr != nil {
				h.logger.Error("failed to restore backup after reload failure", logger.F("error", restoreErr))
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload config, restored backup"})
		return
	}

	h.logger.Info("xray config updated")
	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully"})
}

// GetVersion returns Xray version information.
// GET /api/xray/version
func (h *XrayHandler) GetVersion(c *gin.Context) {
	version, err := h.manager.GetVersion(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get xray version", logger.F("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Xray version"})
		return
	}

	c.JSON(http.StatusOK, version)
}

// UpdateVersionRequest represents a version update request.
type UpdateVersionRequest struct {
	Version string `json:"version"` // Optional, defaults to latest
}

// Update updates Xray to a new version.
// POST /api/xray/update
func (h *XrayHandler) Update(c *gin.Context) {
	// TODO: Implement Xray update functionality
	// This would involve:
	// 1. Downloading the new version
	// 2. Stopping Xray
	// 3. Replacing the binary
	// 4. Starting Xray
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Xray update not implemented yet"})
}

// ValidateConfig validates an Xray configuration without applying it.
// POST /api/xray/validate
func (h *XrayHandler) ValidateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.manager.ValidateConfig(c.Request.Context(), req.Config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Configuration is valid",
	})
}

// Start starts the Xray process.
// POST /api/xray/start
func (h *XrayHandler) Start(c *gin.Context) {
	if err := h.manager.Start(c.Request.Context()); err != nil {
		h.logger.Error("failed to start xray", logger.F("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("xray started by user")
	c.JSON(http.StatusOK, gin.H{"message": "Xray started successfully"})
}

// Stop stops the Xray process.
// POST /api/xray/stop
func (h *XrayHandler) Stop(c *gin.Context) {
	if err := h.manager.Stop(c.Request.Context()); err != nil {
		h.logger.Error("failed to stop xray", logger.F("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("xray stopped by user")
	c.JSON(http.StatusOK, gin.H{"message": "Xray stopped successfully"})
}
