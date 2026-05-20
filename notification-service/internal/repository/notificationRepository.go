package repository

import (
	"context"

	"notification-service/internal/domain"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) (*domain.Notification, error)
	GetByID(ctx context.Context, id uint) (*domain.Notification, error)
	ListByUserID(ctx context.Context, userID uint) ([]domain.Notification, error)
	Update(ctx context.Context, notification *domain.Notification) (*domain.Notification, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) (*domain.Notification, error) {
	result := r.db.WithContext(ctx).Create(notification)
	if result.Error != nil {
		return nil, result.Error
	}
	return notification, nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id uint) (*domain.Notification, error) {
	notification := &domain.Notification{}
	result := r.db.WithContext(ctx).First(notification, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return notification, nil
}

func (r *notificationRepository) ListByUserID(ctx context.Context, userID uint) ([]domain.Notification, error) {
	var notifications []domain.Notification
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&notifications)
	if result.Error != nil {
		return nil, result.Error
	}
	return notifications, nil
}

func (r *notificationRepository) Update(ctx context.Context, notification *domain.Notification) (*domain.Notification, error) {
	result := r.db.WithContext(ctx).Save(notification)
	if result.Error != nil {
		return nil, result.Error
	}
	return notification, nil
}
