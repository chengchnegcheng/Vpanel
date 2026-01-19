// Package xray provides Xray configuration generation and management.
package xray

import (
	"context"
	"encoding/json"
	"fmt"

	"v/internal/database/repository"
	"v/internal/logger"
)

// ConfigGenerator generates Xray configurations for nodes.
type ConfigGenerator struct {
	proxyRepo repository.ProxyRepository
	logger    logger.Logger
}

// NewConfigGenerator creates a new Xray config generator.
func NewConfigGenerator(
	proxyRepo repository.ProxyRepository,
	log logger.Logger,
) *ConfigGenerator {
	return &ConfigGenerator{
		proxyRepo: proxyRepo,
		logger:    log,
	}
}

// XrayConfig represents the complete Xray configuration.
type XrayConfig struct {
	Log       *LogConfig       `json:"log"`
	API       *APIConfig       `json:"api,omitempty"`
	Stats     *StatsConfig     `json:"stats,omitempty"`
	Policy    *PolicyConfig    `json:"policy,omitempty"`
	Inbounds  []InboundConfig  `json:"inbounds"`
	Outbounds []OutboundConfig `json:"outbounds"`
	Routing   *RoutingConfig   `json:"routing,omitempty"`
}

// LogConfig represents Xray log configuration.
type LogConfig struct {
	LogLevel string `json:"loglevel"`
	Access   string `json:"access"`
	Error    string `json:"error"`
}

// APIConfig represents Xray API configuration.
type APIConfig struct {
	Tag      string   `json:"tag"`
	Services []string `json:"services"`
}

// StatsConfig represents Xray stats configuration.
type StatsConfig struct{}

// PolicyConfig represents Xray policy configuration.
type PolicyConfig struct {
	Levels map[string]*PolicyLevel `json:"levels,omitempty"`
	System *SystemPolicy           `json:"system,omitempty"`
}

// PolicyLevel represents policy for a specific level.
type PolicyLevel struct {
	StatsUserUplink   bool `json:"statsUserUplink"`
	StatsUserDownlink bool `json:"statsUserDownlink"`
}

// SystemPolicy represents system-wide policy.
type SystemPolicy struct {
	StatsInboundUplink    bool `json:"statsInboundUplink"`
	StatsInboundDownlink  bool `json:"statsInboundDownlink"`
	StatsOutboundUplink   bool `json:"statsOutboundUplink"`
	StatsOutboundDownlink bool `json:"statsOutboundDownlink"`
}

// InboundConfig represents an Xray inbound configuration.
type InboundConfig struct {
	Tag          string         `json:"tag"`
	Listen       string         `json:"listen,omitempty"`
	Port         int            `json:"port"`
	Protocol     string         `json:"protocol"`
	Settings     map[string]any `json:"settings"`
	StreamSettings *StreamSettings `json:"streamSettings,omitempty"`
	Sniffing     *SniffingConfig `json:"sniffing,omitempty"`
}

// StreamSettings represents stream settings for transport.
type StreamSettings struct {
	Network  string         `json:"network"`
	Security string         `json:"security,omitempty"`
	TLSSettings *TLSSettings `json:"tlsSettings,omitempty"`
	TCPSettings map[string]any `json:"tcpSettings,omitempty"`
	WSSettings  map[string]any `json:"wsSettings,omitempty"`
	HTTPSettings map[string]any `json:"httpSettings,omitempty"`
	QUICSettings map[string]any `json:"quicSettings,omitempty"`
	GRPCSettings map[string]any `json:"grpcSettings,omitempty"`
}

// TLSSettings represents TLS configuration.
type TLSSettings struct {
	ServerName   string   `json:"serverName,omitempty"`
	Certificates []Certificate `json:"certificates,omitempty"`
	ALPN         []string `json:"alpn,omitempty"`
}

// Certificate represents a TLS certificate.
type Certificate struct {
	CertificateFile string `json:"certificateFile,omitempty"`
	KeyFile         string `json:"keyFile,omitempty"`
	Certificate     []string `json:"certificate,omitempty"`
	Key             []string `json:"key,omitempty"`
}

// SniffingConfig represents sniffing configuration.
type SniffingConfig struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
}

// OutboundConfig represents an Xray outbound configuration.
type OutboundConfig struct {
	Tag      string         `json:"tag"`
	Protocol string         `json:"protocol"`
	Settings map[string]any `json:"settings"`
}

// RoutingConfig represents Xray routing configuration.
type RoutingConfig struct {
	Rules []RoutingRule `json:"rules"`
}

