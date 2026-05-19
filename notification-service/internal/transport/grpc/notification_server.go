package grpc

import (
	"context"
	"net/mail"

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
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient email is required")
	}

	recipient, err := mail.ParseAddress(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a valid email address")
	}

	if err := s.usecase.SendNotification(ctx, recipient.Address, req.Message, req.Type); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &notificationpb.NotificationResponse{
		Message: "Email sent successfully",
	}, nil
}

func (s *NotificationServer) GetNotification(ctx context.Context, req *notificationpb.GetNotificationRequest) (*notificationpb.NotificationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetNotification is not supported")
}

func (s *NotificationServer) ListUserNotifications(ctx context.Context, req *notificationpb.ListUserNotificationsRequest) (*notificationpb.NotificationsListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListUserNotifications is not supported")
}

func (s *NotificationServer) MarkAsRead(ctx context.Context, req *notificationpb.MarkAsReadRequest) (*notificationpb.NotificationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "MarkAsRead is not supported")
}
