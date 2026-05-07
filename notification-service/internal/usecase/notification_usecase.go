package usecase

import (
	"context"

	"notification-service/internal/domain"
	"notification-service/internal/repository"
)

type NotificationUsecase interface {
	SendNotification(ctx context.Context, userID uint, message, notificationType string) (*domain.Notification, error)
	GetNotification(ctx context.Context, id uint) (*domain.Notification, error)
	ListUserNotifications(ctx context.Context, userID uint) ([]domain.Notification, error)
	MarkAsRead(ctx context.Context, id uint) (*domain.Notification, error)
}

type notificationUsecase struct {
	repo repository.NotificationRepository
}

func NewNotificationUsecase(repo repository.NotificationRepository) NotificationUsecase {
	return &notificationUsecase{repo: repo}
}

func (u *notificationUsecase) SendNotification(ctx context.Context, userID uint, message, notificationType string) (*domain.Notification, error) {
	notification := &domain.Notification{
		UserID:  userID,
		Message: message,
		Type:    notificationType,
		IsRead:  false,
	}
	return u.repo.Create(ctx, notification)
}

func (u *notificationUsecase) GetNotification(ctx context.Context, id uint) (*domain.Notification, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *notificationUsecase) ListUserNotifications(ctx context.Context, userID uint) ([]domain.Notification, error) {
	return u.repo.ListByUserID(ctx, userID)
}

func (u *notificationUsecase) MarkAsRead(ctx context.Context, id uint) (*domain.Notification, error) {
	notification, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	notification.IsRead = true
	return u.repo.Update(ctx, notification)
}
