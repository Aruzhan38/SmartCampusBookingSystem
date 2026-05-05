package main

import (
	"log"
	"net"
	"room-service/internal/config"
	"room-service/internal/repository"
	grpcServer "room-service/internal/transport/grpc"
	"room-service/internal/usecase"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	dsn := cfg.DBURL
	if dsn == "" {
		dsn = "host=" + cfg.DBHost + " user=" + cfg.DBUser + " password=" + cfg.DBPassword + " dbname=" + cfg.DBName + " port=" + cfg.DBPort + " sslmode=require channel_binding=require"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	// Auto-migrate models if needed
	repo := repository.NewRoomRepository(db)
	uc := usecase.NewRoomUsecase(repo)
	server := grpcServer.NewRoomServer(uc)
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	// Register the service (assuming pb is imported)
	log.Println("Room Service listening on :" + cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
