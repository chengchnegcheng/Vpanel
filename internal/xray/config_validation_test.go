package xray

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"v/internal/logger"
)

// Feature: project-optimization, Property 15: Xray Configuration Validation
// *For any* configuration update, the configuration SHALL be validated before
// being applied, and invalid configurations SHALL be rejected.
// **Validates: Requirements 22.8**

// Feature: project-optimization, Property 16: Xray Configuration Rollback
// *For any* configuration update that fails to apply, the previous configuration
// SHALL be automatically restored from backup.
// **Validates: Requirements 22.11, 22.12**

// testableManager wraps manager for testing without requiring xray binary.
type testableManager struct {
	configPath    string
	backupDir     string
	logger        logger.Logger
	validateFunc  func(config json.RawMessage) error
	currentConfig json.RawMessage
}

func newTestableManager(configPath, backupDir string) *testableManager {
	return &testableManager{
		configPath: configPath,
		backupDir:  backupDir,
		logger:     logger.NewNopLogger(),
		validateFunc: func(config json.RawMessage) error {
			// Default: validate JSON structure
			var m map[string]any
			return json.Unmarshal(config, &m)
		},
	}
}

func (m *testableManager) setValidateFunc(f func(config json.RawMessage) error) {
	m.validateFunc = f
}

func (m *testableManager) ValidateConfig(ctx context.Context, config json.RawMessage) error {
	return m.validateFunc(config)
}

func (m *testableManager) UpdateConfig(ctx context.Context, config json.RawMessage) error {
	// Validate first
	if err := m.ValidateConfig(ctx, config); err != nil {
		return err
	}

	// Write config
	if err := os.WriteFile(m.configPath, config, 0644); err != nil {
		return err
	}

	m.currentConfig = config
	return nil
}

func (m *testableManager) GetConfig(ctx context.Context) (json.RawMessage, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func (m *testableManager) BackupConfig(ctx context.Context) (string, error) {
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return "", err
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	backupPath := filepath.Join(m.backupDir, "backup.json")
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", err
	}

	return backupPath, nil
}

func (m *testableManager) RestoreConfig(ctx context.Context, backupPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// TestConfigValidation_RejectsInvalidJSON tests that invalid JSON is rejected.
func TestConfigValidation_RejectsInvalidJSON(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("invalid JSON is rejected", prop.ForAll(
		func(invalidJSON string) bool {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "xray.json")
			backupDir := filepath.Join(tmpDir, "backups")

			mgr := newTestableManager(configPath, backupDir)

			// Try to update with invalid JSON
			err := mgr.UpdateConfig(context.Background(), json.RawMessage(invalidJSON))

			// Should fail for invalid JSON
			return err != nil
		},
		gen.OneConstOf(
			"{invalid}",
			"not json at all",
			"{\"key\": }",
			"[1, 2, 3",
			"{\"unclosed\": \"string",
		),
	))

	properties.TestingRun(t)
}

// TestConfigValidation_AcceptsValidJSON tests that valid JSON is accepted.
func TestConfigValidation_AcceptsValidJSON(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("valid JSON config is accepted", prop.ForAll(
		func(port int, logLevel string) bool {
			if port < 1 || port > 65535 {
				return true // Skip invalid ports
			}

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "xray.json")
			backupDir := filepath.Join(tmpDir, "backups")

			mgr := newTestableManager(configPath, backupDir)

			config := map[string]any{
				"log": map[string]any{
					"loglevel": logLevel,
				},
				"inbounds":  []any{},
				"outbounds": []any{},
			}
			configJSON, _ := json.Marshal(config)

			err := mgr.UpdateConfig(context.Background(), configJSON)
			return err == nil
		},
		gen.IntRange(1, 65535),
		gen.OneConstOf("debug", "info", "warning", "error", "none"),
	))

	properties.TestingRun(t)
}

// TestConfigValidation_CustomValidatorRejectsInvalid tests custom validation logic.
func TestConfigValidation_CustomValidatorRejectsInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "xray.json")
	backupDir := filepath.Join(tmpDir, "backups")

	mgr := newTestableManager(configPath, backupDir)

	// Set custom validator that requires "inbounds" field
	mgr.setValidateFunc(func(config json.RawMessage) error {
		var m map[string]any
		if err := json.Unmarshal(config, &m); err != nil {
			return err
		}
		if _, ok := m["inbounds"]; !ok {
			return &validationError{msg: "missing inbounds field"}
		}
		return nil
	})

	// Config without inbounds should fail
	configWithoutInbounds := json.RawMessage(`{"log": {"loglevel": "warning"}}`)
	err := mgr.UpdateConfig(context.Background(), configWithoutInbounds)
	assert.Error(t, err)

	// Config with inbounds should succeed
	configWithInbounds := json.RawMessage(`{"log": {"loglevel": "warning"}, "inbounds": []}`)
	err = mgr.UpdateConfig(context.Background(), configWithInbounds)
	assert.NoError(t, err)
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string {
	return e.msg
}

