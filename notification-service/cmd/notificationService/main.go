package main

import (
	"log"
	"net"

	"notification-service/internal/config"
	"notification-service/internal/mail"
	grpcServer "notification-service/internal/transport/grpc"
	"notification-service/internal/usecase"

	notificationpb "github.com/Aruzhan38/smart-campus-generated/proto/notification"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	sender := mail.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)
	uc := usecase.NewNotificationUsecase(sender)
	server := grpcServer.NewNotificationServer(uc)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	notificationpb.RegisterNotificationServiceServer(s, server)

	log.Println("Notification Service listening on :" + cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
