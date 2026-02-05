// Package node provides node management functionality for multi-server management.
package node

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"v/internal/database/repository"
	"v/internal/logger"
	"v/internal/notification"
)

// HealthCheckConfig holds configuration for the health checker.
type HealthCheckConfig struct {
	// Interval is the time between health checks (default: 30 seconds)
	Interval time.Duration
	// Timeout is the timeout for each health check (default: 10 seconds)
	Timeout time.Duration
	// UnhealthyThreshold is the number of consecutive failures before marking unhealthy
	UnhealthyThreshold int
	// HealthyThreshold is the number of consecutive successes before marking healthy
	HealthyThreshold int
	// RetentionDays is how many days to keep health check history
	RetentionDays int
	// MaxConcurrentChecks is the maximum number of concurrent health checks (default: 10)
	MaxConcurrentChecks int
}

// DefaultHealthCheckConfig returns the default health check configuration.
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		Interval:            30 * time.Second,
		Timeout:             10 * time.Second,
		UnhealthyThreshold:  3,
		HealthyThreshold:    2,
		RetentionDays:       7,
		MaxConcurrentChecks: 10,
	}
}

// HealthCheckResult represents the result of a health check.
type HealthCheckResult struct {
	NodeID    int64
	Status    string // healthy, unhealthy
	TCPOk     bool
	APIOk     bool
	XrayOk    bool
	Latency   int // milliseconds
	Message   string
	CheckedAt time.Time
}

// HealthChecker performs periodic health checks on nodes.
type HealthChecker struct {
	config          *HealthCheckConfig
	nodeRepo        repository.NodeRepository
	healthCheckRepo repository.HealthCheckRepository
	logger          logger.Logger
	httpClient      *http.Client
	notificationSvc *notification.Service

	// State tracking for consecutive failures/successes
	consecutiveFailures map[int64]int
	consecutiveSuccesses map[int64]int
	stateMu             sync.RWMutex

	// Control
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    bool
	runningMu  sync.Mutex

	// Notification callback
	onStatusChange func(nodeID int64, oldStatus, newStatus string)
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(
	config *HealthCheckConfig,
	nodeRepo repository.NodeRepository,
	healthCheckRepo repository.HealthCheckRepository,
	log logger.Logger,
) *HealthChecker {
	if config == nil {
		config = DefaultHealthCheckConfig()
	}

	return &HealthChecker{
		config:               config,
		nodeRepo:             nodeRepo,
		healthCheckRepo:      healthCheckRepo,
		logger:               log,
		httpClient:           &http.Client{Timeout: config.Timeout},
		consecutiveFailures:  make(map[int64]int),
		consecutiveSuccesses: make(map[int64]int),
	}
}

// SetOnStatusChange sets the callback for node status changes.
func (hc *HealthChecker) SetOnStatusChange(callback func(nodeID int64, oldStatus, newStatus string)) {
	hc.onStatusChange = callback
}

// SetNotificationService sets the notification service for sending alerts.
func (hc *HealthChecker) SetNotificationService(svc *notification.Service) {
	hc.notificationSvc = svc
}

// Start starts the health checker.
func (hc *HealthChecker) Start(ctx context.Context) error {
	hc.runningMu.Lock()
	defer hc.runningMu.Unlock()

	if hc.running {
		return fmt.Errorf("health checker is already running")
	}

	hc.ctx, hc.cancel = context.WithCancel(ctx)
	hc.running = true

	hc.wg.Add(1)
	go hc.runLoop()

	hc.logger.Info("Health checker started",
		logger.F("interval", hc.config.Interval.String()),
		logger.F("unhealthy_threshold", hc.config.UnhealthyThreshold),
		logger.F("healthy_threshold", hc.config.HealthyThreshold))

	return nil
}

// Stop stops the health checker.
func (hc *HealthChecker) Stop(ctx context.Context) error {
	hc.runningMu.Lock()
	if !hc.running {
		hc.runningMu.Unlock()
		return nil
	}
	hc.cancel()
	hc.running = false
	hc.runningMu.Unlock()

	// Wait for goroutine to finish with timeout
	done := make(chan struct{})
	go func() {
		hc.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		hc.logger.Info("Health checker stopped")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsRunning returns whether the health checker is running.
func (hc *HealthChecker) IsRunning() bool {
	hc.runningMu.Lock()
	defer hc.runningMu.Unlock()
	return hc.running
}

// runLoop is the main loop that performs periodic health checks.
func (hc *HealthChecker) runLoop() {
	defer hc.wg.Done()

	ticker := time.NewTicker(hc.config.Interval)
	defer ticker.Stop()

	// Run initial check immediately
	hc.checkAllNodes()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.checkAllNodes()
		}
	}
}

// checkAllNodes performs health checks on all registered nodes.
// Uses a worker pool to limit concurrent checks and prevent resource exhaustion.
func (hc *HealthChecker) checkAllNodes() {
	nodes, err := hc.nodeRepo.List(hc.ctx, nil)
	if err != nil {
		hc.logger.Error("Failed to list nodes for health check", logger.Err(err))
		return
	}

	if len(nodes) == 0 {
		return
	}

	// Create a semaphore to limit concurrent checks
	maxConcurrent := hc.config.MaxConcurrentChecks
	if maxConcurrent <= 0 {
		maxConcurrent = 10 // Default fallback
	}
	
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	hc.logger.Debug("Starting health checks",
		logger.F("node_count", len(nodes)),
		logger.F("max_concurrent", maxConcurrent))

	for _, node := range nodes {
		wg.Add(1)
		
		// Acquire semaphore
		select {
		case sem <- struct{}{}:
			// Got a slot, proceed
		case <-hc.ctx.Done():
			// Context cancelled, stop spawning new checks
			wg.Done()
			continue
		}

		go func(n *repository.Node) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore
			
			// Check for context cancellation
			select {
			case <-hc.ctx.Done():
				return
			default:
				hc.checkNode(n)
			}
		}(node)
	}

	wg.Wait()

	// Cleanup old health check records
	hc.cleanupOldRecords()
}

