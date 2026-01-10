package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 1: Configuration Precedence
// For any configuration key, when both an environment variable (with V_ prefix)
// and a config file value are set, the environment variable value SHALL take
// precedence over the config file value, and both SHALL take precedence over default values.
// **Validates: Requirements 1.6, 5.1, 5.3**

func TestConfigPrecedence_EnvOverridesFile(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("env vars override config file values for server port", prop.ForAll(
		func(filePort, envPort int) bool {
			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			configContent := `server:
  port: ` + string(rune(filePort+'0')) + `
`
			// Write file with actual port value
			configContent = "server:\n  port: " + itoa(filePort) + "\n"
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to write config file: %v", err)
				return false
			}

			// Set environment variable
			os.Setenv("V_SERVER_PORT", itoa(envPort))
			defer os.Unsetenv("V_SERVER_PORT")

			// Load config
			loader := NewLoader(configPath)
			cfg, err := loader.Load()
			if err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Environment variable should take precedence
			return cfg.Server.Port == envPort
		},
		gen.IntRange(1024, 65535),
		gen.IntRange(1024, 65535),
	))

	properties.Property("env vars override config file values for log level", prop.ForAll(
		func(fileLevel, envLevel string) bool {
			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			configContent := "log:\n  level: " + fileLevel + "\n"
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to write config file: %v", err)
				return false
			}

			// Set environment variable
			os.Setenv("V_LOG_LEVEL", envLevel)
			defer os.Unsetenv("V_LOG_LEVEL")

			// Load config
			loader := NewLoader(configPath)
			cfg, err := loader.Load()
			if err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Environment variable should take precedence
			return cfg.Log.Level == envLevel
		},
		gen.OneConstOf("debug", "info", "warn", "error"),
		gen.OneConstOf("debug", "info", "warn", "error"),
	))

	properties.TestingRun(t)
}

func TestConfigPrecedence_FileOverridesDefaults(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("config file values override defaults for server port", prop.ForAll(
		func(filePort int) bool {
			// Clear any environment variables
			os.Unsetenv("V_SERVER_PORT")

			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			configContent := "server:\n  port: " + itoa(filePort) + "\n"
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to write config file: %v", err)
				return false
			}

			// Load config
			loader := NewLoader(configPath)
			cfg, err := loader.Load()
			if err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// File value should override default (8080)
			return cfg.Server.Port == filePort
		},
		gen.IntRange(1024, 65535),
	))

	properties.TestingRun(t)
}

func TestConfigPrecedence_DefaultsApplied(t *testing.T) {
	// Clear all relevant environment variables
	envVars := []string{
		"V_SERVER_HOST", "V_SERVER_PORT",
		"V_DB_PATH", "V_LOG_LEVEL", "V_LOG_FORMAT",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	// Load config without file
	loader := NewLoader("")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check defaults are applied
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected default host 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Expected default log level info, got %s", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Expected default log format json, got %s", cfg.Log.Format)
	}
}

// Helper function to convert int to string
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}


// Property 4: Configuration Validation at Startup
// For any invalid configuration (missing required fields, invalid values),
// the Backend SHALL fail to start and log a specific error message identifying
// the invalid configuration.
// **Validates: Requirements 5.4**

func TestConfigValidation_InvalidPort(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("invalid port values are rejected", prop.ForAll(
		func(port int) bool {
			cfg := &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: port,
				},
				Database: DatabaseConfig{
					Path: "/data/v.db",
				},
				Auth: AuthConfig{
					JWTSecret:     "test-secret",
					AdminUsername: "admin",
					AdminPassword: "admin123",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "json",
				},
			}

			err := cfg.Validate()

			// Port must be between 1 and 65535
			if port < 1 || port > 65535 {
				// Should return validation error
				if err == nil {
					return false
				}
				validationErr, ok := err.(*ValidationError)
				if !ok {
					return false
				}
				return validationErr.Field == "server.port"
			}
			// Valid port should not cause validation error for this field
			return err == nil || err.(*ValidationError).Field != "server.port"
		},
		gen.OneGenOf(
			gen.IntRange(-1000, 0),      // Invalid: negative or zero
			gen.IntRange(65536, 100000), // Invalid: too high
			gen.IntRange(1, 65535),      // Valid range
		),
	))

	properties.TestingRun(t)
}

