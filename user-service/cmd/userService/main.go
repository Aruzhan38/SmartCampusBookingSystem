package main

import (
	"log"
	"net"
	"user-service/internal/config"
	"user-service/internal/domain"
	"user-service/internal/repository"
	grpcServer "user-service/internal/transport/grpc"
	"user-service/internal/usecase"

	pb "github.com/Aruzhan38/smart-campus-generated/proto/user"
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
	db.AutoMigrate(&domain.User{})
	repo := repository.NewUserRepository(db)
	uc := usecase.NewUserUsecase(repo, cfg.JWTSecret)
	server := grpcServer.NewUserServer(uc)
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, server)
	log.Println("User Service listening on :" + cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
