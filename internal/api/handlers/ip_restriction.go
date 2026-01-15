// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"v/internal/api/middleware"
	"v/internal/ip"
	"v/internal/logger"
	"v/pkg/errors"
)

// IPRestrictionHandler handles IP restriction related requests.
type IPRestrictionHandler struct {
	logger    logger.Logger
	ipService *ip.Service
}

// NewIPRestrictionHandler creates a new IPRestrictionHandler.
func NewIPRestrictionHandler(log logger.Logger, ipService *ip.Service) *IPRestrictionHandler {
	return &IPRestrictionHandler{
		logger:    log,
		ipService: ipService,
	}
}

// GetStats returns IP restriction statistics.
// GET /api/admin/ip-restrictions/stats
func (h *IPRestrictionHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get global statistics
	var totalActiveIPs int64
	var totalBlacklisted int64
	var totalWhitelisted int64

	db := h.ipService.Tracker().GetDB()
	db.WithContext(ctx).Model(&ip.ActiveIP{}).Count(&totalActiveIPs)
	db.WithContext(ctx).Model(&ip.IPBlacklist{}).Count(&totalBlacklisted)
	db.WithContext(ctx).Model(&ip.IPWhitelist{}).Count(&totalWhitelisted)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total_active_ips":   totalActiveIPs,
			"total_blacklisted":  totalBlacklisted,
			"total_whitelisted":  totalWhitelisted,
			"settings":           h.ipService.GetSettings(),
		},
	})
}

// GetUserOnlineIPs returns online IPs for a specific user.
// GET /api/admin/users/:id/online-ips
func (h *IPRestrictionHandler) GetUserOnlineIPs(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid user ID", nil))
		return
	}

	onlineIPs, err := h.ipService.GetOnlineIPs(ctx, uint(userID))
	if err != nil {
		h.logger.Error("Failed to get online IPs", logger.F("error", err), logger.F("user_id", userID))
		middleware.RespondWithError(c, errors.NewDatabaseError("get online IPs", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    onlineIPs,
	})
}

// KickIPRequest represents a request to kick an IP.
type KickIPRequest struct {
	IP            string `json:"ip" binding:"required"`
	AddToBlacklist bool   `json:"add_to_blacklist"`
	BlockDuration  int    `json:"block_duration"` // minutes
}

// KickUserIP kicks a specific IP for a user.
// POST /api/admin/users/:id/kick-ip
func (h *IPRestrictionHandler) KickUserIP(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid user ID", nil))
		return
	}

	var req KickIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid request", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	blockDuration := time.Duration(req.BlockDuration) * time.Minute
	if err := h.ipService.KickIP(ctx, uint(userID), req.IP, req.AddToBlacklist, blockDuration); err != nil {
		h.logger.Error("Failed to kick IP", logger.F("error", err), logger.F("user_id", userID), logger.F("ip", req.IP))
		middleware.RespondWithError(c, errors.NewDatabaseError("kick IP", err))
		return
	}

	h.logger.Info("IP kicked", logger.F("user_id", userID), logger.F("ip", req.IP))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "IP kicked successfully",
	})
}


// WhitelistEntry represents a whitelist entry request.
type WhitelistEntry struct {
	IP          string `json:"ip"`
	CIDR        string `json:"cidr"`
	UserID      *uint  `json:"user_id"`
	Description string `json:"description"`
}

// GetWhitelist returns the IP whitelist.
// GET /api/admin/ip-whitelist
func (h *IPRestrictionHandler) GetWhitelist(c *gin.Context) {
	ctx := c.Request.Context()

	var userID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := strconv.ParseUint(userIDStr, 10, 64)
		if err == nil {
			uid := uint(id)
			userID = &uid
		}
	}

	entries, err := h.ipService.Validator().GetWhitelist(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get whitelist", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("get whitelist", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    entries,
	})
}

// AddWhitelistRequest represents a request to add to whitelist.
type AddWhitelistRequest struct {
	IP          string `json:"ip"`
	CIDR        string `json:"cidr"`
	UserID      *uint  `json:"user_id"`
	Description string `json:"description"`
}

// AddWhitelist adds an IP to the whitelist.
// POST /api/admin/ip-whitelist
func (h *IPRestrictionHandler) AddWhitelist(c *gin.Context) {
	ctx := c.Request.Context()

	var req AddWhitelistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid request", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	if req.IP == "" && req.CIDR == "" {
		middleware.RespondWithError(c, errors.NewValidationError("IP or CIDR is required", nil))
		return
	}

	// Get current user ID from context
	currentUserID := middleware.GetUserID(c)

	entry := &ip.IPWhitelist{
		IP:          req.IP,
		CIDR:        req.CIDR,
		UserID:      req.UserID,
		Description: req.Description,
		CreatedBy:   uint(currentUserID),
	}

	if err := h.ipService.Validator().AddToWhitelist(ctx, entry); err != nil {
		h.logger.Error("Failed to add to whitelist", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("add to whitelist", err))
		return
	}

	h.logger.Info("IP added to whitelist", logger.F("ip", req.IP), logger.F("cidr", req.CIDR))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "added to whitelist",
		"data":    entry,
	})
}

