package grpc

import (
	"context"
	"errors"
	"strconv"
	"time"

	"booking-service/internal/domain"
	"booking-service/internal/usecase"

	bookingpb "github.com/Aruzhan38/smart-campus-generated/proto/booking"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type BookingServer struct {
	bookingpb.UnimplementedBookingServiceServer
	usecase usecase.BookingUsecase
}

func NewBookingServer(uc usecase.BookingUsecase) *BookingServer {
	return &BookingServer{usecase: uc}
}

func (s *BookingServer) CreateBooking(ctx context.Context, req *bookingpb.CreateBookingRequest) (*bookingpb.BookingResponse, error) {
	userID, err := parsePositiveUint(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a positive integer")
	}

	roomID, err := parsePositiveUint(req.RoomId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "room_id must be a positive integer")
	}

	startTime, err := parseRFC3339(req.StartTime, "start_time")
	if err != nil {
		return nil, err
	}

	endTime, err := parseRFC3339(req.EndTime, "end_time")
	if err != nil {
		return nil, err
	}

	booking, err := s.usecase.CreateBooking(ctx, uint(userID), uint(roomID), startTime, endTime, req.Purpose)
	if err != nil {
		return nil, bookingServiceError(err)
	}

	return &bookingpb.BookingResponse{
		Booking: toProtoBooking(booking),
		Message: "Booking created successfully",
	}, nil
}

func (s *BookingServer) GetBookingById(ctx context.Context, req *bookingpb.GetBookingByIdRequest) (*bookingpb.BookingResponse, error) {
	id, err := parsePositiveUint(req.BookingId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "booking_id must be a positive integer")
	}

	booking, err := s.usecase.GetBookingByID(ctx, uint(id))
	if err != nil {
		return nil, bookingServiceError(err)
	}

	return &bookingpb.BookingResponse{
		Booking: toProtoBooking(booking),
		Message: "Booking found",
	}, nil
}

func (s *BookingServer) ListUserBookings(ctx context.Context, req *bookingpb.ListUserBookingsRequest) (*bookingpb.ListBookingsResponse, error) {
	userID, err := parsePositiveUint(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a positive integer")
	}

	bookings, err := s.usecase.ListUserBookings(ctx, uint(userID))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &bookingpb.ListBookingsResponse{
		Bookings: make([]*bookingpb.Booking, 0, len(bookings)),
		Message:  "Bookings loaded",
	}
	for _, booking := range bookings {
		b := booking
		resp.Bookings = append(resp.Bookings, toProtoBooking(&b))
	}

	return resp, nil
}

func (s *BookingServer) CancelBooking(ctx context.Context, req *bookingpb.CancelBookingRequest) (*bookingpb.BookingResponse, error) {
	bookingID, err := parsePositiveUint(req.BookingId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "booking_id must be a positive integer")
	}

	userID, err := parsePositiveUint(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a positive integer")
	}

	booking, err := s.usecase.CancelBooking(ctx, uint(bookingID), uint(userID))
	if err != nil {
		return nil, bookingServiceError(err)
	}

	return &bookingpb.BookingResponse{
		Booking: toProtoBooking(booking),
		Message: "Booking cancelled successfully",
	}, nil
}

func (s *BookingServer) UpdateBookingStatus(ctx context.Context, req *bookingpb.UpdateBookingStatusRequest) (*bookingpb.BookingResponse, error) {
	bookingID, err := parsePositiveUint(req.BookingId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "booking_id must be a positive integer")
	}

	booking, err := s.usecase.UpdateBookingStatus(ctx, uint(bookingID), req.Status)
	if err != nil {
		return nil, bookingServiceError(err)
	}

	return &bookingpb.BookingResponse{
		Booking: toProtoBooking(booking),
		Message: "Booking status updated successfully",
	}, nil
}

func toProtoBooking(booking *domain.Booking) *bookingpb.Booking {
	return &bookingpb.Booking{
		Id:        strconv.Itoa(int(booking.ID)),
		RoomId:    strconv.Itoa(int(booking.RoomID)),
		UserId:    strconv.Itoa(int(booking.UserID)),
		StartTime: booking.StartTime.Format(time.RFC3339),
		EndTime:   booking.EndTime.Format(time.RFC3339),
		Purpose:   booking.Purpose,
		Status:    booking.Status,
		CreatedAt: booking.CreatedAt.Format(time.RFC3339),
	}
}

func parseRFC3339(value string, field string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, status.Errorf(codes.InvalidArgument, "%s must use RFC3339 format", field)
	}
	return parsed, nil
}

func parsePositiveUint(value string) (uint64, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}

func bookingServiceError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.NotFound, "booking not found")
	}
	switch err.Error() {
	case "room is already booked for this time":
		return status.Error(codes.AlreadyExists, err.Error())
	case "start_time must be before end_time", "invalid booking status":
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
