package main

import (
	"log"
	"net"

	"room-service/internal/config"
	"room-service/internal/domain"
	"room-service/internal/repository"
	grpcServer "room-service/internal/transport/grpc"
	"room-service/internal/usecase"

	roompb "github.com/Aruzhan38/smart-campus-generated/proto/room"

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

	if err := db.AutoMigrate(&domain.Room{}); err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	repo := repository.NewRoomRepository(db)
	uc := usecase.NewRoomUsecase(repo)
	server := grpcServer.NewRoomServer(uc)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	roompb.RegisterRoomServiceServer(s, server)

	log.Println("Room Service listening on :" + cfg.GRPCPort)

	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