// DeleteWhitelist removes an IP from the whitelist.
// DELETE /api/admin/ip-whitelist/:id
func (h *IPRestrictionHandler) DeleteWhitelist(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid ID", nil))
		return
	}

	if err := h.ipService.Validator().RemoveFromWhitelist(ctx, uint(id)); err != nil {
		h.logger.Error("Failed to remove from whitelist", logger.F("error", err), logger.F("id", id))
		middleware.RespondWithError(c, errors.NewDatabaseError("remove from whitelist", err))
		return
	}

	h.logger.Info("IP removed from whitelist", logger.F("id", id))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "removed from whitelist",
	})
}

// ImportWhitelistRequest represents a request to import whitelist.
type ImportWhitelistRequest struct {
	IPs         []string `json:"ips" binding:"required"`
	UserID      *uint    `json:"user_id"`
	Description string   `json:"description"`
}

// ImportWhitelist imports multiple IPs to the whitelist.
// POST /api/admin/ip-whitelist/import
func (h *IPRestrictionHandler) ImportWhitelist(c *gin.Context) {
	ctx := c.Request.Context()

	var req ImportWhitelistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid request", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	currentUserID := middleware.GetUserID(c)

	if err := h.ipService.Validator().ImportWhitelist(ctx, req.IPs, req.UserID, req.Description, uint(currentUserID)); err != nil {
		h.logger.Error("Failed to import whitelist", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("import whitelist", err))
		return
	}

	h.logger.Info("Whitelist imported", logger.F("count", len(req.IPs)))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "whitelist imported",
		"data": gin.H{
			"imported": len(req.IPs),
		},
	})
}


// GetBlacklist returns the IP blacklist.
// GET /api/admin/ip-blacklist
func (h *IPRestrictionHandler) GetBlacklist(c *gin.Context) {
	ctx := c.Request.Context()

	var userID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := strconv.ParseUint(userIDStr, 10, 64)
		if err == nil {
			uid := uint(id)
			userID = &uid
		}
	}

	entries, err := h.ipService.Validator().GetBlacklist(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get blacklist", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("get blacklist", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    entries,
	})
}

// AddBlacklistRequest represents a request to add to blacklist.
type AddBlacklistRequest struct {
	IP        string `json:"ip"`
	CIDR      string `json:"cidr"`
	UserID    *uint  `json:"user_id"`
	Reason    string `json:"reason"`
	ExpiresIn int    `json:"expires_in"` // minutes, 0 for permanent
}

// AddBlacklist adds an IP to the blacklist.
// POST /api/admin/ip-blacklist
func (h *IPRestrictionHandler) AddBlacklist(c *gin.Context) {
	ctx := c.Request.Context()

	var req AddBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid request", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	if req.IP == "" && req.CIDR == "" {
		middleware.RespondWithError(c, errors.NewValidationError("IP or CIDR is required", nil))
		return
	}

	currentUserID := middleware.GetUserID(c)
	createdBy := uint(currentUserID)

	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresIn) * time.Minute)
		expiresAt = &t
	}

	entry := &ip.IPBlacklist{
		IP:          req.IP,
		CIDR:        req.CIDR,
		UserID:      req.UserID,
		Reason:      req.Reason,
		ExpiresAt:   expiresAt,
		IsAutomatic: false,
		CreatedBy:   &createdBy,
	}

	if err := h.ipService.Validator().AddToBlacklist(ctx, entry); err != nil {
		h.logger.Error("Failed to add to blacklist", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("add to blacklist", err))
		return
	}

	h.logger.Info("IP added to blacklist", logger.F("ip", req.IP), logger.F("cidr", req.CIDR))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "added to blacklist",
		"data":    entry,
	})
}

// DeleteBlacklist removes an IP from the blacklist.
// DELETE /api/admin/ip-blacklist/:id
func (h *IPRestrictionHandler) DeleteBlacklist(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid ID", nil))
		return
	}

	if err := h.ipService.Validator().RemoveFromBlacklist(ctx, uint(id)); err != nil {
		h.logger.Error("Failed to remove from blacklist", logger.F("error", err), logger.F("id", id))
		middleware.RespondWithError(c, errors.NewDatabaseError("remove from blacklist", err))
		return
	}

	h.logger.Info("IP removed from blacklist", logger.F("id", id))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "removed from blacklist",
	})
}

