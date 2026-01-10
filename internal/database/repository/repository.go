// Package repository provides data access interfaces and implementations.
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// User represents a user in the database.
type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"uniqueIndex;size:50;not null"`
	PasswordHash string    `gorm:"column:password;size:255;not null"`
	Email        string    `gorm:"size:100"`
	Role         string    `gorm:"size:20;default:user"`
	Enabled      bool      `gorm:"default:true"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TableName returns the table name for User.
func (User) TableName() string {
	return "users"
}

// Proxy represents a proxy configuration in the database.
type Proxy struct {
	ID        int64          `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"size:100;not null"`
	Protocol  string         `gorm:"size:20;not null"`
	Port      int            `gorm:"not null"`
	Host      string         `gorm:"size:255"`
	Settings  map[string]any `gorm:"serializer:json"`
	Enabled   bool           `gorm:"default:true"`
	Remark    string         `gorm:"size:255"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

// TableName returns the table name for Proxy.
func (Proxy) TableName() string {
	return "proxies"
}

// Traffic represents traffic statistics in the database.
type Traffic struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	UserID     int64     `gorm:"index"`
	ProxyID    int64     `gorm:"index"`
	Upload     int64     `gorm:"default:0"`
	Download   int64     `gorm:"default:0"`
	RecordedAt time.Time `gorm:"index"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// TableName returns the table name for Traffic.
func (Traffic) TableName() string {
	return "traffic"
}

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

// ProxyRepository defines the interface for proxy data access.
type ProxyRepository interface {
	Create(ctx context.Context, proxy *Proxy) error
	GetByID(ctx context.Context, id int64) (*Proxy, error)
	Update(ctx context.Context, proxy *Proxy) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*Proxy, error)
	GetByProtocol(ctx context.Context, protocol string) ([]*Proxy, error)
	GetEnabled(ctx context.Context) ([]*Proxy, error)
}

// TrafficRepository defines the interface for traffic data access.
type TrafficRepository interface {
	Create(ctx context.Context, traffic *Traffic) error
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*Traffic, error)
	GetByProxyID(ctx context.Context, proxyID int64, limit, offset int) ([]*Traffic, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*Traffic, error)
	GetTotalByUser(ctx context.Context, userID int64) (upload, download int64, err error)
}

// Repositories holds all repository instances.
type Repositories struct {
	User    UserRepository
	Proxy   ProxyRepository
	Traffic TrafficRepository
}

// NewRepositories creates all repository instances.
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:    NewUserRepository(db),
		Proxy:   NewProxyRepository(db),
		Traffic: NewTrafficRepository(db),
	}
}
