package main

import (
	"log"
	"net"

	"booking-service/internal/config"
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

	repo := repository.NewBookingRepository(db)
	uc := usecase.NewBookingUsecase(repo)
	server := grpcServer.NewBookingServer(uc)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	bookingpb.RegisterBookingServiceServer(s, server)

	log.Println("Booking Service listening on :" + cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
