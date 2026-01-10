package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"v/internal/database/repository"
	"v/internal/logger"
	"v/internal/proxy"
	"v/pkg/errors"
)

type ProxyHandler struct {
	proxyManager proxy.Manager
	proxyRepo    repository.ProxyRepository
	logger       logger.Logger
}

func NewProxyHandler(proxyManager proxy.Manager, proxyRepo repository.ProxyRepository, log logger.Logger) *ProxyHandler {
	return &ProxyHandler{
		proxyManager: proxyManager,
		proxyRepo:    proxyRepo,
		logger:       log,
	}
}

type ProxyResponse struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Protocol  string         `json:"protocol"`
	Port      int            `json:"port"`
	Host      string         `json:"host,omitempty"`
	Settings  map[string]any `json:"settings,omitempty"`
	Enabled   bool           `json:"enabled"`
	Remark    string         `json:"remark,omitempty"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

func (h *ProxyHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	proxies, err := h.proxyRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("failed to list proxies", logger.F("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list proxies"})
		return
	}

	response := make([]ProxyResponse, len(proxies))
	for i, p := range proxies {
		response[i] = ProxyResponse{
			ID:        p.ID,
			Name:      p.Name,
			Protocol:  p.Protocol,
			Port:      p.Port,
			Host:      p.Host,
			Settings:  p.Settings,
			Enabled:   p.Enabled,
			Remark:    p.Remark,
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, response)
}

type CreateProxyRequest struct {
	Name     string         `json:"name" binding:"required"`
	Protocol string         `json:"protocol" binding:"required"`
	Port     int            `json:"port" binding:"required,min=1,max=65535"`
	Host     string         `json:"host"`
	Settings map[string]any `json:"settings"`
	Enabled  bool           `json:"enabled"`
	Remark   string         `json:"remark"`
}

func (h *ProxyHandler) Create(c *gin.Context) {
	var req CreateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	protocol, ok := h.proxyManager.GetProtocol(req.Protocol)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported protocol"})
		return
	}

	settings := &proxy.Settings{
		Name:     req.Name,
		Protocol: req.Protocol,
		Port:     req.Port,
		Host:     req.Host,
		Settings: req.Settings,
		Enabled:  req.Enabled,
		Remark:   req.Remark,
	}

	if settings.Settings == nil {
		settings.Settings = protocol.DefaultSettings()
	}

	if err := protocol.Validate(settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proxyModel := &repository.Proxy{
		Name:     req.Name,
		Protocol: req.Protocol,
		Port:     req.Port,
		Host:     req.Host,
		Settings: req.Settings,
		Enabled:  req.Enabled,
		Remark:   req.Remark,
	}

	if err := h.proxyRepo.Create(c.Request.Context(), proxyModel); err != nil {
		h.logger.Error("failed to create proxy", logger.F("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy"})
		return
	}

	c.JSON(http.StatusCreated, ProxyResponse{
		ID:        proxyModel.ID,
		Name:      proxyModel.Name,
		Protocol:  proxyModel.Protocol,
		Port:      proxyModel.Port,
		Host:      proxyModel.Host,
		Settings:  proxyModel.Settings,
		Enabled:   proxyModel.Enabled,
		Remark:    proxyModel.Remark,
		CreatedAt: proxyModel.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: proxyModel.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ProxyHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	p, err := h.proxyRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get proxy"})
		return
	}

	c.JSON(http.StatusOK, ProxyResponse{
		ID:        p.ID,
		Name:      p.Name,
		Protocol:  p.Protocol,
		Port:      p.Port,
		Host:      p.Host,
		Settings:  p.Settings,
		Enabled:   p.Enabled,
		Remark:    p.Remark,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

type UpdateProxyRequest struct {
	Name     string         `json:"name"`
	Port     int            `json:"port"`
	Host     string         `json:"host"`
	Settings map[string]any `json:"settings"`
	Enabled  *bool          `json:"enabled"`
	Remark   string         `json:"remark"`
}

func (h *ProxyHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var req UpdateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	p, err := h.proxyRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get proxy"})
		return
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Port > 0 {
		p.Port = req.Port
	}
	if req.Host != "" {
		p.Host = req.Host
	}
	if req.Settings != nil {
		p.Settings = req.Settings
	}
	if req.Enabled != nil {
		p.Enabled = *req.Enabled
	}
	if req.Remark != "" {
		p.Remark = req.Remark
	}

	if err := h.proxyRepo.Update(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update proxy"})
		return
	}

	c.JSON(http.StatusOK, ProxyResponse{
		ID:        p.ID,
		Name:      p.Name,
		Protocol:  p.Protocol,
		Port:      p.Port,
		Host:      p.Host,
		Settings:  p.Settings,
		Enabled:   p.Enabled,
		Remark:    p.Remark,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ProxyHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	if err := h.proxyRepo.Delete(c.Request.Context(), id); err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete proxy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proxy deleted successfully"})
}

func (h *ProxyHandler) GetShareLink(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	p, err := h.proxyRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get proxy"})
		return
	}

	protocol, ok := h.proxyManager.GetProtocol(p.Protocol)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported protocol"})
		return
	}

	settings := &proxy.Settings{
		ID:       p.ID,
		Name:     p.Name,
		Protocol: p.Protocol,
		Port:     p.Port,
		Host:     p.Host,
		Settings: p.Settings,
		Enabled:  p.Enabled,
		Remark:   p.Remark,
	}

	link, err := protocol.GenerateLink(settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate share link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"link": link})
}

func (h *ProxyHandler) Toggle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	p, err := h.proxyRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get proxy"})
		return
	}

	p.Enabled = !p.Enabled

	if err := h.proxyRepo.Update(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle proxy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"enabled": p.Enabled})
}
