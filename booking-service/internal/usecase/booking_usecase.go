package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"booking-service/internal/client"
	"booking-service/internal/domain"
	"booking-service/internal/messaging"
	"booking-service/internal/repository"
)

const (
	StatusPending   = "PENDING"
	StatusConfirmed = "CONFIRMED"
	StatusCancelled = "CANCELLED"
	StatusRejected  = "REJECTED"
)

type BookingUsecase interface {
	CreateBooking(ctx context.Context, userID, roomID uint, startTime, endTime time.Time, purpose string) (*domain.Booking, error)
	GetBookingByID(ctx context.Context, id uint) (*domain.Booking, error)
	ListUserBookings(ctx context.Context, userID uint) ([]domain.Booking, error)
	CancelBooking(ctx context.Context, id uint, userID uint) (*domain.Booking, error)
	UpdateBookingStatus(ctx context.Context, id uint, status string) (*domain.Booking, error)
}

type bookingUsecase struct {
	repo       repository.BookingRepository
	publisher  messaging.NATSPublisher
	userClient client.UserClient
}

func NewBookingUsecase(repo repository.BookingRepository, publisher messaging.NATSPublisher, userClients ...client.UserClient) BookingUsecase {
	var uc client.UserClient
	if len(userClients) > 0 {
		uc = userClients[0]
	}
	return &bookingUsecase{
		repo:       repo,
		publisher:  publisher,
		userClient: uc,
	}
}

func (u *bookingUsecase) CreateBooking(ctx context.Context, userID, roomID uint, startTime, endTime time.Time, purpose string) (*domain.Booking, error) {
	if !startTime.Before(endTime) {
		return nil, errors.New("start_time must be before end_time")
	}

	booking := &domain.Booking{
		RoomID:    roomID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		Purpose:   purpose,
		Status:    StatusPending,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.repo.CreateWithConflictCheck(ctx, booking); err != nil {
		return nil, err
	}

	if u.publisher != nil {
		email := ""
		if u.userClient != nil {
			uid := strconv.Itoa(int(booking.UserID))
			if usr, err := u.userClient.GetUserByID(ctx, uid); err == nil {
				email = usr.Email
			}
		}

		err := u.publisher.PublishBookingCreated(messaging.BookingCreatedEvent{
			UserID:    booking.UserID,
			BookingID: booking.ID,
			RoomID:    booking.RoomID,
			Email:     email,
			Message: fmt.Sprintf(
				"Your booking #%d for room #%d has been created successfully. Time: %s - %s.",
				booking.ID,
				booking.RoomID,
				booking.StartTime.Format("2006-01-02 15:04"),
				booking.EndTime.Format("2006-01-02 15:04"),
			),
			Type: "BOOKING_CREATED",
		})

		if err != nil {
			log.Println("failed to publish booking.created event:", err)
		}
	}

	return booking, nil
}

func (u *bookingUsecase) GetBookingByID(ctx context.Context, id uint) (*domain.Booking, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *bookingUsecase) ListUserBookings(ctx context.Context, userID uint) ([]domain.Booking, error) {
	if userID == 0 {
		return u.repo.ListAll(ctx)
	}
	return u.repo.ListByUserID(ctx, userID)
}

func (u *bookingUsecase) CancelBooking(ctx context.Context, id uint, userID uint) (*domain.Booking, error) {
	if err := u.repo.Cancel(ctx, id, userID); err != nil {
		return nil, err
	}

	return u.repo.GetByID(ctx, id)
}

func (u *bookingUsecase) UpdateBookingStatus(ctx context.Context, id uint, status string) (*domain.Booking, error) {
	status = strings.ToUpper(strings.TrimSpace(status))

	if !isValidStatus(status) {
		return nil, errors.New("invalid booking status")
	}

	if err := u.repo.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}

	booking, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// publish status change event
	if u.publisher != nil {
		email := ""
		if u.userClient != nil {
			// try to fetch user email
			uid := strconv.Itoa(int(booking.UserID))
			if usr, err := u.userClient.GetUserByID(ctx, uid); err == nil {
				email = usr.Email
			}
		}

		message := "Your booking status has been updated."
		switch status {
		case StatusConfirmed:
			message = "Your booking has been approved by admin."
		case StatusRejected:
			message = "Your booking has been rejected by admin."
		}

		_ = u.publisher.PublishBookingStatusChanged(messaging.BookingStatusChangedEvent{
			UserID:    booking.UserID,
			BookingID: booking.ID,
			RoomID:    booking.RoomID,
			Email:     email,
			Message:   message,
			Status:    status,
			Type:      "BOOKING_STATUS_CHANGED",
		})
	}

	return booking, nil
}

func isValidStatus(status string) bool {
	switch status {
	case StatusPending:
		return true
	case StatusConfirmed, StatusCancelled, StatusRejected:
		return true
	default:
		return false
	}
}
