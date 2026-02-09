// Package agent provides the Node Agent functionality for V Panel.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"v/internal/logger"
)

// Agent represents the Node Agent that runs on each Xray node.
type Agent struct {
	config     *Config
	logger     logger.Logger
	httpClient *http.Client

	// Components
	xrayManager      *XrayManager
	healthServer     *HealthServer
	panelClient      *PanelClient
	metricsCollector *MetricsCollector
	commandExecutor  *CommandExecutor

	// State
	mu         sync.RWMutex
	running    bool
	registered bool
	nodeID     int64

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NodeMetrics represents metrics collected from the node.
type NodeMetrics struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	MemoryTotal  uint64  `json:"memory_total"`
	MemoryUsed   uint64  `json:"memory_used"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkIn    uint64  `json:"network_in"`
	NetworkOut   uint64  `json:"network_out"`
	Connections  int     `json:"connections"`
	XrayRunning  bool    `json:"xray_running"`
	XrayVersion  string  `json:"xray_version"`
	Uptime       int64   `json:"uptime"`
	Timestamp    int64   `json:"timestamp"`
}

// RegisterRequest represents a registration request to the Panel.
type RegisterRequest struct {
	Token   string `json:"token"`
	Name    string `json:"name"`
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// RegisterResponse represents a registration response from the Panel.
type RegisterResponse struct {
	Success bool   `json:"success"`
	NodeID  int64  `json:"node_id"`
	Message string `json:"message"`
}

// HeartbeatRequest represents a heartbeat request to the Panel.
type HeartbeatRequest struct {
	NodeID  int64        `json:"node_id"`
	Token   string       `json:"token"`
	Metrics *NodeMetrics `json:"metrics"`
}

// HeartbeatResponse represents a heartbeat response from the Panel.
type HeartbeatResponse struct {
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	Commands []Command `json:"commands,omitempty"`
}

// Command represents a command from the Panel to execute.
type Command struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// CommandResult represents the result of executing a command.
type CommandResult struct {
	CommandID string `json:"command_id"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
}

// New creates a new Node Agent.
func New(cfg *Config, log logger.Logger) (*Agent, error) {
	httpClient := &http.Client{
		Timeout: cfg.Panel.ConnectTimeout,
	}

	agent := &Agent{
		config:     cfg,
		logger:     log,
		httpClient: httpClient,
	}

	// Initialize Xray manager
	agent.xrayManager = NewXrayManager(XrayManagerConfig{
		BinaryPath: cfg.Xray.BinaryPath,
		ConfigPath: cfg.Xray.ConfigPath,
		BackupDir:  cfg.Xray.BackupDir,
	}, log)

	// Initialize health server
	agent.healthServer = NewHealthServer(HealthServerConfig{
		Host: cfg.Health.Host,
		Port: cfg.Health.Port,
	}, agent, log)

	// Initialize panel client
	agent.panelClient = NewPanelClient(PanelClientConfig{
		URL:               cfg.Panel.URL,
		Token:             cfg.Node.Token,
		TLSSkipVerify:     cfg.Panel.TLSSkipVerify,
		ConnectTimeout:    cfg.Panel.ConnectTimeout,
		ReconnectInterval: cfg.Panel.ReconnectInterval,
		MaxReconnectDelay: cfg.Panel.MaxReconnectDelay,
	}, log)

	// Initialize metrics collector
	agent.metricsCollector = NewMetricsCollector(log)

	// Initialize command executor
	agent.commandExecutor = NewCommandExecutor(agent, log)

	return agent, nil
}