// checkNode performs a health check on a single node.
func (hc *HealthChecker) checkNode(node *repository.Node) {
	result := hc.performCheck(node)

	// Save health check record
	check := &repository.HealthCheck{
		NodeID:    node.ID,
		Status:    result.Status,
		Latency:   result.Latency,
		Message:   result.Message,
		TCPOk:     result.TCPOk,
		APIOk:     result.APIOk,
		XrayOk:    result.XrayOk,
		CheckedAt: result.CheckedAt,
	}

	if err := hc.healthCheckRepo.Create(hc.ctx, check); err != nil {
		hc.logger.Error("Failed to save health check record",
			logger.Err(err),
			logger.F("node_id", node.ID))
	}

	// Update consecutive counters and handle status transitions
	hc.handleCheckResult(node, result)
}

// performCheck performs the actual health check on a node.
func (hc *HealthChecker) performCheck(node *repository.Node) *HealthCheckResult {
	result := &HealthCheckResult{
		NodeID:    node.ID,
		CheckedAt: time.Now(),
	}

	start := time.Now()

	// Check TCP connectivity
	result.TCPOk = hc.checkTCP(node.Address, node.Port)

	// Check API responsiveness
	if result.TCPOk {
		result.APIOk = hc.checkAPI(node.Address, node.Port)
	}

	// Check Xray status (via API)
	if result.APIOk {
		result.XrayOk = hc.checkXray(node.Address, node.Port)
	}

	result.Latency = int(time.Since(start).Milliseconds())

	// Determine overall status
	if result.TCPOk && result.APIOk && result.XrayOk {
		result.Status = repository.HealthCheckStatusSuccess
		result.Message = "All checks passed"
	} else {
		result.Status = repository.HealthCheckStatusFailed
		result.Message = hc.buildFailureMessage(result)
	}

	return result
}

