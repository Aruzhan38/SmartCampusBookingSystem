package grpc

import (
	"context"
	"strconv"

	"notification-service/internal/usecase"

	notificationpb "github.com/Aruzhan38/smart-campus-generated/proto/notification"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationServer struct {
	notificationpb.UnimplementedNotificationServiceServer
	usecase usecase.NotificationUsecase
}

func NewNotificationServer(uc usecase.NotificationUsecase) *NotificationServer {
	return &NotificationServer{usecase: uc}
}

func (s *NotificationServer) SendNotification(ctx context.Context, req *notificationpb.SendNotificationRequest) (*notificationpb.NotificationResponse, error) {
	userID, err := parsePositiveUint(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a positive integer")
	}

	notification, err := s.usecase.SendNotification(ctx, uint(userID), req.Message, req.Type)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &notificationpb.NotificationResponse{
		Notification: toProtoNotification(notification),
		Message:      "Notification sent successfully",
	}, nil
}

func (s *NotificationServer) GetNotification(ctx context.Context, req *notificationpb.GetNotificationRequest) (*notificationpb.NotificationResponse, error) {
	id64, err := parsePositiveUint(req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "notification_id must be a positive integer")
	}

	notification, err := s.usecase.GetNotification(ctx, uint(id64))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &notificationpb.NotificationResponse{
		Notification: toProtoNotification(notification),
		Message:      "Notification retrieved successfully",
	}, nil
}

func (s *NotificationServer) ListUserNotifications(ctx context.Context, req *notificationpb.ListUserNotificationsRequest) (*notificationpb.NotificationsListResponse, error) {
	userID, err := parsePositiveUint(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a positive integer")
	}

	notifications, err := s.usecase.ListUserNotifications(ctx, uint(userID))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoNotifications []*notificationpb.Notification
	for _, n := range notifications {
		notification := n
		protoNotifications = append(protoNotifications, toProtoNotification(&notification))
	}

	return &notificationpb.NotificationsListResponse{
		Notifications: protoNotifications,
		Count:         int32(len(protoNotifications)),
	}, nil
}

func (s *NotificationServer) MarkAsRead(ctx context.Context, req *notificationpb.MarkAsReadRequest) (*notificationpb.NotificationResponse, error) {
	id64, err := parsePositiveUint(req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "notification_id must be a positive integer")
	}

	notification, err := s.usecase.MarkAsRead(ctx, uint(id64))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &notificationpb.NotificationResponse{
		Notification: toProtoNotification(notification),
		Message:      "Notification marked as read",
	}, nil
}

func parsePositiveUint(s string) (int64, error) {
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil || num <= 0 {
		return 0, status.Error(codes.InvalidArgument, "invalid ID format")
	}
	return num, nil
}

func toProtoNotification(n *notificationpb.Notification) *notificationpb.Notification {
	if n == nil {
		return nil
	}
	return &notificationpb.Notification{
		Id:        strconv.FormatUint(uint64(n.Id), 10),
		UserId:    strconv.FormatUint(uint64(n.UserId), 10),
		Message:   n.Message,
		Type:      n.Type,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
	}
}
