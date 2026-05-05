package main

import (
	"api-gateway/internal/client"
	"api-gateway/internal/config"
	"api-gateway/internal/middleware"
	gatewayhttp "api-gateway/internal/transport/http"
	"log"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.Load()
	// Connect to User Service
	userConn, err := grpc.Dial(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer userConn.Close()
	userClient := client.NewUserClient(userConn)
	// Connect to Room Service
	roomConn, err := grpc.Dial(cfg.RoomServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer roomConn.Close()
	roomClient := client.NewRoomClient(roomConn)
	// HTTP Server
	r := gin.Default()

	webHandler := gatewayhttp.NewWebHandler()
	r.SetHTMLTemplate(webHandler.Templates())

	userHandler := gatewayhttp.NewUserHandler(userClient)
	roomHandler := gatewayhttp.NewRoomHandler(roomClient)

	r.GET("/", webHandler.Index)
	r.GET("/login", webHandler.LoginPage)
	r.GET("/register", webHandler.RegisterPage)
	r.GET("/rooms-ui", webHandler.RoomsPage)

	r.POST("/api/register", userHandler.Register)
	r.POST("/api/login", userHandler.Login)

	protected := r.Group("/")
	protected.Use(middleware.JWTAuthMiddleware(userClient))
	protected.GET("/rooms", roomHandler.GetRooms)

	log.Println("API Gateway listening on :" + cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatal(err)
	}
}
