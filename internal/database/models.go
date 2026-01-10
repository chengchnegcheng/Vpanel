// Package database provides data models for the V Panel application.
package database

import (
	"time"

	"gorm.io/datatypes"
)

// User represents a user in the system.
type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password     string    `json:"-" gorm:"size:255;not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;size:100"`
	Role         string    `json:"role" gorm:"size:20;default:user"`
	IsAdmin      bool      `json:"is_admin" gorm:"default:false"`
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	TrafficLimit int64     `json:"traffic_limit" gorm:"default:0"` // 0 means unlimited
	TrafficUsed  int64     `json:"traffic_used" gorm:"default:0"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName returns the table name for User.
func (User) TableName() string {
	return "users"
}

// Proxy represents a proxy configuration.
type Proxy struct {
	ID        int64          `json:"id" gorm:"primaryKey"`
	UserID    int64          `json:"user_id" gorm:"index;not null"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Protocol  string         `json:"protocol" gorm:"size:20;not null"` // vmess, vless, trojan, shadowsocks
	Port      int            `json:"port" gorm:"index;not null"`
	Settings  datatypes.JSON `json:"settings"` // Protocol-specific settings
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	Remark    string         `json:"remark" gorm:"size:255"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// TableName returns the table name for Proxy.
func (Proxy) TableName() string {
	return "proxies"
}

// Traffic represents traffic statistics.
type Traffic struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	UserID     int64     `json:"user_id" gorm:"index;not null"`
	ProxyID    int64     `json:"proxy_id" gorm:"index"`
	Upload     int64     `json:"upload" gorm:"default:0"`
	Download   int64     `json:"download" gorm:"default:0"`
	RecordedAt time.Time `json:"recorded_at" gorm:"index;not null"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName returns the table name for Traffic.
func (Traffic) TableName() string {
	return "traffic"
}

// Certificate represents an SSL certificate.
type Certificate struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Domain      string    `json:"domain" gorm:"uniqueIndex;size:255;not null"`
	Certificate string    `json:"certificate" gorm:"type:text"`
	PrivateKey  string    `json:"-" gorm:"type:text"`
	AutoRenew   bool      `json:"auto_renew" gorm:"default:true"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName returns the table name for Certificate.
func (Certificate) TableName() string {
	return "certificates"
}

// Setting represents a system setting.
type Setting struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex;size:100;not null"`
	Value     string    `json:"value" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for Setting.
func (Setting) TableName() string {
	return "settings"
}

// Log represents a system log entry.
type Log struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Level     string    `json:"level" gorm:"size:10;index"`
	Message   string    `json:"message" gorm:"type:text"`
	Source    string    `json:"source" gorm:"size:50;index"`
	UserID    int64     `json:"user_id" gorm:"index"`
	IP        string    `json:"ip" gorm:"size:50"`
	UserAgent string    `json:"user_agent" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

// TableName returns the table name for Log.
func (Log) TableName() string {
	return "logs"
}
