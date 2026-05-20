package main

import (
	"log"
	"net"

	"notification-service/internal/config"
	"notification-service/internal/domain"
	"notification-service/internal/mail"
	"notification-service/internal/messaging"
	"notification-service/internal/repository"
	grpcServer "notification-service/internal/transport/grpc"
	"notification-service/internal/usecase"

	notificationpb "github.com/Aruzhan38/smart-campus-generated/proto/notification"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DBURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	if err := db.AutoMigrate(&domain.Notification{}); err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	repo := repository.NewNotificationRepository(db)
	uc := usecase.NewNotificationUsecase(repo)
	server := grpcServer.NewNotificationServer(uc)

	mailSender := mail.NewSMTPSender(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)

	consumer := messaging.NewNATSConsumer(
		cfg.NATSURL,
		uc,
		mailSender,
		cfg.DefaultEmail,
	)

	go func() {
		if err := consumer.Start(); err != nil {
			log.Println("failed to start NATS consumer:", err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	s := grpc.NewServer()
	notificationpb.RegisterNotificationServiceServer(s, server)

	log.Println("Notification Service listening on :" + cfg.GRPCPort)
	log.Println("Notification Service connected to NATS:", cfg.NATSURL)

	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to serve grpc: ", err)
	}
}
