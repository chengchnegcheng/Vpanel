// Package proxy provides proxy protocol management for the V Panel application.
package proxy

import (
	"encoding/json"
)

// Protocol defines the interface for proxy protocols.
type Protocol interface {
	// Name returns the protocol name.
	Name() string

	// GenerateConfig generates Xray configuration for this protocol.
	GenerateConfig(settings *Settings) (json.RawMessage, error)

	// GenerateLink generates a share link for this protocol.
	GenerateLink(settings *Settings) (string, error)

	// ParseLink parses a share link and returns settings.
	ParseLink(link string) (*Settings, error)

	// Validate validates the protocol settings.
	Validate(settings *Settings) error

	// DefaultSettings returns default settings for this protocol.
	DefaultSettings() map[string]any
}

// Settings represents proxy configuration settings.
type Settings struct {
	ID       int64          `json:"id"`
	Name     string         `json:"name"`
	Protocol string         `json:"protocol"`
	Port     int            `json:"port"`
	Host     string         `json:"host,omitempty"`
	Settings map[string]any `json:"settings"`
	Enabled  bool           `json:"enabled"`
	Remark   string         `json:"remark,omitempty"`
}

// GetString gets a string value from settings.
func (s *Settings) GetString(key string) string {
	if s.Settings == nil {
		return ""
	}
	if v, ok := s.Settings[key]; ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt gets an int value from settings.
func (s *Settings) GetInt(key string) int {
	if s.Settings == nil {
		return 0
	}
	if v, ok := s.Settings[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

// GetBool gets a bool value from settings.
func (s *Settings) GetBool(key string) bool {
	if s.Settings == nil {
		return false
	}
	if v, ok := s.Settings[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// SetValue sets a value in settings.
func (s *Settings) SetValue(key string, value any) {
	if s.Settings == nil {
		s.Settings = make(map[string]any)
	}
	s.Settings[key] = value
}
