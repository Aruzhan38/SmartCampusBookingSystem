package client

import (
	"context"
	"strconv"

	notificationpb "github.com/Aruzhan38/smart-campus-generated/proto/notification"
	"google.golang.org/grpc"
)

type NotificationClient interface {
	SendNotification(ctx context.Context, recipientEmail, message, notificationType string) (*notificationpb.NotificationResponse, error)
	GetNotificationByID(ctx context.Context, id string) (*notificationpb.NotificationResponse, error)
	ListUserNotifications(ctx context.Context, userID uint) (*notificationpb.NotificationsListResponse, error)
	MarkAsRead(ctx context.Context, id string) (*notificationpb.NotificationResponse, error)
}

type notificationClient struct {
	client notificationpb.NotificationServiceClient
}

func NewNotificationClient(conn *grpc.ClientConn) NotificationClient {
	return &notificationClient{client: notificationpb.NewNotificationServiceClient(conn)}
}

func (c *notificationClient) SendNotification(ctx context.Context, recipientEmail, message, notificationType string) (*notificationpb.NotificationResponse, error) {
	return c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:  recipientEmail,
		Message: message,
		Type:    notificationType,
	})
}

func (c *notificationClient) GetNotificationByID(ctx context.Context, id string) (*notificationpb.NotificationResponse, error) {
	return c.client.GetNotification(ctx, &notificationpb.GetNotificationRequest{NotificationId: id})
}

func (c *notificationClient) ListUserNotifications(ctx context.Context, userID uint) (*notificationpb.NotificationsListResponse, error) {
	return c.client.ListUserNotifications(ctx, &notificationpb.ListUserNotificationsRequest{UserId: strconv.FormatUint(uint64(userID), 10)})
}

func (c *notificationClient) MarkAsRead(ctx context.Context, id string) (*notificationpb.NotificationResponse, error) {
	return c.client.MarkAsRead(ctx, &notificationpb.MarkAsReadRequest{NotificationId: id})
}
