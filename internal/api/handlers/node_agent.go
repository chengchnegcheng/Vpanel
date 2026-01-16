// Package handlers provides HTTP request handlers for the V Panel API.
package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"v/internal/database/repository"
	"v/internal/logger"
	"v/internal/node"
)

// NodeAgentHandler handles Node Agent API requests.
type NodeAgentHandler struct {
	nodeService *node.Service
	nodeRepo    repository.NodeRepository
	logger      logger.Logger
}

// NewNodeAgentHandler creates a new NodeAgentHandler.
func NewNodeAgentHandler(
	nodeService *node.Service,
	nodeRepo repository.NodeRepository,
	log logger.Logger,
) *NodeAgentHandler {
	return &NodeAgentHandler{
		nodeService: nodeService,
		nodeRepo:    nodeRepo,
		logger:      log,
	}
}

// RegisterRequest represents a node registration request.
type RegisterRequest struct {
	Token   string `json:"token" binding:"required"`
	Name    string `json:"name"`
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// RegisterResponse represents a node registration response.
type RegisterResponse struct {
	Success bool   `json:"success"`
	NodeID  int64  `json:"node_id"`
	Message string `json:"message"`
}

// HeartbeatRequest represents a node heartbeat request.
type HeartbeatRequest struct {
	NodeID  int64        `json:"node_id" binding:"required"`
	Token   string       `json:"token" binding:"required"`
	Metrics *NodeMetrics `json:"metrics"`
}

// NodeMetrics represents metrics from a node.
type NodeMetrics struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	MemoryTotal  uint64  `json:"memory_total"`
	MemoryUsed   uint64  `json:"memory_used"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkIn    uint64  `json:"network_in"`
	NetworkOut   uint64  `json:"network_out"`
	Connections  int     `json:"connections"`
	XrayRunning  bool    `json:"xray_running"`
	XrayVersion  string  `json:"xray_version"`
	Uptime       int64   `json:"uptime"`
	Timestamp    int64   `json:"timestamp"`
}

// HeartbeatResponse represents a node heartbeat response.
type HeartbeatResponse struct {
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	Commands []Command `json:"commands,omitempty"`
}

// Command represents a command to send to a node.
type Command struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

// CommandResultRequest represents a command result from a node.
type CommandResultRequest struct {
	CommandID string `json:"command_id" binding:"required"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
}

// Register handles node registration requests.
// POST /api/node/register
func (h *NodeAgentHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RegisterResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate token
	nodeData, err := h.nodeService.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		h.logger.Warn("Node registration failed: invalid token",
			logger.F("token_prefix", truncateToken(req.Token)),
			logger.F("error", err.Error()))
		c.JSON(http.StatusUnauthorized, RegisterResponse{
			Success: false,
			Message: "Invalid or revoked token",
		})
		return
	}

	// Update node status to online
	if err := h.nodeService.UpdateStatus(c.Request.Context(), nodeData.ID, repository.NodeStatusOnline); err != nil {
		h.logger.Error("Failed to update node status",
			logger.F("node_id", nodeData.ID),
			logger.F("error", err.Error()))
	}

	// Update last seen
	if err := h.nodeService.UpdateLastSeen(c.Request.Context(), nodeData.ID); err != nil {
		h.logger.Error("Failed to update node last seen",
			logger.F("node_id", nodeData.ID),
			logger.F("error", err.Error()))
	}

	h.logger.Info("Node registered successfully",
		logger.F("node_id", nodeData.ID),
		logger.F("node_name", nodeData.Name),
		logger.F("version", req.Version),
		logger.F("os", req.OS),
		logger.F("arch", req.Arch))

	c.JSON(http.StatusOK, RegisterResponse{
		Success: true,
		NodeID:  nodeData.ID,
		Message: "Registration successful",
	})
}

