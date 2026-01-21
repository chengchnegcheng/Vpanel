package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"v/internal/logger"
)

// AgentDownloadHandler 处理 Agent 下载请求
type AgentDownloadHandler struct {
	logger logger.Logger
}

// NewAgentDownloadHandler 创建 Agent 下载处理器
func NewAgentDownloadHandler(log logger.Logger) *AgentDownloadHandler {
	return &AgentDownloadHandler{
		logger: log,
	}
}

// DownloadAgent 下载 Agent 二进制文件
// GET /api/admin/nodes/agent/download?arch=amd64
func (h *AgentDownloadHandler) DownloadAgent(c *gin.Context) {
	arch := c.DefaultQuery("arch", "amd64")
	
	// 验证架构
	validArchs := map[string]bool{
		"amd64": true,
		"arm64": true,
		"arm":   true,
	}
	
	if !validArchs[arch] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不支持的架构: " + arch,
		})
		return
	}
	
	// Agent 二进制文件路径
	// 假设编译后的文件在 bin 目录下
	agentPath := filepath.Join("bin", "vpanel-agent-"+arch)
	
	// 检查文件是否存在
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		h.logger.Error("Agent 文件不存在",
			logger.F("arch", arch),
			logger.F("path", agentPath))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent 文件不存在，请先编译: make build-agent-" + arch,
		})
		return
	}
	
	h.logger.Info("下载 Agent",
		logger.F("arch", arch),
		logger.F("path", agentPath))
	
	// 返回文件
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=vpanel-agent")
	c.File(agentPath)
}