func TestConfigValidation_InvalidLogLevel(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	invalidLevels := []string{"trace", "verbose", "critical", "none", "invalid"}

	// Test valid levels pass validation
	for _, level := range validLevels {
		cfg := &Config{
			Server:   ServerConfig{Host: "0.0.0.0", Port: 8080},
			Database: DatabaseConfig{Path: "/data/v.db"},
			Auth:     AuthConfig{JWTSecret: "test", AdminUsername: "admin", AdminPassword: "admin123"},
			Log:      LogConfig{Level: level, Format: "json"},
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("Valid log level %q should not cause error, got: %v", level, err)
		}
	}

	// Test invalid levels fail validation
	for _, level := range invalidLevels {
		cfg := &Config{
			Server:   ServerConfig{Host: "0.0.0.0", Port: 8080},
			Database: DatabaseConfig{Path: "/data/v.db"},
			Auth:     AuthConfig{JWTSecret: "test", AdminUsername: "admin", AdminPassword: "admin123"},
			Log:      LogConfig{Level: level, Format: "json"},
		}
		err := cfg.Validate()
		if err == nil {
			t.Errorf("Invalid log level %q should cause error", level)
			continue
		}
		validationErr, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError for level %q, got %T", level, err)
			continue
		}
		if validationErr.Field != "log.level" {
			t.Errorf("Expected field 'log.level' for level %q, got %q", level, validationErr.Field)
		}
	}
}

func TestConfigValidation_InvalidLogFormat(t *testing.T) {
	validFormats := []string{"json", "text"}
	invalidFormats := []string{"xml", "csv", "yaml", "invalid"}

	// Test valid formats pass validation
	for _, format := range validFormats {
		cfg := &Config{
			Server:   ServerConfig{Host: "0.0.0.0", Port: 8080},
			Database: DatabaseConfig{Path: "/data/v.db"},
			Auth:     AuthConfig{JWTSecret: "test", AdminUsername: "admin", AdminPassword: "admin123"},
			Log:      LogConfig{Level: "info", Format: format},
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("Valid log format %q should not cause error, got: %v", format, err)
		}
	}

	// Test invalid formats fail validation
	for _, format := range invalidFormats {
		cfg := &Config{
			Server:   ServerConfig{Host: "0.0.0.0", Port: 8080},
			Database: DatabaseConfig{Path: "/data/v.db"},
			Auth:     AuthConfig{JWTSecret: "test", AdminUsername: "admin", AdminPassword: "admin123"},
			Log:      LogConfig{Level: "info", Format: format},
		}
		err := cfg.Validate()
		if err == nil {
			t.Errorf("Invalid log format %q should cause error", format)
			continue
		}
		validationErr, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError for format %q, got %T", format, err)
			continue
		}
		if validationErr.Field != "log.format" {
			t.Errorf("Expected field 'log.format' for format %q, got %q", format, validationErr.Field)
		}
	}
}

func TestConfigValidation_EmptyRequiredFields(t *testing.T) {
	// Test empty database path
	cfg := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Path: "",
		},
		Auth: AuthConfig{
			JWTSecret:     "test-secret",
			AdminUsername: "admin",
			AdminPassword: "admin123",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for empty database path")
	}
	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError, got %T", err)
	}
	if validationErr.Field != "database.path" {
		t.Errorf("Expected field 'database.path', got '%s'", validationErr.Field)
	}

	// Test empty admin username
	cfg.Database.Path = "/data/v.db"
	cfg.Auth.AdminUsername = ""
	err = cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for empty admin username")
	}

	// Test empty admin password
	cfg.Auth.AdminUsername = "admin"
	cfg.Auth.AdminPassword = ""
	err = cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for empty admin password")
	}
}
