// Package announcement provides announcement services for the user portal.
package announcement

import (
	"context"

	"v/internal/database/repository"
)

// Service provides announcement operations for the user portal.
type Service struct {
	announcementRepo repository.AnnouncementRepository
}

// NewService creates a new announcement service.
func NewService(announcementRepo repository.AnnouncementRepository) *Service {
	return &Service{
		announcementRepo: announcementRepo,
	}
}

// AnnouncementResult represents an announcement with read status.
type AnnouncementResult struct {
	*repository.Announcement
	IsRead bool `json:"is_read"`
}

// ListAnnouncements retrieves published announcements for a user with read status.
func (s *Service) ListAnnouncements(ctx context.Context, userID int64, limit, offset int) ([]*AnnouncementResult, int64, error) {
	announcements, total, err := s.announcementRepo.ListWithReadStatus(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	results := make([]*AnnouncementResult, len(announcements))
	for i, a := range announcements {
		results[i] = &AnnouncementResult{
			Announcement: &a.Announcement,
			IsRead:       a.IsRead,
		}
	}

	return results, total, nil
}

// GetAnnouncement retrieves a single announcement by ID.
func (s *Service) GetAnnouncement(ctx context.Context, id int64) (*repository.Announcement, error) {
	return s.announcementRepo.GetByID(ctx, id)
}

// MarkAsRead marks an announcement as read for a user.
func (s *Service) MarkAsRead(ctx context.Context, userID, announcementID int64) error {
	// Verify announcement exists and is published
	announcement, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil {
		return err
	}

	if !announcement.IsPublished {
		return nil // Silently ignore unpublished announcements
	}

	return s.announcementRepo.MarkAsRead(ctx, userID, announcementID)
}

// IsRead checks if an announcement is read by a user.
func (s *Service) IsRead(ctx context.Context, userID, announcementID int64) (bool, error) {
	return s.announcementRepo.IsRead(ctx, userID, announcementID)
}

// GetUnreadCount gets the count of unread announcements for a user.
func (s *Service) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	return s.announcementRepo.GetUnreadCount(ctx, userID)
}

// GetRecentAnnouncements retrieves the most recent published announcements.
func (s *Service) GetRecentAnnouncements(ctx context.Context, limit int) ([]*repository.Announcement, error) {
	return s.announcementRepo.GetRecent(ctx, limit)
}

// ListByCategory retrieves announcements by category.
func (s *Service) ListByCategory(ctx context.Context, category string, limit, offset int) ([]*repository.Announcement, int64, error) {
	isPublished := true
	return s.announcementRepo.List(ctx, &repository.AnnouncementFilter{
		Category:    &category,
		IsPublished: &isPublished,
		Limit:       limit,
		Offset:      offset,
	})
}
