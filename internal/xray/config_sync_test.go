package xray

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"v/internal/database/repository"
	"v/internal/logger"
	"v/internal/proxy"
)

// Feature: project-optimization, Property 14: Xray Configuration Sync
// *For any* proxy created, updated, or deleted, the Xray configuration SHALL be
// synchronized to reflect the current state of all enabled proxies.
// **Validates: Requirements 21.6, 21.7, 21.13**

// mockProxyRepoForSync is a mock implementation of ProxyRepository for config sync testing.
type mockProxyRepoForSync struct {
	proxies map[int64]*repository.Proxy
	nextID  int64
}

func newMockProxyRepoForSync() *mockProxyRepoForSync {
	return &mockProxyRepoForSync{
		proxies: make(map[int64]*repository.Proxy),
		nextID:  1,
	}
}

func (m *mockProxyRepoForSync) Create(ctx context.Context, p *repository.Proxy) error {
	p.ID = m.nextID
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.nextID++
	m.proxies[p.ID] = p
	return nil
}

func (m *mockProxyRepoForSync) GetByID(ctx context.Context, id int64) (*repository.Proxy, error) {
	if p, ok := m.proxies[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockProxyRepoForSync) Update(ctx context.Context, p *repository.Proxy) error {
	p.UpdatedAt = time.Now()
	m.proxies[p.ID] = p
	return nil
}

func (m *mockProxyRepoForSync) Delete(ctx context.Context, id int64) error {
	delete(m.proxies, id)
	return nil
}

func (m *mockProxyRepoForSync) List(ctx context.Context, limit, offset int) ([]*repository.Proxy, error) {
	var result []*repository.Proxy
	for _, p := range m.proxies {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockProxyRepoForSync) GetByProtocol(ctx context.Context, protocol string) ([]*repository.Proxy, error) {
	var result []*repository.Proxy
	for _, p := range m.proxies {
		if p.Protocol == protocol {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockProxyRepoForSync) GetEnabled(ctx context.Context) ([]*repository.Proxy, error) {
	var result []*repository.Proxy
	for _, p := range m.proxies {
		if p.Enabled {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockProxyRepoForSync) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*repository.Proxy, error) {
	var result []*repository.Proxy
	for _, p := range m.proxies {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockProxyRepoForSync) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	for _, p := range m.proxies {
		if p.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockProxyRepoForSync) GetByPort(ctx context.Context, port int) (*repository.Proxy, error) {
	for _, p := range m.proxies {
		if p.Port == port {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockProxyRepoForSync) EnableByUserID(ctx context.Context, userID int64) error {
	for _, p := range m.proxies {
		if p.UserID == userID {
			p.Enabled = true
		}
	}
	return nil
}

func (m *mockProxyRepoForSync) DisableByUserID(ctx context.Context, userID int64) error {
	for _, p := range m.proxies {
		if p.UserID == userID {
			p.Enabled = false
		}
	}
	return nil
}

func (m *mockProxyRepoForSync) DeleteByIDs(ctx context.Context, ids []int64) error {
	for _, id := range ids {
		delete(m.proxies, id)
	}
	return nil
}

func (m *mockProxyRepoForSync) Count(ctx context.Context) (int64, error) {
	return int64(len(m.proxies)), nil
}

func (m *mockProxyRepoForSync) CountEnabled(ctx context.Context) (int64, error) {
	var count int64
	for _, p := range m.proxies {
		if p.Enabled {
			count++
		}
	}
	return count, nil
}

func (m *mockProxyRepoForSync) CountByProtocol(ctx context.Context) ([]*repository.ProtocolCount, error) {
	counts := make(map[string]int64)
	for _, p := range m.proxies {
		counts[p.Protocol]++
	}
	var result []*repository.ProtocolCount
	for protocol, count := range counts {
		result = append(result, &repository.ProtocolCount{Protocol: protocol, Count: count})
	}
	return result, nil
}

func (m *mockProxyRepoForSync) GetByNodeID(ctx context.Context, nodeID int64) ([]*repository.Proxy, error) {
	var result []*repository.Proxy
	for _, p := range m.proxies {
		if p.NodeID != nil && *p.NodeID == nodeID && p.Enabled {
			result = append(result, p)
		}
	}
	return result, nil
}

// mockProxyManagerForSync is a mock implementation of proxy.Manager for config sync testing.
type mockProxyManagerForSync struct {
	protocols map[string]proxy.Protocol
}

func newMockProxyManagerForSync() *mockProxyManagerForSync {
	mgr := &mockProxyManagerForSync{
		protocols: make(map[string]proxy.Protocol),
	}
	// Register mock protocols
	mgr.protocols["vmess"] = &mockProtocolForSync{name: "vmess"}
	mgr.protocols["vless"] = &mockProtocolForSync{name: "vless"}
	mgr.protocols["trojan"] = &mockProtocolForSync{name: "trojan"}
	return mgr
}

func (m *mockProxyManagerForSync) RegisterProtocol(protocol proxy.Protocol) {
	m.protocols[protocol.Name()] = protocol
}

func (m *mockProxyManagerForSync) GetProtocol(name string) (proxy.Protocol, bool) {
	p, ok := m.protocols[name]
	return p, ok
}

func (m *mockProxyManagerForSync) ListProtocols() []string {
	var names []string
	for name := range m.protocols {
		names = append(names, name)
	}
	return names
}

func (m *mockProxyManagerForSync) CreateProxy(ctx context.Context, settings *proxy.Settings) error {
	return nil
}

func (m *mockProxyManagerForSync) UpdateProxy(ctx context.Context, settings *proxy.Settings) error {
	return nil
}

func (m *mockProxyManagerForSync) DeleteProxy(ctx context.Context, id int64) error {
	return nil
}

func (m *mockProxyManagerForSync) GetProxy(ctx context.Context, id int64) (*proxy.Settings, error) {
	return nil, nil
}

func (m *mockProxyManagerForSync) ListProxies(ctx context.Context, page, pageSize int) ([]*proxy.Settings, int64, error) {
	return nil, 0, nil
}

func (m *mockProxyManagerForSync) GetProxiesByUser(ctx context.Context, userID int64) ([]*proxy.Settings, error) {
	return nil, nil
}

func (m *mockProxyManagerForSync) GenerateLink(ctx context.Context, id int64) (string, error) {
	return "", nil
}

func (m *mockProxyManagerForSync) GenerateConfig(ctx context.Context, id int64) (json.RawMessage, error) {
	return nil, nil
}

// mockProtocolForSync is a mock implementation of proxy.Protocol for config sync testing.
type mockProtocolForSync struct {
	name string
}

func (m *mockProtocolForSync) Name() string {
	return m.name
}

func (m *mockProtocolForSync) Validate(settings *proxy.Settings) error {
	return nil
}

func (m *mockProtocolForSync) DefaultSettings() map[string]any {
	return map[string]any{}
}

func (m *mockProtocolForSync) GenerateLink(settings *proxy.Settings) (string, error) {
	return "mock://link", nil
}

func (m *mockProtocolForSync) GenerateConfig(settings *proxy.Settings) (json.RawMessage, error) {
	config := map[string]any{
		"tag":      settings.Name,
		"protocol": m.name,
		"port":     settings.Port,
		"listen":   "0.0.0.0",
		"settings": map[string]any{},
	}
	return json.Marshal(config)
}

func (m *mockProtocolForSync) ParseLink(link string) (*proxy.Settings, error) {
	return &proxy.Settings{}, nil
}

// mockXrayManagerForSync is a mock implementation of xray.Manager for config sync testing.
type mockXrayManagerForSync struct {
	reloadCalled bool
}

func (m *mockXrayManagerForSync) Start(ctx context.Context) error {
	return nil
}

func (m *mockXrayManagerForSync) Stop(ctx context.Context) error {
	return nil
}

func (m *mockXrayManagerForSync) Restart(ctx context.Context) error {
	return nil
}

func (m *mockXrayManagerForSync) GetStatus(ctx context.Context) (*Status, error) {
	return &Status{Running: true}, nil
}

func (m *mockXrayManagerForSync) GetConfig(ctx context.Context) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}

func (m *mockXrayManagerForSync) UpdateConfig(ctx context.Context, config json.RawMessage) error {
	return nil
}

func (m *mockXrayManagerForSync) ValidateConfig(ctx context.Context, config json.RawMessage) error {
	return nil
}

func (m *mockXrayManagerForSync) ReloadConfig(ctx context.Context) error {
	m.reloadCalled = true
	return nil
}

func (m *mockXrayManagerForSync) GetVersion(ctx context.Context) (*Version, error) {
	return &Version{Current: "1.8.0"}, nil
}

func (m *mockXrayManagerForSync) BackupConfig(ctx context.Context) (string, error) {
	return "/tmp/backup.json", nil
}

func (m *mockXrayManagerForSync) RestoreConfig(ctx context.Context, backupPath string) error {
	return nil
}

// createTestConfigSyncer creates a ConfigSyncer for testing with a temp config file.
func createTestConfigSyncer(t *testing.T) (ConfigSyncer, *mockProxyRepoForSync, *mockXrayManagerForSync, string) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "xray.json")

	repo := newMockProxyRepoForSync()
	proxyMgr := newMockProxyManagerForSync()
	xrayMgr := &mockXrayManagerForSync{}

	syncer := NewConfigSyncer(
		SyncConfig{ConfigPath: configPath},
		repo,
		proxyMgr,
		xrayMgr,
		logger.NewNopLogger(),
	)

	return syncer, repo, xrayMgr, configPath
}

// TestConfigSync_SyncAllGeneratesValidConfig tests that SyncAll generates a valid Xray config.
func TestConfigSync_SyncAllGeneratesValidConfig(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("SyncAll generates valid JSON config with all enabled proxies", prop.ForAll(
		func(numEnabled int, numDisabled int) bool {
			if numEnabled < 0 || numEnabled > 10 || numDisabled < 0 || numDisabled > 10 {
				return true // Skip extreme cases
			}

			syncer, repo, _, configPath := createTestConfigSyncer(t)

			// Create enabled proxies
			for i := 0; i < numEnabled; i++ {
				repo.Create(context.Background(), &repository.Proxy{
					Name:     "enabled-proxy",
					Protocol: "vmess",
					Port:     10000 + i,
					Enabled:  true,
				})
			}

			// Create disabled proxies
			for i := 0; i < numDisabled; i++ {
				repo.Create(context.Background(), &repository.Proxy{
					Name:     "disabled-proxy",
					Protocol: "vmess",
					Port:     20000 + i,
					Enabled:  false,
				})
			}

			// Sync all
			err := syncer.SyncAll(context.Background())
			if err != nil {
				return false
			}

			// Read and parse the config file
			data, err := os.ReadFile(configPath)
			if err != nil {
				return false
			}

			var config map[string]any
			if err := json.Unmarshal(data, &config); err != nil {
				return false
			}

			// Verify config structure
			if _, ok := config["log"]; !ok {
				return false
			}
			if _, ok := config["outbounds"]; !ok {
				return false
			}
			if _, ok := config["routing"]; !ok {
				return false
			}

			// Verify inbounds count matches enabled proxies
			inbounds, ok := config["inbounds"].([]any)
			if !ok {
				return false
			}

			return len(inbounds) == numEnabled
		},
		gen.IntRange(0, 5),
		gen.IntRange(0, 5),
	))

	properties.TestingRun(t)
}

// TestConfigSync_SyncProxyTriggersReload tests that SyncProxy triggers Xray reload.
func TestConfigSync_SyncProxyTriggersReload(t *testing.T) {
	syncer, repo, xrayMgr, _ := createTestConfigSyncer(t)

	// Create a proxy
	p := &repository.Proxy{
		Name:     "test-proxy",
		Protocol: "vmess",
		Port:     10000,
		Enabled:  true,
	}
	repo.Create(context.Background(), p)

	// Sync the proxy
	err := syncer.SyncProxy(context.Background(), p)
	require.NoError(t, err)

	// Verify reload was called
	assert.True(t, xrayMgr.reloadCalled, "ReloadConfig should be called after SyncProxy")
}

// TestConfigSync_RemoveProxyUpdatesConfig tests that RemoveProxy updates the config.
func TestConfigSync_RemoveProxyUpdatesConfig(t *testing.T) {
	syncer, repo, _, configPath := createTestConfigSyncer(t)

	// Create two proxies
	p1 := &repository.Proxy{
		Name:     "proxy-1",
		Protocol: "vmess",
		Port:     10000,
		Enabled:  true,
	}
	p2 := &repository.Proxy{
		Name:     "proxy-2",
		Protocol: "vmess",
		Port:     10001,
		Enabled:  true,
	}
	repo.Create(context.Background(), p1)
	repo.Create(context.Background(), p2)

	// Sync all
	err := syncer.SyncAll(context.Background())
	require.NoError(t, err)

	// Verify both proxies are in config
	data, _ := os.ReadFile(configPath)
	var config map[string]any
	json.Unmarshal(data, &config)
	inbounds := config["inbounds"].([]any)
	assert.Equal(t, 2, len(inbounds))

	// Delete one proxy from repo
	repo.Delete(context.Background(), p1.ID)

	// Remove proxy from config
	err = syncer.RemoveProxy(context.Background(), p1.ID)
	require.NoError(t, err)

	// Verify only one proxy remains in config
	data, _ = os.ReadFile(configPath)
	json.Unmarshal(data, &config)
	inbounds = config["inbounds"].([]any)
	assert.Equal(t, 1, len(inbounds))
}

// TestConfigSync_OnlyEnabledProxiesInConfig tests that only enabled proxies appear in config.
func TestConfigSync_OnlyEnabledProxiesInConfig(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("only enabled proxies appear in generated config", prop.ForAll(
		func(enabledPorts []int, disabledPorts []int) bool {
			if len(enabledPorts) > 10 || len(disabledPorts) > 10 {
				return true // Skip extreme cases
			}

			syncer, repo, _, configPath := createTestConfigSyncer(t)

			// Create enabled proxies
			for _, port := range enabledPorts {
				if port < 1 || port > 65535 {
					continue
				}
				repo.Create(context.Background(), &repository.Proxy{
					Name:     "enabled",
					Protocol: "vmess",
					Port:     port,
					Enabled:  true,
				})
			}

			// Create disabled proxies
			for _, port := range disabledPorts {
				if port < 1 || port > 65535 {
					continue
				}
				repo.Create(context.Background(), &repository.Proxy{
					Name:     "disabled",
					Protocol: "vmess",
					Port:     port,
					Enabled:  false,
				})
			}

			// Sync all
			err := syncer.SyncAll(context.Background())
			if err != nil {
				return false
			}

			// Read config
			data, err := os.ReadFile(configPath)
			if err != nil {
				return false
			}

			var config map[string]any
			if err := json.Unmarshal(data, &config); err != nil {
				return false
			}

			inbounds, ok := config["inbounds"].([]any)
			if !ok {
				return false
			}

			// Count valid enabled ports
			validEnabledCount := 0
			for _, port := range enabledPorts {
				if port >= 1 && port <= 65535 {
					validEnabledCount++
				}
			}

			return len(inbounds) == validEnabledCount
		},
		gen.SliceOf(gen.IntRange(1, 65535)),
		gen.SliceOf(gen.IntRange(1, 65535)),
	))

	properties.TestingRun(t)
}

// TestConfigSync_GetInboundsReturnsEnabledOnly tests that GetInbounds returns only enabled proxies.
func TestConfigSync_GetInboundsReturnsEnabledOnly(t *testing.T) {
	syncer, repo, _, _ := createTestConfigSyncer(t)

	// Create mixed proxies
	repo.Create(context.Background(), &repository.Proxy{
		Name:     "enabled-1",
		Protocol: "vmess",
		Port:     10000,
		Enabled:  true,
	})
	repo.Create(context.Background(), &repository.Proxy{
		Name:     "disabled-1",
		Protocol: "vmess",
		Port:     10001,
		Enabled:  false,
	})
	repo.Create(context.Background(), &repository.Proxy{
		Name:     "enabled-2",
		Protocol: "vless",
		Port:     10002,
		Enabled:  true,
	})

	// Get inbounds
	inbounds, err := syncer.GetInbounds(context.Background())
	require.NoError(t, err)

	// Should only have 2 enabled proxies
	assert.Equal(t, 2, len(inbounds))
}

// TestConfigSync_ConfigHasRequiredSections tests that generated config has all required sections.
func TestConfigSync_ConfigHasRequiredSections(t *testing.T) {
	syncer, repo, _, configPath := createTestConfigSyncer(t)

	// Create a proxy
	repo.Create(context.Background(), &repository.Proxy{
		Name:     "test-proxy",
		Protocol: "vmess",
		Port:     10000,
		Enabled:  true,
	})

	// Sync
	err := syncer.SyncAll(context.Background())
	require.NoError(t, err)

	// Read config
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var config map[string]any
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	// Verify required sections
	assert.Contains(t, config, "log", "config should have log section")
	assert.Contains(t, config, "inbounds", "config should have inbounds section")
	assert.Contains(t, config, "outbounds", "config should have outbounds section")
	assert.Contains(t, config, "routing", "config should have routing section")

	// Verify outbounds has direct and blocked
	outbounds, ok := config["outbounds"].([]any)
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(outbounds), 2, "should have at least direct and blocked outbounds")
}

// TestConfigSync_UnsupportedProtocolSkipped tests that proxies with unsupported protocols are skipped.
func TestConfigSync_UnsupportedProtocolSkipped(t *testing.T) {
	syncer, repo, _, configPath := createTestConfigSyncer(t)

	// Create proxy with supported protocol
	repo.Create(context.Background(), &repository.Proxy{
		Name:     "supported",
		Protocol: "vmess",
		Port:     10000,
		Enabled:  true,
	})

	// Create proxy with unsupported protocol
	repo.Create(context.Background(), &repository.Proxy{
		Name:     "unsupported",
		Protocol: "unknown-protocol",
		Port:     10001,
		Enabled:  true,
	})

	// Sync all
	err := syncer.SyncAll(context.Background())
	require.NoError(t, err)

	// Read config
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var config map[string]any
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	// Should only have 1 inbound (the supported one)
	inbounds := config["inbounds"].([]any)
	assert.Equal(t, 1, len(inbounds))
}

// TestConfigSync_EmptyProxiesGeneratesValidConfig tests that empty proxy list generates valid config.
func TestConfigSync_EmptyProxiesGeneratesValidConfig(t *testing.T) {
	syncer, _, _, configPath := createTestConfigSyncer(t)

	// Sync with no proxies
	err := syncer.SyncAll(context.Background())
	require.NoError(t, err)

	// Read config
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var config map[string]any
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	// Should have empty inbounds
	inbounds := config["inbounds"].([]any)
	assert.Equal(t, 0, len(inbounds))

	// But still have outbounds and routing
	assert.Contains(t, config, "outbounds")
	assert.Contains(t, config, "routing")
}
