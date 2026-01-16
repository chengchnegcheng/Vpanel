// Package handlers provides HTTP request handlers for the V Panel API.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"v/internal/logger"
	"v/internal/node"
)

// NodeHandler handles node management API requests.
type NodeHandler struct {
	nodeService *node.Service
	logger      logger.Logger
}

// NewNodeHandler creates a new node handler.
func NewNodeHandler(nodeService *node.Service, log logger.Logger) *NodeHandler {
	return &NodeHandler{
		nodeService: nodeService,
		logger:      log,
	}
}

// NodeResponse represents a node in API responses.
type NodeResponse struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Address      string   `json:"address"`
	Port         int      `json:"port"`
	Status       string   `json:"status"`
	Tags         []string `json:"tags"`
	Region       string   `json:"region"`
	Weight       int      `json:"weight"`
	MaxUsers     int      `json:"max_users"`
	CurrentUsers int      `json:"current_users"`
	Latency      int      `json:"latency"`
	LastSeenAt   string   `json:"last_seen_at,omitempty"`
	SyncStatus   string   `json:"sync_status"`
	SyncedAt     string   `json:"synced_at,omitempty"`
	IPWhitelist  []string `json:"ip_whitelist,omitempty"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// NodeWithTokenResponse includes the token (only returned on create).
type NodeWithTokenResponse struct {
	NodeResponse
	Token string `json:"token"`
}

// CreateNodeRequest represents a request to create a node.
type CreateNodeRequest struct {
	Name        string   `json:"name" binding:"required"`
	Address     string   `json:"address" binding:"required"`
	Port        int      `json:"port"`
	Tags        []string `json:"tags"`
	Region      string   `json:"region"`
	Weight      int      `json:"weight"`
	MaxUsers    int      `json:"max_users"`
	IPWhitelist []string `json:"ip_whitelist"`
}

// UpdateNodeRequest represents a request to update a node.
type UpdateNodeRequest struct {
	Name        *string   `json:"name"`
	Address     *string   `json:"address"`
	Port        *int      `json:"port"`
	Tags        *[]string `json:"tags"`
	Region      *string   `json:"region"`
	Weight      *int      `json:"weight"`
	MaxUsers    *int      `json:"max_users"`
	IPWhitelist *[]string `json:"ip_whitelist"`
}

// toNodeResponse converts a node to API response format.
func toNodeResponse(n *node.Node) *NodeResponse {
	resp := &NodeResponse{
		ID:           n.ID,
		Name:         n.Name,
		Address:      n.Address,
		Port:         n.Port,
		Status:       n.Status,
		Tags:         n.Tags,
		Region:       n.Region,
		Weight:       n.Weight,
		MaxUsers:     n.MaxUsers,
		CurrentUsers: n.CurrentUsers,
		Latency:      n.Latency,
		SyncStatus:   n.SyncStatus,
		IPWhitelist:  n.IPWhitelist,
		CreatedAt:    n.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    n.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if n.LastSeenAt != nil {
		resp.LastSeenAt = n.LastSeenAt.Format("2006-01-02T15:04:05Z")
	}
	if n.SyncedAt != nil {
		resp.SyncedAt = n.SyncedAt.Format("2006-01-02T15:04:05Z")
	}
	if resp.Tags == nil {
		resp.Tags = []string{}
	}
	if resp.IPWhitelist == nil {
		resp.IPWhitelist = []string{}
	}
	return resp
}

// List returns all nodes with optional filtering.
// GET /api/admin/nodes
func (h *NodeHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")
	region := c.Query("region")

	filter := node.NodeFilter{
		Status: status,
		Region: region,
		Limit:  limit,
		Offset: offset,
	}

	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
		if err == nil {
			filter.GroupID = &groupID
		}
	}

	nodes, total, err := h.nodeService.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list nodes", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list nodes"})
		return
	}

	response := make([]*NodeResponse, len(nodes))
	for i, n := range nodes {
		response[i] = toNodeResponse(n)
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": response,
		"total": total,
	})
}

// Get returns a single node by ID.
// GET /api/admin/nodes/:id
func (h *NodeHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	n, err := h.nodeService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == node.ErrNodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		h.logger.Error("Failed to get node", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get node"})
		return
	}

	c.JSON(http.StatusOK, toNodeResponse(n))
}


// Create creates a new node.
// POST /api/admin/nodes
func (h *NodeHandler) Create(c *gin.Context) {
	var req CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	createReq := &node.CreateNodeRequest{
		Name:        req.Name,
		Address:     req.Address,
		Port:        req.Port,
		Tags:        req.Tags,
		Region:      req.Region,
		Weight:      req.Weight,
		MaxUsers:    req.MaxUsers,
		IPWhitelist: req.IPWhitelist,
	}

	n, err := h.nodeService.Create(c.Request.Context(), createReq)
	if err != nil {
		if err == node.ErrInvalidAddress {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node address format"})
			return
		}
		if err == node.ErrInvalidNode {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node data"})
			return
		}
		h.logger.Error("Failed to create node", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create node"})
		return
	}

	h.logger.Info("Node created", logger.F("node_id", n.ID), logger.F("name", n.Name))

	// Return response with token (only on create)
	resp := &NodeWithTokenResponse{
		NodeResponse: *toNodeResponse(n),
		Token:        n.Token,
	}

	c.JSON(http.StatusCreated, resp)
}

// Update updates an existing node.
// PUT /api/admin/nodes/:id
func (h *NodeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	var req UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updateReq := &node.UpdateNodeRequest{
		Name:        req.Name,
		Address:     req.Address,
		Port:        req.Port,
		Tags:        req.Tags,
		Region:      req.Region,
		Weight:      req.Weight,
		MaxUsers:    req.MaxUsers,
		IPWhitelist: req.IPWhitelist,
	}

	n, err := h.nodeService.Update(c.Request.Context(), id, updateReq)
	if err != nil {
		if err == node.ErrNodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		if err == node.ErrInvalidAddress {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node address format"})
			return
		}
		h.logger.Error("Failed to update node", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update node"})
		return
	}

	h.logger.Info("Node updated", logger.F("node_id", id))

	c.JSON(http.StatusOK, toNodeResponse(n))
}

// Delete deletes a node.
// DELETE /api/admin/nodes/:id
func (h *NodeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	if err := h.nodeService.Delete(c.Request.Context(), id); err != nil {
		if err == node.ErrNodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		h.logger.Error("Failed to delete node", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete node"})
		return
	}

	h.logger.Info("Node deleted", logger.F("node_id", id))

	c.JSON(http.StatusOK, gin.H{"message": "Node deleted successfully"})
}

// GenerateToken generates a new token for a node.
// POST /api/admin/nodes/:id/token
func (h *NodeHandler) GenerateToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	token, err := h.nodeService.GenerateNodeToken(c.Request.Context(), id)
	if err != nil {
		if err == node.ErrNodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		h.logger.Error("Failed to generate token", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	h.logger.Info("Token generated for node", logger.F("node_id", id))

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// RotateToken rotates a node's token.
// POST /api/admin/nodes/:id/token/rotate
func (h *NodeHandler) RotateToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	token, err := h.nodeService.RotateToken(c.Request.Context(), id)
	if err != nil {
		if err == node.ErrNodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		h.logger.Error("Failed to rotate token", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rotate token"})
		return
	}

	h.logger.Info("Token rotated for node", logger.F("node_id", id))

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// RevokeToken revokes a node's token.
// POST /api/admin/nodes/:id/token/revoke
func (h *NodeHandler) RevokeToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	if err := h.nodeService.RevokeToken(c.Request.Context(), id); err != nil {
		if err == node.ErrNodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		h.logger.Error("Failed to revoke token", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
		return
	}

	h.logger.Info("Token revoked for node", logger.F("node_id", id))

	c.JSON(http.StatusOK, gin.H{"message": "Token revoked successfully"})
}

// GetStatistics returns node statistics.
// GET /api/admin/nodes/statistics
func (h *NodeHandler) GetStatistics(c *gin.Context) {
	stats, err := h.nodeService.GetStatistics(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get node statistics", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	totalUsers, err := h.nodeService.GetTotalUsers(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get total users", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"by_status":   stats,
		"total_users": totalUsers,
	})
}

// UpdateStatus updates a node's status.
// PUT /api/admin/nodes/:id/status
func (h *NodeHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"online":    true,
		"offline":   true,
		"unhealthy": true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be one of: online, offline, unhealthy"})
		return
	}

	if err := h.nodeService.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		if err == node.ErrNodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
			return
		}
		h.logger.Error("Failed to update node status", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	h.logger.Info("Node status updated", logger.F("node_id", id), logger.F("status", req.Status))

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}
