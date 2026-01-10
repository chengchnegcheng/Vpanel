// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"v/internal/database/repository"
	"v/internal/logger"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	repos  *repository.Repositories
	logger logger.Logger
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(repos *repository.Repositories, log logger.Logger) *HealthHandler {
	return &HealthHandler{
		repos:  repos,
		logger: log,
	}
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Health returns a simple health check response.
// This endpoint is used by load balancers and monitoring systems.
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadyResponse represents a readiness check response.
type ReadyResponse struct {
	Status    string           `json:"status"`
	Timestamp string           `json:"timestamp"`
	Checks    map[string]Check `json:"checks"`
}

// Check represents a single health check result.
type Check struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// Ready returns a detailed readiness check response.
// This endpoint checks all dependencies and returns their status.
func (h *HealthHandler) Ready(c *gin.Context) {
	checks := make(map[string]Check)
	allHealthy := true

	// Check database connection
	dbCheck := h.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status != "ok" {
		allHealthy = false
	}

	status := "ok"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, ReadyResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	})
}

// checkDatabase checks the database connection.
func (h *HealthHandler) checkDatabase() Check {
	start := time.Now()

	// Try to perform a simple query
	if h.repos == nil || h.repos.User == nil {
		return Check{
			Status:  "error",
			Message: "database not initialized",
		}
	}

	// Try to count users as a simple health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.repos.User.List(ctx, 1, 0)
	latency := time.Since(start)

	if err != nil {
		h.logger.Warn("database health check failed", logger.F("error", err))
		return Check{
			Status:  "error",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return Check{
		Status:  "ok",
		Latency: latency.String(),
	}
}
