package grpc

import (
	"context"
	"strconv"

	"notification-service/internal/domain"
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

func (s *NotificationServer) SendNotification(
	ctx context.Context,
	req *notificationpb.SendNotificationRequest,
) (*notificationpb.NotificationResponse, error) {
	userID, err := parsePositiveUint(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a positive integer")
	}

	notification, err := s.usecase.SendNotification(
		ctx,
		uint(userID),
		req.Message,
		req.Type,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &notificationpb.NotificationResponse{
		Notification: toProtoNotification(notification, req.Title),
		Message:      "Notification sent successfully",
	}, nil
}

func (s *NotificationServer) GetNotificationsByUser(
	ctx context.Context,
	req *notificationpb.GetNotificationsByUserRequest,
) (*notificationpb.ListNotificationsResponse, error) {
	userID, err := parsePositiveUint(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a positive integer")
	}

	notifications, err := s.usecase.ListUserNotifications(ctx, uint(userID))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoNotifications := make([]*notificationpb.Notification, 0, len(notifications))

	for i := range notifications {
		protoNotifications = append(
			protoNotifications,
			toProtoNotification(&notifications[i], ""),
		)
	}

	return &notificationpb.ListNotificationsResponse{
		Notifications: protoNotifications,
		Message:       "Notifications retrieved successfully",
	}, nil
}

func (s *NotificationServer) MarkNotificationAsRead(
	ctx context.Context,
	req *notificationpb.MarkNotificationAsReadRequest,
) (*notificationpb.NotificationResponse, error) {
	notificationID, err := parsePositiveUint(req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "notification_id must be a positive integer")
	}

	notification, err := s.usecase.MarkAsRead(ctx, uint(notificationID))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &notificationpb.NotificationResponse{
		Notification: toProtoNotification(notification, ""),
		Message:      "Notification marked as read",
	}, nil
}

func parsePositiveUint(value string) (uint64, error) {
	num, err := strconv.ParseUint(value, 10, 64)
	if err != nil || num == 0 {
		return 0, status.Error(codes.InvalidArgument, "invalid ID format")
	}

	return num, nil
}

func toProtoNotification(n *domain.Notification, title string) *notificationpb.Notification {
	if n == nil {
		return nil
	}

	if title == "" {
		title = "Smart Campus Notification"
	}

	statusValue := "UNREAD"
	if n.IsRead {
		statusValue = "READ"
	}

	return &notificationpb.Notification{
		Id:        strconv.FormatUint(uint64(n.ID), 10),
		UserId:    strconv.FormatUint(uint64(n.UserID), 10),
		Title:     title,
		Message:   n.Message,
		Type:      n.Type,
		Status:    statusValue,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