// GetIPRestrictionSettings returns IP restriction settings.
// GET /api/admin/settings/ip-restriction
func (h *IPRestrictionHandler) GetIPRestrictionSettings(c *gin.Context) {
	settings := h.ipService.GetSettings()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    settings,
	})
}

// UpdateIPRestrictionSettings updates IP restriction settings.
// PUT /api/admin/settings/ip-restriction
func (h *IPRestrictionHandler) UpdateIPRestrictionSettings(c *gin.Context) {
	ctx := c.Request.Context()

	var settings ip.IPRestrictionSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		middleware.RespondWithError(c, errors.NewValidationError("invalid request", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	if err := h.ipService.SaveSettings(ctx, &settings); err != nil {
		h.logger.Error("Failed to save IP restriction settings", logger.F("error", err))
		middleware.RespondWithError(c, errors.NewDatabaseError("save settings", err))
		return
	}

	h.logger.Info("IP restriction settings updated")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "settings updated",
		"data":    settings,
	})
}


// ===== User API Endpoints =====

// GetUserDevices returns the current user's online devices.
// GET /api/user/devices
func (h *IPRestrictionHandler) GetUserDevices(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == 0 {
		middleware.RespondWithError(c, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	onlineIPs, err := h.ipService.GetOnlineIPs(ctx, uint(userID))
	if err != nil {
		h.logger.Error("Failed to get user devices", logger.F("error", err), logger.F("user_id", userID))
		middleware.RespondWithError(c, errors.NewDatabaseError("get devices", err))
		return
	}

	// Get user's max concurrent IPs (would need to be passed or fetched)
	// For now, use default from settings
	settings := h.ipService.GetSettings()
	maxDevices := settings.DefaultMaxConcurrentIPs

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"devices":         onlineIPs,
			"max_devices":     maxDevices,
			"current_count":   len(onlineIPs),
			"remaining_slots": maxDevices - len(onlineIPs),
		},
	})
}

// UserKickDeviceRequest represents a request to kick a device.
type UserKickDeviceRequest struct {
	AddToBlacklist bool `json:"add_to_blacklist"`
	BlockDuration  int  `json:"block_duration"` // minutes
}

// KickUserDevice kicks a specific device for the current user.
// POST /api/user/devices/:ip/kick
func (h *IPRestrictionHandler) KickUserDevice(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == 0 {
		middleware.RespondWithError(c, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	ipAddr := c.Param("ip")
	if ipAddr == "" {
		middleware.RespondWithError(c, errors.NewValidationError("IP address is required", nil))
		return
	}

	var req UserKickDeviceRequest
	// Bind JSON if provided, otherwise use defaults
	_ = c.ShouldBindJSON(&req)

	blockDuration := time.Duration(req.BlockDuration) * time.Minute
	if err := h.ipService.KickIP(ctx, uint(userID), ipAddr, req.AddToBlacklist, blockDuration); err != nil {
		h.logger.Error("Failed to kick device", logger.F("error", err), logger.F("user_id", userID), logger.F("ip", ipAddr))
		middleware.RespondWithError(c, errors.NewDatabaseError("kick device", err))
		return
	}

	h.logger.Info("User kicked device", logger.F("user_id", userID), logger.F("ip", ipAddr))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "device kicked successfully",
	})
}

// GetUserIPStats returns IP statistics for the current user.
// GET /api/user/ip-stats
func (h *IPRestrictionHandler) GetUserIPStats(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == 0 {
		middleware.RespondWithError(c, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Get user's max concurrent IPs (would need to be passed or fetched)
	settings := h.ipService.GetSettings()
	maxConcurrentIPs := settings.DefaultMaxConcurrentIPs

	stats, err := h.ipService.GetIPStats(ctx, uint(userID), maxConcurrentIPs)
	if err != nil {
		h.logger.Error("Failed to get IP stats", logger.F("error", err), logger.F("user_id", userID))
		middleware.RespondWithError(c, errors.NewDatabaseError("get IP stats", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// GetUserIPHistory returns IP history for the current user.
// GET /api/user/ip-history
func (h *IPRestrictionHandler) GetUserIPHistory(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == 0 {
		middleware.RespondWithError(c, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Parse query parameters
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	filter := &ip.IPHistoryFilter{
		Limit:  limit,
		Offset: offset,
	}

	history, err := h.ipService.Tracker().GetIPHistory(ctx, uint(userID), filter)
	if err != nil {
		h.logger.Error("Failed to get IP history", logger.F("error", err), logger.F("user_id", userID))
		middleware.RespondWithError(c, errors.NewDatabaseError("get IP history", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    history,
	})
}
