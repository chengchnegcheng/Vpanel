// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"v/internal/config"
	"v/internal/logger"
)

// SystemHandler handles system-related requests.
type SystemHandler struct {
	config    *config.Config
	logger    logger.Logger
	startTime time.Time
}

// NewSystemHandler creates a new SystemHandler.
func NewSystemHandler(cfg *config.Config, log logger.Logger) *SystemHandler {
	return &SystemHandler{
		config:    cfg,
		logger:    log,
		startTime: time.Now(),
	}
}

// SystemInfoResponse represents system information.
type SystemInfoResponse struct {
	Hostname    string `json:"hostname"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	GoVersion   string `json:"go_version"`
	NumCPU      int    `json:"num_cpu"`
	Goroutines  int    `json:"goroutines"`
	Uptime      string `json:"uptime"`
	UptimeSecs  int64  `json:"uptime_secs"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}

// GetInfo returns system information.
func (h *SystemHandler) GetInfo(c *gin.Context) {
	hostname, _ := os.Hostname()
	uptime := time.Since(h.startTime)

	c.JSON(http.StatusOK, SystemInfoResponse{
		Hostname:    hostname,
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		GoVersion:   runtime.Version(),
		NumCPU:      runtime.NumCPU(),
		Goroutines:  runtime.NumGoroutine(),
		Uptime:      uptime.Round(time.Second).String(),
		UptimeSecs:  int64(uptime.Seconds()),
		Version:     h.config.Version,
		Environment: h.config.Server.Mode,
	})
}

// SystemStatusResponse represents system status.
type SystemStatusResponse struct {
	Status     string      `json:"status"`
	CPU        CPUInfo     `json:"cpu"`
	Memory     MemoryInfo  `json:"memory"`
	Goroutines int         `json:"goroutines"`
	Uptime     string      `json:"uptime"`
}

// CPUInfo represents CPU information.
type CPUInfo struct {
	Cores int     `json:"cores"`
	Usage float64 `json:"usage"`
}

// MemoryInfo represents memory information.
type MemoryInfo struct {
	Alloc      uint64  `json:"alloc"`
	TotalAlloc uint64  `json:"total_alloc"`
	Sys        uint64  `json:"sys"`
	HeapAlloc  uint64  `json:"heap_alloc"`
	HeapSys    uint64  `json:"heap_sys"`
	UsagePercent float64 `json:"usage_percent"`
}

// GetStatus returns system status.
func (h *SystemHandler) GetStatus(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(h.startTime)

	// Calculate memory usage percentage (heap used / heap sys)
	usagePercent := float64(0)
	if m.HeapSys > 0 {
		usagePercent = float64(m.HeapAlloc) / float64(m.HeapSys) * 100
	}

	c.JSON(http.StatusOK, SystemStatusResponse{
		Status: "running",
		CPU: CPUInfo{
			Cores: runtime.NumCPU(),
			Usage: 0, // Would need external package for actual CPU usage
		},
		Memory: MemoryInfo{
			Alloc:        m.Alloc / 1024 / 1024,      // MB
			TotalAlloc:   m.TotalAlloc / 1024 / 1024, // MB
			Sys:          m.Sys / 1024 / 1024,        // MB
			HeapAlloc:    m.HeapAlloc / 1024 / 1024,  // MB
			HeapSys:      m.HeapSys / 1024 / 1024,    // MB
			UsagePercent: usagePercent,
		},
		Goroutines: runtime.NumGoroutine(),
		Uptime:     uptime.Round(time.Second).String(),
	})
}

// SystemStatsResponse represents system statistics.
type SystemStatsResponse struct {
	TotalProxies   int64 `json:"total_proxies"`
	ActiveProxies  int64 `json:"active_proxies"`
	TotalUsers     int64 `json:"total_users"`
	TotalTraffic   int64 `json:"total_traffic"`
	UploadTraffic  int64 `json:"upload_traffic"`
	DownloadTraffic int64 `json:"download_traffic"`
}

// GetStats returns system statistics.
func (h *SystemHandler) GetStats(c *gin.Context) {
	// In a real implementation, these would come from the database
	// For now, return placeholder values
	c.JSON(http.StatusOK, SystemStatsResponse{
		TotalProxies:    0,
		ActiveProxies:   0,
		TotalUsers:      0,
		TotalTraffic:    0,
		UploadTraffic:   0,
		DownloadTraffic: 0,
	})
}
