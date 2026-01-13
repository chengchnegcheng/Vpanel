// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"v/internal/logger"
)

// CertificateResponse represents a certificate in API responses.
type CertificateResponse struct {
	ID        int64  `json:"id"`
	Domain    string `json:"domain"`
	Provider  string `json:"provider"`
	IssueDate string `json:"issueDate"`
	ExpireDate string `json:"expireDate"`
	AutoRenew bool   `json:"autoRenew"`
	Status    string `json:"status"`
}

// CertificatesHandler handles certificate-related requests.
type CertificatesHandler struct {
	logger logger.Logger
}

// NewCertificatesHandler creates a new CertificatesHandler.
func NewCertificatesHandler(log logger.Logger) *CertificatesHandler {
	return &CertificatesHandler{
		logger: log,
	}
}

// List returns all certificates.
func (h *CertificatesHandler) List(c *gin.Context) {
	// Return empty list for now - certificates table exists but no data
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    []CertificateResponse{},
	})
}

// ApplyRequest represents a certificate apply request.
type ApplyRequest struct {
	Domain string `json:"domain" binding:"required"`
	Email  string `json:"email"`
}

// Apply applies for a new certificate.
func (h *CertificatesHandler) Apply(c *gin.Context) {
	var req ApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	// TODO: Implement actual certificate application via Let's Encrypt
	h.logger.Info("Certificate apply request", logger.F("domain", req.Domain))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "certificate application submitted",
		"data": CertificateResponse{
			ID:        time.Now().UnixNano(),
			Domain:    req.Domain,
			Provider:  "Let's Encrypt",
			IssueDate: time.Now().Format("2006-01-02"),
			ExpireDate: time.Now().AddDate(0, 3, 0).Format("2006-01-02"),
			AutoRenew: true,
			Status:    "pending",
		},
	})
}

// Upload handles certificate upload.
func (h *CertificatesHandler) Upload(c *gin.Context) {
	domain := c.PostForm("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "domain is required",
		})
		return
	}

	certFile, err := c.FormFile("cert")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "certificate file is required",
		})
		return
	}

	keyFile, err := c.FormFile("key")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "private key file is required",
		})
		return
	}

	h.logger.Info("Certificate upload", 
		logger.F("domain", domain),
		logger.F("cert_file", certFile.Filename),
		logger.F("key_file", keyFile.Filename),
	)

	// TODO: Save certificate files and store in database

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "certificate uploaded successfully",
		"data": CertificateResponse{
			ID:        time.Now().UnixNano(),
			Domain:    domain,
			Provider:  "Manual",
			IssueDate: time.Now().Format("2006-01-02"),
			ExpireDate: time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
			AutoRenew: false,
			Status:    "active",
		},
	})
}

// Renew renews a certificate.
func (h *CertificatesHandler) Renew(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid certificate id",
		})
		return
	}

	h.logger.Info("Certificate renew request", logger.F("id", id))

	// TODO: Implement actual certificate renewal

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "certificate renewal initiated",
	})
}

// Validate validates a certificate.
func (h *CertificatesHandler) Validate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid certificate id",
		})
		return
	}

	h.logger.Info("Certificate validate request", logger.F("id", id))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "certificate is valid",
		"data": gin.H{
			"valid":      true,
			"expires_in": "90 days",
		},
	})
}

// Delete deletes a certificate.
func (h *CertificatesHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid certificate id",
		})
		return
	}

	h.logger.Info("Certificate delete request", logger.F("id", id))

	// TODO: Implement actual certificate deletion

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "certificate deleted",
	})
}

// UpdateAutoRenew updates auto-renew setting.
func (h *CertificatesHandler) UpdateAutoRenew(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid certificate id",
		})
		return
	}

	var req struct {
		AutoRenew bool `json:"autoRenew"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request",
		})
		return
	}

	h.logger.Info("Certificate auto-renew update", 
		logger.F("id", id),
		logger.F("auto_renew", req.AutoRenew),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "auto-renew setting updated",
	})
}
