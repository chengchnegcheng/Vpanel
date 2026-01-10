// Package repository provides data access implementations.
package repository

import (
	"context"

	"gorm.io/gorm"

	"v/pkg/errors"
)

// proxyRepository implements ProxyRepository.
type proxyRepository struct {
	db *gorm.DB
}

// NewProxyRepository creates a new proxy repository.
func NewProxyRepository(db *gorm.DB) ProxyRepository {
	return &proxyRepository{db: db}
}

// Create creates a new proxy.
func (r *proxyRepository) Create(ctx context.Context, proxy *Proxy) error {
	result := r.db.WithContext(ctx).Create(proxy)
	if result.Error != nil {
		return errors.NewDatabaseError("failed to create proxy", result.Error)
	}
	return nil
}

// GetByID retrieves a proxy by ID.
func (r *proxyRepository) GetByID(ctx context.Context, id int64) (*Proxy, error) {
	var proxy Proxy
	result := r.db.WithContext(ctx).First(&proxy, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("proxy", id)
		}
		return nil, errors.NewDatabaseError("failed to get proxy", result.Error)
	}
	return &proxy, nil
}

// Update updates a proxy.
func (r *proxyRepository) Update(ctx context.Context, proxy *Proxy) error {
	result := r.db.WithContext(ctx).Save(proxy)
	if result.Error != nil {
		return errors.NewDatabaseError("failed to update proxy", result.Error)
	}
	return nil
}

// Delete deletes a proxy by ID.
func (r *proxyRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&Proxy{}, id)
	if result.Error != nil {
		return errors.NewDatabaseError("failed to delete proxy", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundError("proxy", id)
	}
	return nil
}

// List retrieves proxies with pagination.
func (r *proxyRepository) List(ctx context.Context, limit, offset int) ([]*Proxy, error) {
	var proxies []*Proxy
	result := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&proxies)
	if result.Error != nil {
		return nil, errors.NewDatabaseError("failed to list proxies", result.Error)
	}
	return proxies, nil
}

// GetByProtocol retrieves proxies by protocol.
func (r *proxyRepository) GetByProtocol(ctx context.Context, protocol string) ([]*Proxy, error) {
	var proxies []*Proxy
	result := r.db.WithContext(ctx).Where("protocol = ?", protocol).Find(&proxies)
	if result.Error != nil {
		return nil, errors.NewDatabaseError("failed to get proxies by protocol", result.Error)
	}
	return proxies, nil
}

// GetEnabled retrieves all enabled proxies.
func (r *proxyRepository) GetEnabled(ctx context.Context) ([]*Proxy, error) {
	var proxies []*Proxy
	result := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&proxies)
	if result.Error != nil {
		return nil, errors.NewDatabaseError("failed to get enabled proxies", result.Error)
	}
	return proxies, nil
}
