package usecase

import (
	"context"
	"strings"

	"notification-service/internal/mail"
)

type NotificationUsecase interface {
	SendNotification(ctx context.Context, recipientEmail, message, notificationType string) error
}

type notificationUsecase struct {
	sender mail.Sender
}

func NewNotificationUsecase(sender mail.Sender) NotificationUsecase {
	return &notificationUsecase{sender: sender}
}

func (u *notificationUsecase) SendNotification(ctx context.Context, recipientEmail, message, notificationType string) error {
	subject := "Smart Campus Notification"
	switch strings.ToLower(notificationType) {
	case "booking_confirmed":
		subject = "Your booking has been approved"
	case "booking_rejected":
		subject = "Your booking request was rejected"
	case "booking_cancelled":
		subject = "Your booking has been cancelled"
	case "booking_created":
		subject = "Booking created"
	}
	return u.sender.SendEmail(ctx, recipientEmail, subject, message)
}
