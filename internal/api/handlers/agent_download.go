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
// DownloadAgent 下载 Agent 二进制文件
// GET /api/admin/nodes/agent/download?arch=amd64
func (h *AgentDownloadHandler) DownloadAgent(c *gin.Context) {
	arch := c.DefaultQuery("arch", "amd64")

	// 架构映射：将 uname -m 的输出映射到 Go 架构名称
	archMap := map[string]string{
		"x86_64":  "amd64",
		"aarch64": "arm64",
		"arm64":   "arm64",
		"armv7l":  "arm",
		"amd64":   "amd64",
		"arm":     "arm",
	}

	// 标准化架构名称
	if mappedArch, ok := archMap[arch]; ok {
		arch = mappedArch
	}

	// 验证架构
	validArchs := map[string]bool{
		"amd64": true,
		"arm64": true,
		"arm":   true,
	}

	if !validArchs[arch] {
		h.logger.Warn("不支持的架构",
			logger.F("arch", arch),
			logger.F("query_arch", c.Query("arch")))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不支持的架构: " + arch,
		})
		return
	}

	// Agent 二进制文件路径（尝试多个可能的路径）
	possiblePaths := []string{
		filepath.Join("bin", "vpanel-agent-"+arch),
		filepath.Join("bin", "vpanel-agent"),
		"./bin/vpanel-agent-" + arch,
		"./bin/vpanel-agent",
	}

	var agentPath string
	var fileSize int64

	// 查找存在的文件
	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil {
			agentPath = path
			fileSize = info.Size()
			break
		}
	}

	// 如果没找到文件
	if agentPath == "" {
		h.logger.Error("Agent 文件不存在",
			logger.F("arch", arch),
			logger.F("tried_paths", possiblePaths))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent 文件不存在，请先编译: cd cmd/agent && go build -o ../../bin/vpanel-agent",
		})
		return
	}

	h.logger.Info("找到 Agent 文件",
		logger.F("path", agentPath),
		logger.F("size", fileSize))

	// 返回文件
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=vpanel-agent")
	c.File(agentPath)
}