// checkTCP checks TCP connectivity to the node.
func (hc *HealthChecker) checkTCP(address string, port int) bool {
	addr := fmt.Sprintf("%s:%d", address, port)
	conn, err := net.DialTimeout("tcp", addr, hc.config.Timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// checkAPI checks the node agent API responsiveness.
func (hc *HealthChecker) checkAPI(address string, port int) bool {
	url := fmt.Sprintf("http://%s:%d/health", address, port)
	resp, err := hc.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// checkXray checks the Xray process status via the node agent.
func (hc *HealthChecker) checkXray(address string, port int) bool {
	url := fmt.Sprintf("http://%s:%d/xray/status", address, port)
	resp, err := hc.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// buildFailureMessage builds a failure message based on check results.
func (hc *HealthChecker) buildFailureMessage(result *HealthCheckResult) string {
	var failures []string
	if !result.TCPOk {
		failures = append(failures, "TCP connection failed")
	}
	if !result.APIOk {
		failures = append(failures, "API not responding")
	}
	if !result.XrayOk {
		failures = append(failures, "Xray not running")
	}
	if len(failures) == 0 {
		return "Unknown failure"
	}
	return fmt.Sprintf("Health check failed: %v", failures)
}

// handleCheckResult handles the result of a health check and updates node status.
func (hc *HealthChecker) handleCheckResult(node *repository.Node, result *HealthCheckResult) {
	hc.stateMu.Lock()
	defer hc.stateMu.Unlock()

	oldStatus := node.Status

	if result.Status == repository.HealthCheckStatusSuccess {
		// Reset failure counter, increment success counter
		hc.consecutiveFailures[node.ID] = 0
		hc.consecutiveSuccesses[node.ID]++

		// Check if we should transition to healthy
		if node.Status == repository.NodeStatusUnhealthy &&
			hc.consecutiveSuccesses[node.ID] >= hc.config.HealthyThreshold {
			hc.transitionToHealthy(node, oldStatus)
		} else if node.Status == repository.NodeStatusOffline {
			// Node came online
			hc.transitionToOnline(node, oldStatus)
		}

		// Update latency
		if err := hc.nodeRepo.UpdateMetrics(hc.ctx, node.ID, result.Latency, node.CurrentUsers); err != nil {
			hc.logger.Error("Failed to update node metrics",
				logger.Err(err),
				logger.F("node_id", node.ID))
		}

		// Update last seen
		if err := hc.nodeRepo.UpdateLastSeen(hc.ctx, node.ID, result.CheckedAt); err != nil {
			hc.logger.Error("Failed to update node last seen",
				logger.Err(err),
				logger.F("node_id", node.ID))
		}
	} else {
		// Reset success counter, increment failure counter
		hc.consecutiveSuccesses[node.ID] = 0
		hc.consecutiveFailures[node.ID]++

		// Check if we should transition to unhealthy
		if node.Status == repository.NodeStatusOnline &&
			hc.consecutiveFailures[node.ID] >= hc.config.UnhealthyThreshold {
			hc.transitionToUnhealthy(node, oldStatus, result.Message)
		}
	}
}

// transitionToHealthy transitions a node to healthy (online) status.
func (hc *HealthChecker) transitionToHealthy(node *repository.Node, oldStatus string) {
	if err := hc.nodeRepo.UpdateStatus(hc.ctx, node.ID, repository.NodeStatusOnline); err != nil {
		hc.logger.Error("Failed to update node status to online",
			logger.Err(err),
			logger.F("node_id", node.ID))
		return
	}

	hc.logger.Info("Node recovered to healthy",
		logger.F("node_id", node.ID),
		logger.F("node_name", node.Name),
		logger.F("old_status", oldStatus))

	// Send notification
	hc.sendStatusChangeNotification(node, oldStatus, repository.NodeStatusOnline, "Node recovered after consecutive successful health checks")

	// Trigger notification callback
	if hc.onStatusChange != nil {
		hc.onStatusChange(node.ID, oldStatus, repository.NodeStatusOnline)
	}
}

// transitionToOnline transitions a node to online status.
func (hc *HealthChecker) transitionToOnline(node *repository.Node, oldStatus string) {
	if err := hc.nodeRepo.UpdateStatus(hc.ctx, node.ID, repository.NodeStatusOnline); err != nil {
		hc.logger.Error("Failed to update node status to online",
			logger.Err(err),
			logger.F("node_id", node.ID))
		return
	}

	hc.logger.Info("Node came online",
		logger.F("node_id", node.ID),
		logger.F("node_name", node.Name),
		logger.F("old_status", oldStatus))

	// Send notification
	hc.sendStatusChangeNotification(node, oldStatus, repository.NodeStatusOnline, "Node came online")

	// Trigger notification callback
	if hc.onStatusChange != nil {
		hc.onStatusChange(node.ID, oldStatus, repository.NodeStatusOnline)
	}
}

// transitionToUnhealthy transitions a node to unhealthy status.
func (hc *HealthChecker) transitionToUnhealthy(node *repository.Node, oldStatus string, reason string) {
	if err := hc.nodeRepo.UpdateStatus(hc.ctx, node.ID, repository.NodeStatusUnhealthy); err != nil {
		hc.logger.Error("Failed to update node status to unhealthy",
			logger.Err(err),
			logger.F("node_id", node.ID))
		return
	}

	hc.logger.Warn("Node became unhealthy",
		logger.F("node_id", node.ID),
		logger.F("node_name", node.Name),
		logger.F("old_status", oldStatus),
		logger.F("reason", reason),
		logger.F("consecutive_failures", hc.consecutiveFailures[node.ID]))

	// Send notification
	hc.sendStatusChangeNotification(node, oldStatus, repository.NodeStatusUnhealthy, reason)

	// Trigger notification callback
	if hc.onStatusChange != nil {
		hc.onStatusChange(node.ID, oldStatus, repository.NodeStatusUnhealthy)
	}
}

// cleanupOldRecords removes health check records older than retention period.
func (hc *HealthChecker) cleanupOldRecords() {
	cutoff := time.Now().AddDate(0, 0, -hc.config.RetentionDays)
	deleted, err := hc.healthCheckRepo.DeleteOlderThan(hc.ctx, cutoff)
	if err != nil {
		hc.logger.Error("Failed to cleanup old health check records", logger.Err(err))
		return
	}
	if deleted > 0 {
		hc.logger.Debug("Cleaned up old health check records",
			logger.F("deleted", deleted),
			logger.F("cutoff", cutoff))
	}
}

// CheckNode performs a manual health check on a specific node.
func (hc *HealthChecker) CheckNode(ctx context.Context, nodeID int64) (*HealthCheckResult, error) {
	node, err := hc.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("node not found: %w", err)
	}

	result := hc.performCheck(node)

	// Save health check record
	check := &repository.HealthCheck{
		NodeID:    node.ID,
		Status:    result.Status,
		Latency:   result.Latency,
		Message:   result.Message,
		TCPOk:     result.TCPOk,
		APIOk:     result.APIOk,
		XrayOk:    result.XrayOk,
		CheckedAt: result.CheckedAt,
	}

	if err := hc.healthCheckRepo.Create(ctx, check); err != nil {
		hc.logger.Error("Failed to save health check record",
			logger.Err(err),
			logger.F("node_id", node.ID))
	}

	return result, nil
}

// CheckAll performs manual health checks on all nodes.
// Uses a worker pool to limit concurrent checks.
func (hc *HealthChecker) CheckAll(ctx context.Context) ([]*HealthCheckResult, error) {
	nodes, err := hc.nodeRepo.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes) == 0 {
		return []*HealthCheckResult{}, nil
	}

	results := make([]*HealthCheckResult, 0, len(nodes))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create a semaphore to limit concurrent checks
	maxConcurrent := hc.config.MaxConcurrentChecks
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}
	sem := make(chan struct{}, maxConcurrent)

	for _, node := range nodes {
		wg.Add(1)
		
		// Acquire semaphore
		select {
		case sem <- struct{}{}:
			// Got a slot
		case <-ctx.Done():
			wg.Done()
			continue
		}

		go func(n *repository.Node) {
			defer wg.Done()
			defer func() { <-sem }()
			
			// Check for context cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			result := hc.performCheck(n)

			// Save health check record
			check := &repository.HealthCheck{
				NodeID:    n.ID,
				Status:    result.Status,
				Latency:   result.Latency,
				Message:   result.Message,
				TCPOk:     result.TCPOk,
				APIOk:     result.APIOk,
				XrayOk:    result.XrayOk,
				CheckedAt: result.CheckedAt,
			}

			if err := hc.healthCheckRepo.Create(ctx, check); err != nil {
				hc.logger.Error("Failed to save health check record",
					logger.Err(err),
					logger.F("node_id", n.ID))
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(node)
	}

	wg.Wait()
	return results, nil
}

// GetHistory retrieves health check history for a node.
func (hc *HealthChecker) GetHistory(ctx context.Context, nodeID int64, limit int) ([]*repository.HealthCheck, error) {
	return hc.healthCheckRepo.GetByNodeID(ctx, nodeID, limit)
}

// GetConsecutiveFailures returns the current consecutive failure count for a node.
func (hc *HealthChecker) GetConsecutiveFailures(nodeID int64) int {
	hc.stateMu.RLock()
	defer hc.stateMu.RUnlock()
	return hc.consecutiveFailures[nodeID]
}

// GetConsecutiveSuccesses returns the current consecutive success count for a node.
func (hc *HealthChecker) GetConsecutiveSuccesses(nodeID int64) int {
	hc.stateMu.RLock()
	defer hc.stateMu.RUnlock()
	return hc.consecutiveSuccesses[nodeID]
}

// UpdateConfig updates the health checker configuration.
func (hc *HealthChecker) UpdateConfig(config *HealthCheckConfig) {
	hc.runningMu.Lock()
	defer hc.runningMu.Unlock()
	hc.config = config
	hc.httpClient.Timeout = config.Timeout
}

// GetConfig returns the current health checker configuration.
func (hc *HealthChecker) GetConfig() *HealthCheckConfig {
	hc.runningMu.Lock()
	defer hc.runningMu.Unlock()
	return hc.config
}

// sendStatusChangeNotification sends a notification when node status changes.
func (hc *HealthChecker) sendStatusChangeNotification(node *repository.Node, oldStatus, newStatus, reason string) {
	if hc.notificationSvc == nil {
		return
	}

	data := notification.NodeStatusChangeData{
		NodeID:    node.ID,
		NodeName:  node.Name,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Reason:    reason,
		Timestamp: time.Now(),
	}

	if err := hc.notificationSvc.NotifyNodeStatusChange(data); err != nil {
		hc.logger.Error("Failed to send node status change notification",
			logger.Err(err),
			logger.F("node_id", node.ID),
			logger.F("old_status", oldStatus),
			logger.F("new_status", newStatus))
	}
}
