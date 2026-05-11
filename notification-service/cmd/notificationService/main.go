package main

import (
	"log"
	"net"

	"notification-service/internal/config"
	"notification-service/internal/domain"
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
