// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"v/internal/logger"
	"v/internal/portal/node"
)

// PortalNodeHandler handles portal node requests.
type PortalNodeHandler struct {
	nodeService *node.Service
	logger      logger.Logger
}

// NewPortalNodeHandler creates a new PortalNodeHandler.
func NewPortalNodeHandler(nodeService *node.Service, log logger.Logger) *PortalNodeHandler {
	return &PortalNodeHandler{
		nodeService: nodeService,
		logger:      log,
	}
}

// ListNodes returns available nodes for the current user.
func (h *PortalNodeHandler) ListNodes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// Parse filter parameters
	filter := &node.NodeFilter{
		Region:   c.Query("region"),
		Protocol: c.Query("protocol"),
	}

	// Get nodes
	nodes, err := h.nodeService.ListNodes(c.Request.Context(), userID.(int64), filter)
	if err != nil {
		h.logger.Error("failed to list nodes", logger.F("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取节点列表失败"})
		return
	}

	// Apply sorting if specified
	sortField := c.Query("sort")
	sortOrder := c.Query("order")
	if sortField != "" {
		sortOpt := &node.SortOption{
			Field: sortField,
			Order: sortOrder,
		}
		nodes = node.SortNodes(nodes, sortOpt)
	}

	// Get available regions and protocols for filtering
	regions := node.GetAvailableRegions(nodes)
	protocols := node.GetAvailableProtocols(nodes)

	c.JSON(http.StatusOK, gin.H{
		"nodes":     nodes,
		"total":     len(nodes),
		"regions":   regions,
		"protocols": protocols,
	})
}

// GetNode returns a single node by ID.
func (h *PortalNodeHandler) GetNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的节点ID"})
		return
	}

	nodeInfo, err := h.nodeService.GetNode(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}

	c.JSON(http.StatusOK, nodeInfo)
}

// TestLatencyRequest represents a latency test request.
type TestLatencyRequest struct {
	NodeIDs []int64 `json:"node_ids"`
}

// TestLatency tests latency to specified nodes.
// Note: This is a placeholder - actual latency testing would require
// client-side implementation or server-side ping functionality.
func (h *PortalNodeHandler) TestLatency(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的节点ID"})
		return
	}

	// Get node to verify it exists
	nodeInfo, err := h.nodeService.GetNode(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}

	// In a real implementation, this would perform actual latency testing
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"node_id": nodeInfo.ID,
		"host":    nodeInfo.Host,
		"latency": -1, // -1 indicates not tested
		"message": "延迟测试需要客户端实现",
	})
}