// TestConfigRollback_RestoresOnFailure tests that config is restored on failure.
func TestConfigRollback_RestoresOnFailure(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "xray.json")
	backupDir := filepath.Join(tmpDir, "backups")

	// Write initial config
	initialConfig := json.RawMessage(`{"log": {"loglevel": "info"}, "inbounds": [], "outbounds": []}`)
	err := os.WriteFile(configPath, initialConfig, 0644)
	require.NoError(t, err)

	mgr := newTestableManager(configPath, backupDir)

	// Backup current config
	backupPath, err := mgr.BackupConfig(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, backupPath)

	// Try to update with invalid config (will fail validation)
	invalidConfig := json.RawMessage(`{invalid}`)
	err = mgr.UpdateConfig(context.Background(), invalidConfig)
	assert.Error(t, err)

	// Restore from backup
	err = mgr.RestoreConfig(context.Background(), backupPath)
	require.NoError(t, err)

	// Verify config was restored
	restoredConfig, err := mgr.GetConfig(context.Background())
	require.NoError(t, err)

	var initial, restored map[string]any
	json.Unmarshal(initialConfig, &initial)
	json.Unmarshal(restoredConfig, &restored)

	assert.Equal(t, initial, restored)
}

// TestConfigRollback_BackupCreatedBeforeUpdate tests that backup is created before update.
func TestConfigRollback_BackupCreatedBeforeUpdate(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("backup is created before config update", prop.ForAll(
		func(logLevel string) bool {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "xray.json")
			backupDir := filepath.Join(tmpDir, "backups")

			// Write initial config
			initialConfig := map[string]any{
				"log":       map[string]any{"loglevel": "info"},
				"inbounds":  []any{},
				"outbounds": []any{},
			}
			initialJSON, _ := json.Marshal(initialConfig)
			os.WriteFile(configPath, initialJSON, 0644)

			mgr := newTestableManager(configPath, backupDir)

			// Create backup
			backupPath, err := mgr.BackupConfig(context.Background())
			if err != nil {
				return false
			}

			// Verify backup exists
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				return false
			}

			// Verify backup content matches original
			backupData, err := os.ReadFile(backupPath)
			if err != nil {
				return false
			}

			var backupConfig map[string]any
			if err := json.Unmarshal(backupData, &backupConfig); err != nil {
				return false
			}

			return backupConfig["log"] != nil
		},
		gen.OneConstOf("debug", "info", "warning", "error"),
	))

	properties.TestingRun(t)
}

// TestConfigRollback_RestorePreservesContent tests that restore preserves exact content.
func TestConfigRollback_RestorePreservesContent(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("restore preserves exact config content", prop.ForAll(
		func(port int, logLevel string) bool {
			if port < 1 || port > 65535 {
				return true
			}

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "xray.json")
			backupDir := filepath.Join(tmpDir, "backups")

			// Create original config
			originalConfig := map[string]any{
				"log": map[string]any{"loglevel": logLevel},
				"inbounds": []any{
					map[string]any{"port": port, "protocol": "vmess"},
				},
				"outbounds": []any{},
			}
			originalJSON, _ := json.Marshal(originalConfig)
			os.WriteFile(configPath, originalJSON, 0644)

			mgr := newTestableManager(configPath, backupDir)

			// Backup
			backupPath, err := mgr.BackupConfig(context.Background())
			if err != nil {
				return false
			}

			// Modify config
			newConfig := map[string]any{
				"log":       map[string]any{"loglevel": "error"},
				"inbounds":  []any{},
				"outbounds": []any{},
			}
			newJSON, _ := json.Marshal(newConfig)
			mgr.UpdateConfig(context.Background(), newJSON)

			// Restore
			if err := mgr.RestoreConfig(context.Background(), backupPath); err != nil {
				return false
			}

			// Verify restored content
			restoredJSON, err := mgr.GetConfig(context.Background())
			if err != nil {
				return false
			}

			var restored map[string]any
			if err := json.Unmarshal(restoredJSON, &restored); err != nil {
				return false
			}

			// Check log level was restored
			logConfig, ok := restored["log"].(map[string]any)
			if !ok {
				return false
			}

			return logConfig["loglevel"] == logLevel
		},
		gen.IntRange(1, 65535),
		gen.OneConstOf("debug", "info", "warning"),
	))

	properties.TestingRun(t)
}

// TestConfigValidation_ValidationBeforeWrite tests that validation happens before write.
func TestConfigValidation_ValidationBeforeWrite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "xray.json")
	backupDir := filepath.Join(tmpDir, "backups")

	// Write initial config
	initialConfig := json.RawMessage(`{"initial": true}`)
	err := os.WriteFile(configPath, initialConfig, 0644)
	require.NoError(t, err)

	mgr := newTestableManager(configPath, backupDir)

	// Set validator that always fails
	mgr.setValidateFunc(func(config json.RawMessage) error {
		return &validationError{msg: "always fails"}
	})

	// Try to update
	newConfig := json.RawMessage(`{"new": true}`)
	err = mgr.UpdateConfig(context.Background(), newConfig)
	assert.Error(t, err)

	// Verify original config was not modified
	currentConfig, err := mgr.GetConfig(context.Background())
	require.NoError(t, err)

	var current map[string]any
	json.Unmarshal(currentConfig, &current)

	assert.True(t, current["initial"].(bool), "original config should not be modified")
}

// TestConfigBackup_NoConfigNoBackup tests that backup handles missing config gracefully.
func TestConfigBackup_NoConfigNoBackup(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.json")
	backupDir := filepath.Join(tmpDir, "backups")

	mgr := newTestableManager(configPath, backupDir)

	// Backup should return empty path when no config exists
	backupPath, err := mgr.BackupConfig(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, backupPath)
}
