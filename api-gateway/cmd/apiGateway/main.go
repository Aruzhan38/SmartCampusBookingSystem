package main

import (
	"api-gateway/internal/client"
	"api-gateway/internal/config"
	"api-gateway/internal/metrics"
	"api-gateway/internal/middleware"

	gatewayhttp "api-gateway/internal/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	// Connect to Booking Service
	bookingConn, err := grpc.Dial(cfg.BookingServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer bookingConn.Close()
	bookingClient := client.NewBookingClient(bookingConn)
	// HTTP Server
	r := gin.Default()
	r.Use(metrics.PrometheusMiddleware())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	webHandler := gatewayhttp.NewWebHandler()
	r.SetHTMLTemplate(webHandler.Templates())

	userHandler := gatewayhttp.NewUserHandler(userClient)
	roomHandler := gatewayhttp.NewRoomHandler(roomClient)
	bookingHandler := gatewayhttp.NewBookingHandler(bookingClient)

	r.GET("/", webHandler.Index)
	r.GET("/login", webHandler.LoginPage)
	r.GET("/register", webHandler.RegisterPage)
	r.GET("/rooms-ui", webHandler.RoomsPage)
	r.GET("/bookings-ui", webHandler.BookingsPage)
	r.GET("/admin-ui", webHandler.AdminPage)
	r.GET("/profile-ui", webHandler.ProfilePage)

	r.POST("/api/register", userHandler.Register)
	r.POST("/api/login", userHandler.Login)

	protected := r.Group("/")
	protected.Use(middleware.JWTAuthMiddleware(userClient))
	protected.GET("/rooms", roomHandler.GetRooms)
	protected.POST("/rooms", roomHandler.CreateRoom)
	protected.GET("/rooms/search", roomHandler.SearchRoomsByCapacity)
	protected.GET("/rooms/:id", roomHandler.GetRoomByID)
	protected.PUT("/rooms/:id", roomHandler.UpdateRoom)
	protected.POST("/bookings", bookingHandler.CreateBooking)
	protected.GET("/bookings/my", bookingHandler.ListUserBookings)
	protected.GET("/bookings/:id", bookingHandler.GetBookingByID)
	protected.DELETE("/bookings/:id", bookingHandler.CancelBooking)
	protected.PATCH("/bookings/:id/status", bookingHandler.UpdateBookingStatus)

	log.Println("API Gateway listening on :" + cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatal(err)
	}
}
