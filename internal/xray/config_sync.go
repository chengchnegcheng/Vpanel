// Package xray provides Xray-core process management.
package xray

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"v/internal/database/repository"
	"v/internal/logger"
	"v/internal/proxy"
	"v/pkg/errors"
)

// ConfigSyncer synchronizes proxy configurations with Xray.
type ConfigSyncer interface {
	// SyncAll regenerates the entire Xray configuration from all enabled proxies.
	SyncAll(ctx context.Context) error

	// SyncProxy adds or updates a single proxy in the Xray configuration.
	SyncProxy(ctx context.Context, p *repository.Proxy) error

	// RemoveProxy removes a proxy from the Xray configuration.
	RemoveProxy(ctx context.Context, proxyID int64) error

	// GetInbounds returns all inbound configurations.
	GetInbounds(ctx context.Context) ([]json.RawMessage, error)
}

// configSyncer implements ConfigSyncer.
type configSyncer struct {
	mu           sync.Mutex
	configPath   string
	proxyRepo    repository.ProxyRepository
	proxyManager proxy.Manager
	xrayManager  Manager
	logger       logger.Logger
}

// SyncConfig holds Xray configuration sync settings.
type SyncConfig struct {
	ConfigPath string
}

// NewConfigSyncer creates a new config syncer.
func NewConfigSyncer(
	cfg SyncConfig,
	proxyRepo repository.ProxyRepository,
	proxyManager proxy.Manager,
	xrayManager Manager,
	log logger.Logger,
) ConfigSyncer {
	return &configSyncer{
		configPath:   cfg.ConfigPath,
		proxyRepo:    proxyRepo,
		proxyManager: proxyManager,
		xrayManager:  xrayManager,
		logger:       log,
	}
}

// SyncAll regenerates the entire Xray configuration from all enabled proxies.
func (s *configSyncer) SyncAll(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get all enabled proxies
	proxies, err := s.proxyRepo.GetEnabled(ctx)
	if err != nil {
		return errors.NewDatabaseError("failed to get enabled proxies", err)
	}

	// Generate inbounds for all proxies
	inbounds := make([]json.RawMessage, 0, len(proxies))
	for _, p := range proxies {
		inbound, err := s.generateInbound(p)
		if err != nil {
			s.logger.Warn("failed to generate inbound for proxy",
				logger.F("proxy_id", p.ID),
				logger.F("error", err))
			continue
		}
		inbounds = append(inbounds, inbound)
	}

	// Build complete config
	config, err := s.buildConfig(inbounds)
	if err != nil {
		return err
	}

	// Write config
	if err := s.writeConfig(config); err != nil {
		return err
	}

	// Reload Xray if running
	if s.xrayManager != nil {
		if err := s.xrayManager.ReloadConfig(ctx); err != nil {
			s.logger.Warn("failed to reload xray config", logger.F("error", err))
		}
	}

	s.logger.Info("xray config synced", logger.F("proxy_count", len(inbounds)))
	return nil
}

// SyncProxy adds or updates a single proxy in the Xray configuration.
func (s *configSyncer) SyncProxy(ctx context.Context, p *repository.Proxy) error {
	// For simplicity, just sync all proxies
	// A more efficient implementation would update only the specific inbound
	return s.SyncAll(ctx)
}

// RemoveProxy removes a proxy from the Xray configuration.
func (s *configSyncer) RemoveProxy(ctx context.Context, proxyID int64) error {
	// For simplicity, just sync all proxies
	// A more efficient implementation would remove only the specific inbound
	return s.SyncAll(ctx)
}

// GetInbounds returns all inbound configurations.
func (s *configSyncer) GetInbounds(ctx context.Context) ([]json.RawMessage, error) {
	proxies, err := s.proxyRepo.GetEnabled(ctx)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get enabled proxies", err)
	}

	inbounds := make([]json.RawMessage, 0, len(proxies))
	for _, p := range proxies {
		inbound, err := s.generateInbound(p)
		if err != nil {
			continue
		}
		inbounds = append(inbounds, inbound)
	}

	return inbounds, nil
}

// generateInbound generates an inbound configuration for a proxy.
func (s *configSyncer) generateInbound(p *repository.Proxy) (json.RawMessage, error) {
	protocol, ok := s.proxyManager.GetProtocol(p.Protocol)
	if !ok {
		return nil, errors.NewValidationError("unsupported protocol", p.Protocol)
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

	return protocol.GenerateConfig(settings)
}

// buildConfig builds a complete Xray configuration.
func (s *configSyncer) buildConfig(inbounds []json.RawMessage) ([]byte, error) {
	config := map[string]any{
		"log": map[string]any{
			"loglevel": "warning",
		},
		"inbounds":  inbounds,
		"outbounds": s.getDefaultOutbounds(),
		"routing":   s.getDefaultRouting(),
	}

	return json.MarshalIndent(config, "", "  ")
}

// getDefaultOutbounds returns default outbound configurations.
func (s *configSyncer) getDefaultOutbounds() []map[string]any {
	return []map[string]any{
		{
			"tag":      "direct",
			"protocol": "freedom",
			"settings": map[string]any{},
		},
		{
			"tag":      "blocked",
			"protocol": "blackhole",
			"settings": map[string]any{},
		},
	}
}

// getDefaultRouting returns default routing configuration.
func (s *configSyncer) getDefaultRouting() map[string]any {
	return map[string]any{
		"domainStrategy": "AsIs",
		"rules": []map[string]any{
			{
				"type":        "field",
				"outboundTag": "direct",
				"network":     "tcp,udp",
			},
		},
	}
}

// writeConfig writes the configuration to file.
func (s *configSyncer) writeConfig(config []byte) error {
	if err := os.WriteFile(s.configPath, config, 0644); err != nil {
		return errors.NewInternalError("failed to write xray config", err)
	}
	return nil
}
