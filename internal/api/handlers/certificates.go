// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"v/internal/database/repository"
	"v/internal/logger"
	"v/pkg/errors"
)

// CertificateHandler handles certificate management requests.
type CertificateHandler struct {
	certRepo repository.CertificateRepository
	logger   logger.Logger
}

// NewCertificateHandler creates a new certificate handler.
func NewCertificateHandler(certRepo repository.CertificateRepository, log logger.Logger) *CertificateHandler {
	return &CertificateHandler{
		certRepo: certRepo,
		logger:   log,
	}
}

// CertificateResponse represents a certificate in API responses.
type CertificateResponse struct {
	ID          int64  `json:"id"`
	Domain      string `json:"domain"`
	AutoRenew   bool   `json:"auto_renew"`
	ExpiresAt   string `json:"expires_at"`
	DaysLeft    int    `json:"days_left"`
	Status      string `json:"status"` // valid, expiring, expired
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CreateCertificateRequest represents a request to create/upload a certificate.
type CreateCertificateRequest struct {
	Domain      string `json:"domain" binding:"required"`
	Certificate string `json:"certificate" binding:"required"`
	PrivateKey  string `json:"private_key" binding:"required"`
	AutoRenew   bool   `json:"auto_renew"`
}

// UpdateCertificateRequest represents a request to update a certificate.
type UpdateCertificateRequest struct {
	Certificate *string `json:"certificate"`
	PrivateKey  *string `json:"private_key"`
	AutoRenew   *bool   `json:"auto_renew"`
}

// ApplyCertificateRequest represents a request to apply for a certificate using ACME.
type ApplyCertificateRequest struct {
	Domain    string `json:"domain" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	AutoRenew bool   `json:"auto_renew"`
}

// toCertificateResponse converts a certificate to API response format.
func toCertificateResponse(cert *repository.Certificate) *CertificateResponse {
	daysLeft := int(time.Until(cert.ExpiresAt).Hours() / 24)
	
	status := "valid"
	if daysLeft < 0 {
		status = "expired"
	} else if daysLeft < 30 {
		status = "expiring"
	}
	
	return &CertificateResponse{
		ID:        cert.ID,
		Domain:    cert.Domain,
		AutoRenew: cert.AutoRenew,
		ExpiresAt: cert.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		DaysLeft:  daysLeft,
		Status:    status,
		CreatedAt: cert.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: cert.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// List returns all certificates.
// GET /api/admin/certificates
func (h *CertificateHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	certs, err := h.certRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list certificates", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取证书列表失败"})
		return
	}

	total, err := h.certRepo.Count(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to count certificates", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取证书数量失败"})
		return
	}

	response := make([]*CertificateResponse, len(certs))
	for i, cert := range certs {
		response[i] = toCertificateResponse(cert)
	}

	c.JSON(http.StatusOK, gin.H{
		"certificates": response,
		"total":        total,
	})
}

// Get returns a single certificate by ID.
// GET /api/admin/certificates/:id
func (h *CertificateHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的证书 ID"})
		return
	}

	cert, err := h.certRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "证书不存在"})
			return
		}
		h.logger.Error("Failed to get certificate", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取证书失败"})
		return
	}

	c.JSON(http.StatusOK, toCertificateResponse(cert))
}

// GetByDomain returns a certificate by domain.
// GET /api/admin/certificates/domain/:domain
func (h *CertificateHandler) GetByDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "域名不能为空"})
		return
	}

	cert, err := h.certRepo.GetByDomain(c.Request.Context(), domain)
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "该域名的证书不存在"})
			return
		}
		h.logger.Error("Failed to get certificate by domain", logger.Err(err), logger.F("domain", domain))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取证书失败"})
		return
	}

	c.JSON(http.StatusOK, toCertificateResponse(cert))
}

// Create creates or uploads a new certificate.
// POST /api/admin/certificates
func (h *CertificateHandler) Create(c *gin.Context) {
	var req CreateCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// TODO: 验证证书格式和有效性
	// TODO: 从证书中提取过期时间

	cert := &repository.Certificate{
		Domain:      req.Domain,
		Certificate: req.Certificate,
		PrivateKey:  req.PrivateKey,
		AutoRenew:   req.AutoRenew,
		ExpiresAt:   time.Now().AddDate(0, 3, 0), // 默认 3 个月后过期
	}

	if err := h.certRepo.Create(c.Request.Context(), cert); err != nil {
		if errors.IsConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "该域名的证书已存在"})
			return
		}
		h.logger.Error("Failed to create certificate", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建证书失败"})
		return
	}

	h.logger.Info("Certificate created", logger.F("cert_id", cert.ID), logger.F("domain", cert.Domain))

	c.JSON(http.StatusCreated, toCertificateResponse(cert))
}

// Update updates an existing certificate.
// PUT /api/admin/certificates/:id
func (h *CertificateHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的证书 ID"})
		return
	}

	var req UpdateCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	cert, err := h.certRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "证书不存在"})
			return
		}
		h.logger.Error("Failed to get certificate", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取证书失败"})
		return
	}

	// Update fields
	if req.Certificate != nil {
		cert.Certificate = *req.Certificate
		// TODO: 从新证书中提取过期时间
		cert.ExpiresAt = time.Now().AddDate(0, 3, 0)
	}
	if req.PrivateKey != nil {
		cert.PrivateKey = *req.PrivateKey
	}
	if req.AutoRenew != nil {
		cert.AutoRenew = *req.AutoRenew
	}

	if err := h.certRepo.Update(c.Request.Context(), cert); err != nil {
		h.logger.Error("Failed to update certificate", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新证书失败"})
		return
	}

	h.logger.Info("Certificate updated", logger.F("cert_id", id))

	c.JSON(http.StatusOK, toCertificateResponse(cert))
}

// Delete deletes a certificate.
// DELETE /api/admin/certificates/:id
func (h *CertificateHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的证书 ID"})
		return
	}

	if err := h.certRepo.Delete(c.Request.Context(), id); err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "证书不存在"})
			return
		}
		h.logger.Error("Failed to delete certificate", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除证书失败"})
		return
	}

	h.logger.Info("Certificate deleted", logger.F("cert_id", id))

	c.JSON(http.StatusOK, gin.H{"message": "证书删除成功"})
}

// Apply applies for a new certificate using ACME (Let's Encrypt).
// POST /api/admin/certificates/apply
func (h *CertificateHandler) Apply(c *gin.Context) {
	var req ApplyCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// TODO: 实现 ACME 证书申请逻辑
	// 1. 使用 acme.sh 或 lego 库申请证书
	// 2. 验证域名所有权（HTTP-01 或 DNS-01 challenge）
	// 3. 保存证书到数据库

	h.logger.Info("Certificate application requested",
		logger.F("domain", req.Domain),
		logger.F("email", req.Email))

	c.JSON(http.StatusAccepted, gin.H{
		"message": "证书申请已提交，请稍后查看结果",
		"domain":  req.Domain,
	})
}

// Renew renews an existing certificate.
// POST /api/admin/certificates/:id/renew
func (h *CertificateHandler) Renew(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的证书 ID"})
		return
	}

	cert, err := h.certRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "证书不存在"})
			return
		}
		h.logger.Error("Failed to get certificate", logger.Err(err), logger.F("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取证书失败"})
		return
	}

	// TODO: 实现证书续期逻辑
	// 1. 使用 ACME 续期证书
	// 2. 更新数据库中的证书内容和过期时间

	h.logger.Info("Certificate renewal requested",
		logger.F("cert_id", id),
		logger.F("domain", cert.Domain))

	c.JSON(http.StatusAccepted, gin.H{
		"message": "证书续期已提交，请稍后查看结果",
		"domain":  cert.Domain,
	})
}

// GetExpiring returns certificates that are expiring soon.
// GET /api/admin/certificates/expiring
func (h *CertificateHandler) GetExpiring(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	certs, err := h.certRepo.GetExpiring(c.Request.Context(), days)
	if err != nil {
		h.logger.Error("Failed to get expiring certificates", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取即将过期的证书失败"})
		return
	}

	response := make([]*CertificateResponse, len(certs))
	for i, cert := range certs {
		response[i] = toCertificateResponse(cert)
	}

	c.JSON(http.StatusOK, gin.H{
		"certificates": response,
		"total":        len(response),
		"days":         days,
	})
}
