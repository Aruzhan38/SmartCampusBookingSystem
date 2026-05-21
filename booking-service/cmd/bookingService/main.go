package main

import (
	"log"
	"net"

	"booking-service/internal/client"
	"booking-service/internal/config"
	"booking-service/internal/domain"
	"booking-service/internal/messaging"
	"booking-service/internal/repository"
	grpcServer "booking-service/internal/transport/grpc"
	"booking-service/internal/usecase"

	bookingpb "github.com/Aruzhan38/smart-campus-generated/proto/booking"
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

	if err := db.AutoMigrate(&domain.Booking{}); err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	repo := repository.NewBookingRepository(db)
	publisher, err := messaging.NewNATSPublisher(cfg.NATSURL)
	if err != nil {
		log.Println("failed to connect to NATS:", err)
		log.Println("booking service will continue without notification events")
	}
	var userClientInstance client.UserClient
	if cfg.UserServiceAddr != "" {
		userConn, err := client.DialUserService(cfg.UserServiceAddr)
		if err != nil {
			log.Println("failed to connect to user service:", err)
		} else {
			userClientInstance = client.NewUserClient(userConn)
		}
	}

	uc := usecase.NewBookingUsecase(repo, publisher, userClientInstance)
	server := grpcServer.NewBookingServer(uc)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	s := grpc.NewServer()
	bookingpb.RegisterBookingServiceServer(s, server)

	log.Println("Booking Service listening on :" + cfg.GRPCPort)
	log.Println("Booking Service NATS URL:", cfg.NATSURL)

	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to serve grpc: ", err)
	}
}