// Start starts the Node Agent.
func (a *Agent) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("agent is already running")
	}
	a.running = true
	a.ctx, a.cancel = context.WithCancel(ctx)
	a.mu.Unlock()

	// Ensure Xray is installed
	installer := NewXrayInstaller(a.logger)
	if err := installer.EnsureXrayInstalled(ctx, a.config.Xray.ConfigPath); err != nil {
		a.logger.Error("failed to ensure xray installation", logger.F("error", err.Error()))
		// Continue anyway - Xray might be installed in a custom location
	}

	// Start Xray service if not running
	if err := a.ensureXrayRunning(ctx); err != nil {
		a.logger.Warn("failed to start xray service", logger.F("error", err.Error()))
		// Continue anyway - Xray might be managed externally
	}

	// Start health server
	if err := a.healthServer.Start(); err != nil {
		return fmt.Errorf("failed to start health server: %w", err)
	}

	// Register with Panel
	if err := a.register(); err != nil {
		a.logger.Warn("initial registration failed, will retry",
			logger.F("error", err.Error()))
	}

	// Start heartbeat loop
	a.wg.Add(1)
	go a.heartbeatLoop()

	// Start command processor
	a.wg.Add(1)
	go a.commandProcessorLoop()

	a.logger.Info("agent started successfully")
	return nil
}

// Stop stops the Node Agent.
func (a *Agent) Stop(ctx context.Context) error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil
	}
	a.cancel()
	a.running = false
	a.mu.Unlock()

	// Stop health server
	if err := a.healthServer.Stop(ctx); err != nil {
		a.logger.Error("failed to stop health server", logger.F("error", err))
	}

	// Wait for goroutines to finish
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("agent stopped")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// register registers the agent with the Panel Server.
func (a *Agent) register() error {
	req := &RegisterRequest{
		Token:   a.config.Node.Token,
		Name:    a.config.Node.Name,
		Version: "1.0.0",
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	resp, err := a.panelClient.Register(a.ctx, req)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration rejected: %s", resp.Message)
	}

	a.mu.Lock()
	a.nodeID = resp.NodeID
	a.registered = true
	a.mu.Unlock()

	a.logger.Info("registered with panel",
		logger.F("node_id", resp.NodeID),
		logger.F("message", resp.Message))

	return nil
}

// HeartbeatConfig holds configuration for heartbeat behavior.
type HeartbeatConfig struct {
	Interval       time.Duration
	RetryInterval  time.Duration
	MaxRetries     int
}

// DefaultHeartbeatConfig returns default heartbeat configuration.
func DefaultHeartbeatConfig() *HeartbeatConfig {
	return &HeartbeatConfig{
		Interval:      30 * time.Second,
		RetryInterval: 5 * time.Second,
		MaxRetries:    3,
	}
}

// heartbeatLoop sends periodic heartbeats to the Panel.
func (a *Agent) heartbeatLoop() {
	defer a.wg.Done()

	config := DefaultHeartbeatConfig()
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	// Send initial heartbeat
	a.sendHeartbeat()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a heartbeat to the Panel.
func (a *Agent) sendHeartbeat() {
	a.mu.RLock()
	if !a.registered {
		a.mu.RUnlock()
		// Try to register first with reconnection logic
		if a.panelClient.ShouldReconnect() {
			if err := a.panelClient.WaitForReconnect(a.ctx); err != nil {
				return // Context cancelled
			}
			if err := a.register(); err != nil {
				a.logger.Warn("registration failed during heartbeat",
					logger.F("error", err.Error()),
					logger.F("consecutive_fails", a.panelClient.GetConsecutiveFails()))
			}
		} else {
			a.logger.Error("max reconnection attempts reached, giving up",
				logger.F("consecutive_fails", a.panelClient.GetConsecutiveFails()))
		}
		return
	}
	nodeID := a.nodeID
	a.mu.RUnlock()

	// Collect metrics
	metrics := a.collectMetrics()

	req := &HeartbeatRequest{
		NodeID:  nodeID,
		Token:   a.config.Node.Token,
		Metrics: metrics,
	}

	resp, err := a.panelClient.Heartbeat(a.ctx, req)
	if err != nil {
		a.logger.Warn("heartbeat failed",
			logger.F("error", err.Error()),
			logger.F("node_id", nodeID),
			logger.F("consecutive_fails", a.panelClient.GetConsecutiveFails()))
		
		// Check if we need to re-register
		a.mu.Lock()
		a.registered = false
		a.mu.Unlock()
		return
	}

	if !resp.Success {
		a.logger.Warn("heartbeat rejected",
			logger.F("message", resp.Message),
			logger.F("node_id", nodeID))
		return
	}

	a.logger.Debug("heartbeat sent successfully",
		logger.F("node_id", nodeID),
		logger.F("commands_received", len(resp.Commands)))

	// Process any commands from the response
	if len(resp.Commands) > 0 {
		a.processCommands(resp.Commands)
	}
}

// collectMetrics collects current node metrics.
func (a *Agent) collectMetrics() *NodeMetrics {
	metrics := a.metricsCollector.Collect()

	// Add Xray status
	status := a.xrayManager.GetStatus()
	metrics.XrayRunning = status.Running
	metrics.XrayVersion = status.Version

	return metrics
}

// commandProcessorLoop processes commands from the Panel.
func (a *Agent) commandProcessorLoop() {
	defer a.wg.Done()

	// This loop handles any async command processing
	// Commands are primarily received via heartbeat responses
	<-a.ctx.Done()
}

// processCommands processes commands received from the Panel.
func (a *Agent) processCommands(commands []Command) {
	for _, cmd := range commands {
		result := a.commandExecutor.Execute(a.ctx, &cmd)
		
		// Report command result back to Panel
		if err := a.panelClient.ReportCommandResult(a.ctx, result); err != nil {
			a.logger.Error("failed to report command result",
				logger.F("command_id", cmd.ID),
				logger.F("error", err.Error()))
		}
	}
}

// executeCommand executes a single command (legacy method for backward compatibility).
func (a *Agent) executeCommand(cmd Command) *CommandResult {
	return a.commandExecutor.Execute(a.ctx, &cmd)
}

// IsRunning returns whether the agent is running.
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// IsRegistered returns whether the agent is registered with the Panel.
func (a *Agent) IsRegistered() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.registered
}

