// Package handlers provides HTTP request handlers for the V Panel API.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"v/internal/config"
	"v/internal/logger"
	"v/internal/node"
)

// NodeDeployHandler handles node deployment requests.
type NodeDeployHandler struct {
	deployService *node.RemoteDeployService
	nodeService   *node.Service
	config        *config.Config
	logger        logger.Logger
}

// NewNodeDeployHandler creates a new node deploy handler.
func NewNodeDeployHandler(
	deployService *node.RemoteDeployService,
	nodeService *node.Service,
	cfg *config.Config,
	log logger.Logger,
) *NodeDeployHandler {
	return &NodeDeployHandler{
		deployService: deployService,
		nodeService:   nodeService,
		config:        cfg,
		logger:        log,
	}
}

// DeployAgentRequest represents a deploy agent request.
type DeployAgentRequest struct {
	Host       string `json:"host" binding:"required"`
	Port       int    `json:"port"`
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	PanelURL   string `json:"panel_url"` // Panel 服务器地址（可选，优先使用此值）
}

// DeployAgent deploys the agent to a remote server.
// POST /api/admin/nodes/:id/deploy
func (h *NodeDeployHandler) DeployAgent(c *gin.Context) {
	nodeIDStr := c.Param("id")
	nodeID, err := strconv.ParseInt(nodeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid node ID",
		})
		return
	}

	var req DeployAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate authentication method
	if req.Password == "" && req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Either password or private_key must be provided",
		})
		return
	}

	// Get node info
	nodeData, err := h.nodeService.GetByID(c.Request.Context(), nodeID)
	if err != nil {
		h.logger.Error("Failed to get node",
			logger.F("node_id", nodeID),
			logger.F("error", err.Error()))
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Node not found",
		})
		return
	}

	// Generate token if not exists
	if nodeData.Token == "" {
		token, err := h.nodeService.GenerateNodeToken(c.Request.Context(), nodeID)
		if err != nil {
			h.logger.Error("Failed to generate token",
				logger.F("node_id", nodeID),
				logger.F("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to generate node token",
			})
			return
		}
		nodeData.Token = token
	}

	// Get panel URL - 优先级：节点保存的 PanelURL > 请求参数 > 配置文件 > 请求头 > 请求 Host
	panelURL := nodeData.PanelURL // 优先使用节点创建时保存的 Panel URL
	if panelURL == "" {
		panelURL = req.PanelURL
	}
	if panelURL == "" {
		panelURL = h.config.Server.PublicURL
	}
	if panelURL == "" {
		panelURL = c.Request.Header.Get("X-Panel-URL")
	}
	if panelURL == "" {
		// 最后才使用请求的 Host（可能是 localhost）
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		panelURL = scheme + "://" + c.Request.Host
	}
	
	h.logger.Info("Using Panel URL for deployment",
		logger.F("node_id", nodeID),
		logger.F("panel_url", panelURL))

	// Prepare deploy config
	deployConfig := &node.DeployConfig{
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
		PanelURL:   panelURL,
		NodeToken:  nodeData.Token,
	}

	h.logger.Info("Starting agent deployment",
		logger.F("node_id", nodeID),
		logger.F("host", req.Host),
		logger.F("username", req.Username))

	// Deploy agent
	result, err := h.deployService.Deploy(c.Request.Context(), deployConfig)
	if err != nil {
		h.logger.Error("Agent deployment failed",
			logger.F("node_id", nodeID),
			logger.F("error", err.Error()))
		
		// 返回 200 状态码，但 success 为 false，这样前端可以正确显示错误信息和日志
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": result.Message,
			"steps":   result.Steps,
			"logs":    result.Logs,
		})
		return
	}

	h.logger.Info("Agent deployed successfully",
		logger.F("node_id", nodeID),
		logger.F("host", req.Host))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": result.Message,
		"steps":   result.Steps,
		"logs":    result.Logs,
	})
}

// GetDeployScript returns the deployment script for manual installation.
// GET /api/admin/nodes/:id/deploy/script
func (h *NodeDeployHandler) GetDeployScript(c *gin.Context) {
	nodeIDStr := c.Param("id")
	nodeID, err := strconv.ParseInt(nodeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid node ID",
		})
		return
	}

	// Get node info
	nodeData, err := h.nodeService.GetByID(c.Request.Context(), nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Node not found",
		})
		return
	}

	// Generate token if not exists
	if nodeData.Token == "" {
		token, err := h.nodeService.GenerateNodeToken(c.Request.Context(), nodeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to generate node token",
			})
			return
		}
		nodeData.Token = token
	}

	// Get panel URL from config or query parameter
	panelURL := h.config.Server.PublicURL
	if queryURL := c.Query("panel_url"); queryURL != "" {
		panelURL = queryURL
	}
	if panelURL == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		panelURL = scheme + "://" + c.Request.Host
	}

	// Generate script
	script := h.deployService.GetDeployScript(panelURL, nodeData.Token)

	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", "attachment; filename=install-agent.sh")
	c.String(http.StatusOK, script)
}

// TestConnection tests SSH connection to a remote server.
// POST /api/admin/nodes/test-connection
func (h *NodeDeployHandler) TestConnection(c *gin.Context) {
	var req DeployAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate authentication method
	if req.Password == "" && req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Either password or private_key must be provided",
		})
		return
	}

	deployConfig := &node.DeployConfig{
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
	}

	// Try to connect using a test deployment
	err := h.deployService.TestConnection(c.Request.Context(), deployConfig)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Connection failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Connection successful",
	})
}
