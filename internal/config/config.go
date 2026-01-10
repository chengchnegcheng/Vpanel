// Package config provides configuration management for the V Panel application.
// It supports loading configuration from YAML files and environment variables,
// with environment variables taking precedence over file values.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration.
type Config struct {
	Version  string         `yaml:"-"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Xray     XrayConfig     `yaml:"xray"`
	Log      LogConfig      `yaml:"log"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host            string        `yaml:"host" env:"V_SERVER_HOST" default:"0.0.0.0"`
	Port            int           `yaml:"port" env:"V_SERVER_PORT" default:"8080"`
	Mode            string        `yaml:"mode" env:"V_SERVER_MODE" default:"debug"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"V_SERVER_READ_TIMEOUT" default:"30s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"V_SERVER_WRITE_TIMEOUT" default:"30s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"V_SERVER_IDLE_TIMEOUT" default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"V_SERVER_SHUTDOWN_TIMEOUT" default:"30s"`
	TLSCert         string        `yaml:"tls_cert" env:"V_SERVER_TLS_CERT" default:""`
	TLSKey          string        `yaml:"tls_key" env:"V_SERVER_TLS_KEY" default:""`
	CORSOrigins     []string      `yaml:"cors_origins" env:"V_SERVER_CORS_ORIGINS"`
	StaticPath      string        `yaml:"static_path" env:"V_SERVER_STATIC_PATH" default:"./web/dist"`
}

// DatabaseConfig contains database connection settings.
type DatabaseConfig struct {
	Driver          string `yaml:"driver" env:"V_DB_DRIVER" default:"sqlite"`
	DSN             string `yaml:"dsn" env:"V_DB_DSN" default:"./data/v.db"`
	Path            string `yaml:"path" env:"V_DB_PATH" default:"./data/v.db"`
	MaxOpenConns    int    `yaml:"max_open_conns" env:"V_DB_MAX_OPEN_CONNS" default:"10"`
	MaxIdleConns    int    `yaml:"max_idle_conns" env:"V_DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime" env:"V_DB_CONN_MAX_LIFETIME" default:"1h"`
}

// AuthConfig contains authentication settings.
type AuthConfig struct {
	JWTSecret          string        `yaml:"jwt_secret" env:"V_JWT_SECRET" default:""`
	TokenExpiry        time.Duration `yaml:"token_expiry" env:"V_TOKEN_EXPIRY" default:"24h"`
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry" env:"V_REFRESH_TOKEN_EXPIRY" default:"168h"`
	AdminUsername      string        `yaml:"admin_username" env:"V_ADMIN_USER" default:"admin"`
	AdminPassword      string        `yaml:"admin_password" env:"V_ADMIN_PASS" default:"admin123"`
}

// XrayConfig contains Xray-core settings.
type XrayConfig struct {
	BinPath    string `yaml:"bin_path" env:"V_XRAY_BIN_PATH" default:"./xray/bin"`
	ConfigPath string `yaml:"config_path" env:"V_XRAY_CONFIG_PATH" default:"./xray/config.json"`
	Version    string `yaml:"version" env:"V_XRAY_VERSION" default:"latest"`
}

// LogConfig contains logging settings.
type LogConfig struct {
	Level  string `yaml:"level" env:"V_LOG_LEVEL" default:"info"`
	Format string `yaml:"format" env:"V_LOG_FORMAT" default:"json"`
	Output string `yaml:"output" env:"V_LOG_OUTPUT" default:"stdout"`
}

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation error: %s - %s", e.Field, e.Message)
}

// Loader handles configuration loading from various sources.
type Loader struct {
	configPath string
}

// NewLoader creates a new configuration loader.
func NewLoader(configPath string) *Loader {
	return &Loader{configPath: configPath}
}