// Heartbeat handles node heartbeat requests.
// POST /api/node/heartbeat
func (h *NodeAgentHandler) Heartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, HeartbeatResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate token
	nodeData, err := h.nodeService.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		h.logger.Warn("Heartbeat failed: invalid token",
			logger.F("node_id", req.NodeID),
			logger.F("error", err.Error()))
		c.JSON(http.StatusUnauthorized, HeartbeatResponse{
			Success: false,
			Message: "Invalid or revoked token",
		})
		return
	}

	// Verify node ID matches token
	if nodeData.ID != req.NodeID {
		h.logger.Warn("Heartbeat failed: node ID mismatch",
			logger.F("expected_node_id", nodeData.ID),
			logger.F("received_node_id", req.NodeID))
		c.JSON(http.StatusUnauthorized, HeartbeatResponse{
			Success: false,
			Message: "Node ID does not match token",
		})
		return
	}

	// Update node status to online
	if err := h.nodeService.UpdateStatus(c.Request.Context(), nodeData.ID, repository.NodeStatusOnline); err != nil {
		h.logger.Error("Failed to update node status",
			logger.F("node_id", nodeData.ID),
			logger.F("error", err.Error()))
	}

	// Update last seen
	if err := h.nodeService.UpdateLastSeen(c.Request.Context(), nodeData.ID); err != nil {
		h.logger.Error("Failed to update node last seen",
			logger.F("node_id", nodeData.ID),
			logger.F("error", err.Error()))
	}

	// Update metrics if provided
	if req.Metrics != nil {
		metrics := &node.NodeMetrics{
			Latency:      0, // Will be calculated from health checks
			CurrentUsers: req.Metrics.Connections,
		}
		if err := h.nodeService.UpdateMetrics(c.Request.Context(), nodeData.ID, metrics); err != nil {
			h.logger.Error("Failed to update node metrics",
				logger.F("node_id", nodeData.ID),
				logger.F("error", err.Error()))
		}
	}

	// Get any pending commands for this node
	commands := h.getPendingCommands(nodeData.ID)

	c.JSON(http.StatusOK, HeartbeatResponse{
		Success:  true,
		Message:  "Heartbeat received",
		Commands: commands,
	})
}

// ReportCommandResult handles command result reports from nodes.
// POST /api/node/command/result
func (h *NodeAgentHandler) ReportCommandResult(c *gin.Context) {
	var req CommandResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Get token from header
	token := c.GetHeader("X-Node-Token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Missing node token",
		})
		return
	}

	// Validate token
	nodeData, err := h.nodeService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid or revoked token",
		})
		return
	}

	h.logger.Info("Command result received",
		logger.F("node_id", nodeData.ID),
		logger.F("command_id", req.CommandID),
		logger.F("success", req.Success),
		logger.F("message", req.Message))

	// TODO: Store command result for tracking

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Result received",
	})
}

// GetConfig returns the configuration for a node.
// GET /api/node/:id/config
func (h *NodeAgentHandler) GetConfig(c *gin.Context) {
	// Get token from header
	token := c.GetHeader("X-Node-Token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Missing node token",
		})
		return
	}

	// Validate token
	nodeData, err := h.nodeService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid or revoked token",
		})
		return
	}

	// TODO: Build and return node configuration
	// This would include proxy configurations, etc.

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"node_id":   nodeData.ID,
		"version":   "1.0",
		"timestamp": time.Now().Unix(),
		"proxies":   []any{},
	})
}

// GetSystemInfo returns system information for the Panel.
// GET /api/node/system-info
func (h *NodeAgentHandler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"os":           runtime.GOOS,
			"arch":         runtime.GOARCH,
			"go_version":   runtime.Version(),
			"num_cpu":      runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
		},
	})
}

// getPendingCommands returns pending commands for a node.
// In a production system, this would query a command queue.
func (h *NodeAgentHandler) getPendingCommands(nodeID int64) []Command {
	// TODO: Implement command queue
	// For now, return empty list
	return []Command{}
}

// truncateToken truncates a token for logging purposes.
func truncateToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
