package main

import (
	"api-gateway/internal/client"
	"api-gateway/internal/config"
	"api-gateway/internal/middleware"
	gatewayhttp "api-gateway/internal/transport/http"
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func dialGRPC(address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}

func main() {
	cfg := config.Load()
	// Connect to User Service
	userConn, err := dialGRPC(cfg.UserServiceAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer userConn.Close()
	userClient := client.NewUserClient(userConn)
	// Connect to Room Service
	roomConn, err := dialGRPC(cfg.RoomServiceAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer roomConn.Close()
	roomClient := client.NewRoomClient(roomConn)
	// Connect to Booking Service
	bookingConn, err := dialGRPC(cfg.BookingServiceAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer bookingConn.Close()
	bookingClient := client.NewBookingClient(bookingConn)
	// HTTP Server
	r := gin.Default()

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
	r.GET("/bookings", webHandler.BookingsPage)

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