// GetNodeID returns the node ID assigned by the Panel.
func (a *Agent) GetNodeID() int64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.nodeID
}

// GetXrayStatus returns the current Xray status.
func (a *Agent) GetXrayStatus() *XrayStatus {
	return a.xrayManager.GetStatus()
}

// GetMetrics returns current node metrics.
func (a *Agent) GetMetrics() *NodeMetrics {
	return a.collectMetrics()
}

// GetXrayVersion returns the Xray version string.
func GetXrayVersion(binaryPath string) string {
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return string(output)
}

// ensureXrayRunning ensures Xray service is running.
func (a *Agent) ensureXrayRunning(ctx context.Context) error {
	// Check if Xray is running
	if a.isXrayRunning() {
		a.logger.Info("Xray service is already running")
		return nil
	}

	a.logger.Info("Starting Xray service...")

	// Try to start Xray using systemctl (Linux)
	if runtime.GOOS == "linux" {
		cmd := exec.CommandContext(ctx, "systemctl", "start", "xray")
		if err := cmd.Run(); err != nil {
			a.logger.Warn("Failed to start xray via systemctl", logger.F("error", err.Error()))
			// Try alternative method
			return a.startXrayDirect(ctx)
		}

		// Wait a moment for service to start
		time.Sleep(2 * time.Second)

		if a.isXrayRunning() {
			a.logger.Info("Xray service started successfully via systemctl")
			return nil
		}
	}

	// Fallback: start Xray directly
	return a.startXrayDirect(ctx)
}

// isXrayRunning checks if Xray process is running.
func (a *Agent) isXrayRunning() bool {
	// Try to check via systemctl first (Linux)
	if runtime.GOOS == "linux" {
		cmd := exec.Command("systemctl", "is-active", "xray")
		if err := cmd.Run(); err == nil {
			return true
		}
	}

	// Check if xray process exists
	cmd := exec.Command("pgrep", "-x", "xray")
	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

// startXrayDirect starts Xray process directly.
func (a *Agent) startXrayDirect(ctx context.Context) error {
	a.logger.Info("Starting Xray directly...")

	// Start Xray in background
	cmd := exec.Command("xray", "run", "-c", a.config.Xray.ConfigPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start xray: %w", err)
	}

	// Detach from process
	go func() {
		if err := cmd.Wait(); err != nil {
			a.logger.Error("Xray process exited", logger.F("error", err.Error()))
		}
	}()

	// Wait a moment for process to start
	time.Sleep(2 * time.Second)

	if a.isXrayRunning() {
		a.logger.Info("Xray started successfully")
		return nil
	}

	return fmt.Errorf("xray failed to start")
}
