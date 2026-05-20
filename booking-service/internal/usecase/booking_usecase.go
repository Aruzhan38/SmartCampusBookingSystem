package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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
	repo      repository.BookingRepository
	publisher messaging.NATSPublisher
}

func NewBookingUsecase(repo repository.BookingRepository, publisher messaging.NATSPublisher) BookingUsecase {
	return &bookingUsecase{
		repo:      repo,
		publisher: publisher,
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
		err := u.publisher.PublishBookingCreated(messaging.BookingCreatedEvent{
			UserID:    booking.UserID,
			BookingID: booking.ID,
			RoomID:    booking.RoomID,
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

	return u.repo.GetByID(ctx, id)
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
