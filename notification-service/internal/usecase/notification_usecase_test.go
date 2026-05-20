package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"notification-service/internal/domain"
)

type fakeNotificationRepository struct {
	createCalled     bool
	getByIDCalled    bool
	listByUserCalled bool
	updateCalled     bool

	createError     error
	getByIDError    error
	listByUserError error
	updateError     error

	notifications map[uint]*domain.Notification
}

func (f *fakeNotificationRepository) Create(ctx context.Context, notification *domain.Notification) (*domain.Notification, error) {
	_ = ctx

	f.createCalled = true

	if f.createError != nil {
		return nil, f.createError
	}

	if f.notifications == nil {
		f.notifications = make(map[uint]*domain.Notification)
	}

	if notification.ID == 0 {
		notification.ID = uint(len(f.notifications) + 1)
	}

	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now().UTC()
	}

	f.notifications[notification.ID] = notification

	return notification, nil
}

func (f *fakeNotificationRepository) GetByID(ctx context.Context, id uint) (*domain.Notification, error) {
	_ = ctx

	f.getByIDCalled = true

	if f.getByIDError != nil {
		return nil, f.getByIDError
	}

	notification, ok := f.notifications[id]
	if !ok {
		return nil, errors.New("notification not found")
	}

	return notification, nil
}

func (f *fakeNotificationRepository) ListByUserID(ctx context.Context, userID uint) ([]domain.Notification, error) {
	_ = ctx

	f.listByUserCalled = true

	if f.listByUserError != nil {
		return nil, f.listByUserError
	}

	var result []domain.Notification

	for _, notification := range f.notifications {
		if notification.UserID == userID {
			result = append(result, *notification)
		}
	}

	return result, nil
}

func (f *fakeNotificationRepository) Update(ctx context.Context, notification *domain.Notification) (*domain.Notification, error) {
	_ = ctx

	f.updateCalled = true

	if f.updateError != nil {
		return nil, f.updateError
	}

	if notification == nil {
		return nil, errors.New("notification is nil")
	}

	if f.notifications == nil {
		return nil, errors.New("notification not found")
	}

	existingNotification, ok := f.notifications[notification.ID]
	if !ok {
		return nil, errors.New("notification not found")
	}

	existingNotification.UserID = notification.UserID
	existingNotification.Message = notification.Message
	existingNotification.Type = notification.Type
	existingNotification.IsRead = notification.IsRead
	existingNotification.CreatedAt = notification.CreatedAt

	return existingNotification, nil
}

func TestSendNotificationSuccess(t *testing.T) {
	repo := &fakeNotificationRepository{}
	uc := NewNotificationUsecase(repo)

	notification, err := uc.SendNotification(
		context.Background(),
		10,
		"Your booking has been created successfully.",
		"BOOKING_CREATED",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if notification == nil {
		t.Fatal("expected notification, got nil")
	}

	if notification.ID == 0 {
		t.Error("expected notification ID to be assigned")
	}

	if notification.UserID != 10 {
		t.Errorf("expected user ID 10, got %d", notification.UserID)
	}

	if notification.Message != "Your booking has been created successfully." {
		t.Errorf("unexpected message: %s", notification.Message)
	}

	if notification.Type != "BOOKING_CREATED" {
		t.Errorf("expected type BOOKING_CREATED, got %s", notification.Type)
	}

	if notification.IsRead {
		t.Error("expected notification to be unread by default")
	}

	if !repo.createCalled {
		t.Error("expected repository Create to be called")
	}
}

func TestSendNotificationRepositoryError(t *testing.T) {
	repo := &fakeNotificationRepository{
		createError: errors.New("database error"),
	}

	uc := NewNotificationUsecase(repo)

	notification, err := uc.SendNotification(
		context.Background(),
		10,
		"Message",
		"BOOKING_CREATED",
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if notification != nil {
		t.Fatal("expected nil notification")
	}

	if !repo.createCalled {
		t.Error("expected repository Create to be called")
	}
}

func TestGetNotificationSuccess(t *testing.T) {
	repo := &fakeNotificationRepository{
		notifications: map[uint]*domain.Notification{
			1: {
				ID:        1,
				UserID:    10,
				Message:   "Booking created",
				Type:      "BOOKING_CREATED",
				IsRead:    false,
				CreatedAt: time.Now().UTC(),
			},
		},
	}

	uc := NewNotificationUsecase(repo)

	notification, err := uc.GetNotification(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if notification == nil {
		t.Fatal("expected notification, got nil")
	}

	if notification.ID != 1 {
		t.Errorf("expected notification ID 1, got %d", notification.ID)
	}

	if !repo.getByIDCalled {
		t.Error("expected repository GetByID to be called")
	}
}

func TestListUserNotificationsSuccess(t *testing.T) {
	repo := &fakeNotificationRepository{
		notifications: map[uint]*domain.Notification{
			1: {
				ID:      1,
				UserID:  10,
				Message: "Booking created",
				Type:    "BOOKING_CREATED",
				IsRead:  false,
			},
			2: {
				ID:      2,
				UserID:  10,
				Message: "Booking updated",
				Type:    "BOOKING_UPDATED",
				IsRead:  false,
			},
			3: {
				ID:      3,
				UserID:  20,
				Message: "Other user notification",
				Type:    "BOOKING_CREATED",
				IsRead:  false,
			},
		},
	}

	uc := NewNotificationUsecase(repo)

	notifications, err := uc.ListUserNotifications(context.Background(), 10)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notifications) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(notifications))
	}

	if !repo.listByUserCalled {
		t.Error("expected repository ListByUserID to be called")
	}
}

func TestMarkAsReadSuccess(t *testing.T) {
	repo := &fakeNotificationRepository{
		notifications: map[uint]*domain.Notification{
			1: {
				ID:      1,
				UserID:  10,
				Message: "Booking created",
				Type:    "BOOKING_CREATED",
				IsRead:  false,
			},
		},
	}

	uc := NewNotificationUsecase(repo)

	notification, err := uc.MarkAsRead(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if notification == nil {
		t.Fatal("expected notification, got nil")
	}

	if !notification.IsRead {
		t.Error("expected notification to be marked as read")
	}

	if !repo.getByIDCalled {
		t.Error("expected repository GetByID to be called")
	}

	if !repo.updateCalled {
		t.Error("expected repository Update to be called")
	}
}

func TestMarkAsReadNotFound(t *testing.T) {
	repo := &fakeNotificationRepository{
		notifications: map[uint]*domain.Notification{},
	}

	uc := NewNotificationUsecase(repo)

	notification, err := uc.MarkAsRead(context.Background(), 1)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if notification != nil {
		t.Fatal("expected nil notification")
	}
}
