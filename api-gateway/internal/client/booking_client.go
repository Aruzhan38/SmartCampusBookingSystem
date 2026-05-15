package client

import (
	"context"
	"strconv"

	bookingpb "github.com/Aruzhan38/smart-campus-generated/proto/booking"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type BookingClient interface {
	CreateBooking(ctx context.Context, userID uint, roomID uint, startTime string, endTime string, purpose string, role string) (*bookingpb.BookingResponse, error)
	GetBookingByID(ctx context.Context, id string) (*bookingpb.BookingResponse, error)
	ListUserBookings(ctx context.Context, userID uint) (*bookingpb.ListBookingsResponse, error)
	CancelBooking(ctx context.Context, id string, userID uint) (*bookingpb.BookingResponse, error)
	UpdateBookingStatus(ctx context.Context, id string, status string) (*bookingpb.BookingResponse, error)
}

type bookingClient struct {
	client bookingpb.BookingServiceClient
}

func NewBookingClient(conn *grpc.ClientConn) BookingClient {
	return &bookingClient{
		client: bookingpb.NewBookingServiceClient(conn),
	}
}

func (c *bookingClient) CreateBooking(ctx context.Context, userID uint, roomID uint, startTime string, endTime string, purpose string, role string) (*bookingpb.BookingResponse, error) {
	if role != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("user-role", role))
	}
	return c.client.CreateBooking(ctx, &bookingpb.CreateBookingRequest{
		UserId:    strconv.Itoa(int(userID)),
		RoomId:    strconv.Itoa(int(roomID)),
		StartTime: startTime,
		EndTime:   endTime,
		Purpose:   purpose,
	})
}

func (c *bookingClient) GetBookingByID(ctx context.Context, id string) (*bookingpb.BookingResponse, error) {
	return c.client.GetBookingById(ctx, &bookingpb.GetBookingByIdRequest{
		BookingId: id,
	})
}

func (c *bookingClient) ListUserBookings(ctx context.Context, userID uint) (*bookingpb.ListBookingsResponse, error) {
	return c.client.ListUserBookings(ctx, &bookingpb.ListUserBookingsRequest{
		UserId: strconv.Itoa(int(userID)),
	})
}

func (c *bookingClient) CancelBooking(ctx context.Context, id string, userID uint) (*bookingpb.BookingResponse, error) {
	return c.client.CancelBooking(ctx, &bookingpb.CancelBookingRequest{
		BookingId: id,
		UserId:    strconv.Itoa(int(userID)),
	})
}

func (c *bookingClient) UpdateBookingStatus(ctx context.Context, id string, status string) (*bookingpb.BookingResponse, error) {
	return c.client.UpdateBookingStatus(ctx, &bookingpb.UpdateBookingStatusRequest{
		BookingId: id,
		Status:    status,
	})
}