// RoutingRule represents a routing rule.
type RoutingRule struct {
	Type        string   `json:"type"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	OutboundTag string   `json:"outboundTag"`
	Protocol    []string `json:"protocol,omitempty"`
}

// GenerateForNode generates Xray configuration for a specific node.
func (g *ConfigGenerator) GenerateForNode(ctx context.Context, nodeID int64) (*XrayConfig, error) {
	// Get all enabled proxies for users assigned to this node
	allProxies, err := g.proxyRepo.GetByNodeID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proxies for node: %w", err)
	}

	g.logger.Info("generating config for node",
		logger.F("node_id", nodeID),
		logger.F("proxy_count", len(allProxies)))

	// Generate configuration
	config := &XrayConfig{
		Log: &LogConfig{
			LogLevel: "warning",
			Access:   "",
			Error:    "",
		},
		API: &APIConfig{
			Tag:      "api",
			Services: []string{"HandlerService", "LoggerService", "StatsService"},
		},
		Stats: &StatsConfig{},
		Policy: &PolicyConfig{
			Levels: map[string]*PolicyLevel{
				"0": {
					StatsUserUplink:   true,
					StatsUserDownlink: true,
				},
			},
			System: &SystemPolicy{
				StatsInboundUplink:    true,
				StatsInboundDownlink:  true,
				StatsOutboundUplink:   true,
				StatsOutboundDownlink: true,
			},
		},
		Inbounds:  g.generateInbounds(allProxies),
		Outbounds: g.generateOutbounds(),
		Routing:   g.generateRouting(),
	}

	return config, nil
}

// generateInbounds generates inbound configurations from proxies.
func (g *ConfigGenerator) generateInbounds(proxies []*repository.Proxy) []InboundConfig {
	inbounds := []InboundConfig{
		// API inbound for stats
		{
			Tag:      "api",
			Listen:   "127.0.0.1",
			Port:     62789,
			Protocol: "dokodemo-door",
			Settings: map[string]any{
				"address": "127.0.0.1",
			},
		},
	}

	// Generate inbound for each proxy
	for _, proxy := range proxies {
		inbound := g.proxyToInbound(proxy)
		if inbound != nil {
			inbounds = append(inbounds, *inbound)
		}
	}

	return inbounds
}

// proxyToInbound converts a proxy to an Xray inbound configuration.
func (g *ConfigGenerator) proxyToInbound(proxy *repository.Proxy) *InboundConfig {
	tag := fmt.Sprintf("inbound-%d", proxy.ID)

	inbound := &InboundConfig{
		Tag:      tag,
		Port:     proxy.Port,
		Protocol: proxy.Protocol,
		Settings: make(map[string]any),
		Sniffing: &SniffingConfig{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		},
	}

	// Extract settings from proxy
	settings := proxy.Settings
	if settings == nil {
		settings = make(map[string]any)
	}

	switch proxy.Protocol {
	case "vless":
		inbound.Settings = g.generateVLESSSettings(proxy, settings)
		inbound.StreamSettings = g.generateStreamSettings(settings)
	case "vmess":
		inbound.Settings = g.generateVMessSettings(proxy, settings)
		inbound.StreamSettings = g.generateStreamSettings(settings)
	case "trojan":
		inbound.Settings = g.generateTrojanSettings(proxy, settings)
		inbound.StreamSettings = g.generateStreamSettings(settings)
	case "shadowsocks":
		inbound.Settings = g.generateShadowsocksSettings(proxy, settings)
	default:
		g.logger.Warn("Unsupported protocol",
			logger.F("proxy_id", proxy.ID),
			logger.F("protocol", proxy.Protocol))
		return nil
	}

	return inbound
}

// generateVLESSSettings generates VLESS protocol settings.
func (g *ConfigGenerator) generateVLESSSettings(proxy *repository.Proxy, settings map[string]any) map[string]any {
	clients := []map[string]any{}
	
	// Extract UUID from settings
	if uuid, ok := settings["uuid"].(string); ok && uuid != "" {
		clients = append(clients, map[string]any{
			"id":    uuid,
			"email": fmt.Sprintf("user-%d-proxy-%d", proxy.UserID, proxy.ID),
			"level": 0,
		})
	}

	return map[string]any{
		"clients":    clients,
		"decryption": "none",
		"fallbacks":  []map[string]any{},
	}
}

// generateVMessSettings generates VMess protocol settings.
func (g *ConfigGenerator) generateVMessSettings(proxy *repository.Proxy, settings map[string]any) map[string]any {
	clients := []map[string]any{}
	
	// Extract UUID from settings
	if uuid, ok := settings["uuid"].(string); ok && uuid != "" {
		client := map[string]any{
			"id":    uuid,
			"email": fmt.Sprintf("user-%d-proxy-%d", proxy.UserID, proxy.ID),
			"level": 0,
		}
		
		// Optional: alterId
		if alterId, ok := settings["alter_id"]; ok {
			client["alterId"] = alterId
		} else {
			client["alterId"] = 0
		}
		
		clients = append(clients, client)
	}

	return map[string]any{
		"clients": clients,
	}
}

// generateTrojanSettings generates Trojan protocol settings.
func (g *ConfigGenerator) generateTrojanSettings(proxy *repository.Proxy, settings map[string]any) map[string]any {
	clients := []map[string]any{}
	
	// Extract password from settings
	if password, ok := settings["password"].(string); ok && password != "" {
		clients = append(clients, map[string]any{
			"password": password,
			"email":    fmt.Sprintf("user-%d-proxy-%d", proxy.UserID, proxy.ID),
			"level":    0,
		})
	}

	return map[string]any{
		"clients":   clients,
		"fallbacks": []map[string]any{},
	}
}

// generateShadowsocksSettings generates Shadowsocks protocol settings.
func (g *ConfigGenerator) generateShadowsocksSettings(proxy *repository.Proxy, settings map[string]any) map[string]any {
	result := map[string]any{
		"network": "tcp,udp",
	}
	
	// Extract method and password
	if method, ok := settings["method"].(string); ok {
		result["method"] = method
	} else {
		result["method"] = "aes-256-gcm"
	}
	
	if password, ok := settings["password"].(string); ok {
		result["password"] = password
	}

	return result
}

// generateStreamSettings generates stream settings from proxy settings.
func (g *ConfigGenerator) generateStreamSettings(settings map[string]any) *StreamSettings {
	stream := &StreamSettings{
		Network: "tcp", // default
	}

	// Extract network type
	if network, ok := settings["network"].(string); ok {
		stream.Network = network
	}

	// Extract security settings
	if security, ok := settings["security"].(string); ok {
		stream.Security = security
		
		if security == "tls" {
			stream.TLSSettings = &TLSSettings{}
			
			if serverName, ok := settings["server_name"].(string); ok {
				stream.TLSSettings.ServerName = serverName
			}
			
			// Certificate settings
			if certFile, ok := settings["cert_file"].(string); ok {
				if keyFile, ok := settings["key_file"].(string); ok {
					stream.TLSSettings.Certificates = []Certificate{
						{
							CertificateFile: certFile,
							KeyFile:         keyFile,
						},
					}
				}
			}
			
			// ALPN
			if alpn, ok := settings["alpn"].([]string); ok {
				stream.TLSSettings.ALPN = alpn
			}
		}
	}

	// Network-specific settings
	switch stream.Network {
	case "ws":
		if wsSettings, ok := settings["ws_settings"].(map[string]any); ok {
			stream.WSSettings = wsSettings
		}
	case "tcp":
		if tcpSettings, ok := settings["tcp_settings"].(map[string]any); ok {
			stream.TCPSettings = tcpSettings
		}
	case "http":
		if httpSettings, ok := settings["http_settings"].(map[string]any); ok {
			stream.HTTPSettings = httpSettings
		}
	case "quic":
		if quicSettings, ok := settings["quic_settings"].(map[string]any); ok {
			stream.QUICSettings = quicSettings
		}
	case "grpc":
		if grpcSettings, ok := settings["grpc_settings"].(map[string]any); ok {
			stream.GRPCSettings = grpcSettings
		}
	}

	return stream
}

// generateOutbounds generates outbound configurations.
func (g *ConfigGenerator) generateOutbounds() []OutboundConfig {
	return []OutboundConfig{
		{
			Tag:      "direct",
			Protocol: "freedom",
			Settings: map[string]any{},
		},
		{
			Tag:      "blocked",
			Protocol: "blackhole",
			Settings: map[string]any{},
		},
	}
}

// generateRouting generates routing configuration.
func (g *ConfigGenerator) generateRouting() *RoutingConfig {
	return &RoutingConfig{
		Rules: []RoutingRule{
			{
				Type:        "field",
				InboundTag:  []string{"api"},
				OutboundTag: "api",
			},
			{
				Type:        "field",
				Protocol:    []string{"bittorrent"},
				OutboundTag: "blocked",
			},
		},
	}
}

// ToJSON converts the configuration to JSON.
func (c *XrayConfig) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}
