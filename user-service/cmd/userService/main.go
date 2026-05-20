package main

import (
	"log"
	"net"
	"strings"
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
	// If using a pooled Postgres (e.g., Supabase pooler/pgbouncer), prefer simple protocol
	// to avoid server-side prepared statement issues (prepared statement already exists).
	if strings.Contains(dsn, "pooler") && !strings.Contains(dsn, "prefer_simple_protocol") {
		if strings.Contains(dsn, "?") {
			dsn = dsn + "&prefer_simple_protocol=true"
		} else {
			dsn = dsn + "?prefer_simple_protocol=true"
		}
	}
	if dsn == "" {
		dsn = "host=" + cfg.DBHost + " user=" + cfg.DBUser + " password=" + cfg.DBPassword + " dbname=" + cfg.DBName + " port=" + cfg.DBPort + " sslmode=require channel_binding=require"
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		log.Fatal(err)
	}
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