// Load loads configuration from file and environment variables.
// Environment variables take precedence over file values.
func (l *Loader) Load() (*Config, error) {
	cfg := &Config{}

	// Apply defaults first
	if err := applyDefaults(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply defaults: %w", err)
	}

	// Load from file if exists
	if l.configPath != "" {
		if err := l.loadFromFile(cfg); err != nil {
			// File not found is not an error, we'll use defaults
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load config file: %w", err)
			}
		}
	}

	// Override with environment variables (highest precedence)
	if err := applyEnvOverrides(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply env overrides: %w", err)
	}

	return cfg, nil
}

// loadFromFile loads configuration from a YAML file.
func (l *Loader) loadFromFile(cfg *Config) error {
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return nil
}

// Validate validates the configuration and returns any errors.
func (cfg *Config) Validate() error {
	// Validate server config
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return &ValidationError{Field: "server.port", Message: "must be between 1 and 65535"}
	}

	// Validate database config
	if cfg.Database.Path == "" {
		return &ValidationError{Field: "database.path", Message: "must not be empty"}
	}

	// Validate auth config - JWT secret is required in production
	if cfg.Auth.JWTSecret == "" {
		// Generate a warning but don't fail - use a default for development
		cfg.Auth.JWTSecret = "development-secret-change-in-production"
	}

	if cfg.Auth.AdminUsername == "" {
		return &ValidationError{Field: "auth.admin_username", Message: "must not be empty"}
	}

	if cfg.Auth.AdminPassword == "" {
		return &ValidationError{Field: "auth.admin_password", Message: "must not be empty"}
	}

	// Validate log config
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true, "fatal": true}
	if !validLevels[strings.ToLower(cfg.Log.Level)] {
		return &ValidationError{Field: "log.level", Message: "must be one of: debug, info, warn, error, fatal"}
	}

	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[strings.ToLower(cfg.Log.Format)] {
		return &ValidationError{Field: "log.format", Message: "must be one of: json, text"}
	}

	return nil
}

// applyDefaults applies default values to the configuration.
func applyDefaults(cfg *Config) error {
	return applyDefaultsToStruct(reflect.ValueOf(cfg).Elem())
}

func applyDefaultsToStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Struct {
			if err := applyDefaultsToStruct(field); err != nil {
				return err
			}
			continue
		}

		defaultTag := fieldType.Tag.Get("default")
		if defaultTag == "" {
			continue
		}

		if err := setFieldValue(field, defaultTag); err != nil {
			return fmt.Errorf("failed to set default for %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// applyEnvOverrides applies environment variable overrides to the configuration.
func applyEnvOverrides(cfg *Config) error {
	return applyEnvOverridesToStruct(reflect.ValueOf(cfg).Elem())
}

func applyEnvOverridesToStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Struct {
			if err := applyEnvOverridesToStruct(field); err != nil {
				return err
			}
			continue
		}

		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}

		envValue := os.Getenv(envTag)
		if envValue == "" {
			continue
		}

		if err := setFieldValue(field, envValue); err != nil {
			return fmt.Errorf("failed to set env var %s: %w", envTag, err)
		}
	}

	return nil
}

// setFieldValue sets a field value from a string.
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(d))
		} else {
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(i)
		}
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}

// GetConfigPath returns the default configuration file path.
func GetConfigPath() string {
	// Check environment variable first
	if path := os.Getenv("V_CONFIG_PATH"); path != "" {
		return path
	}

	// Check common locations
	paths := []string{
		"./configs/config.yaml",
		"./config.yaml",
		"/app/configs/config.yaml",
		"/etc/v/config.yaml",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "./configs/config.yaml"
}

// EnsureDataDir ensures the data directory exists.
func EnsureDataDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}


// Load loads configuration from the specified path.
func Load(configPath string) (*Config, error) {
	loader := NewLoader(configPath)
	cfg, err := loader.Load()
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Ensure data directory exists
	if err := EnsureDataDir(cfg.Database.Path); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return cfg, nil
}
