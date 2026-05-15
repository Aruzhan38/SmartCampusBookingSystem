package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"booking-service/internal/domain"
	"booking-service/internal/repository"
)

const (
	StatusPending   = "Pending"
	StatusConfirmed = "Confirmed"
	StatusCancelled = "Cancelled"
	StatusRejected  = "Rejected"
)

type BookingUsecase interface {
	CreateBooking(ctx context.Context, userID, roomID uint, startTime, endTime time.Time, purpose string, status string) (*domain.Booking, error)
	GetBookingByID(ctx context.Context, id uint) (*domain.Booking, error)
	ListUserBookings(ctx context.Context, userID uint) ([]domain.Booking, error)
	CancelBooking(ctx context.Context, id uint, userID uint) (*domain.Booking, error)
	UpdateBookingStatus(ctx context.Context, id uint, status string) (*domain.Booking, error)
}

type bookingUsecase struct {
	repo repository.BookingRepository
}

func NewBookingUsecase(repo repository.BookingRepository) BookingUsecase {
	return &bookingUsecase{repo: repo}
}

func (u *bookingUsecase) CreateBooking(ctx context.Context, userID, roomID uint, startTime, endTime time.Time, purpose string, status string) (*domain.Booking, error) {
	if !startTime.Before(endTime) {
		return nil, errors.New("start_time must be before end_time")
	}

	// Enforce bookings to be within a single day and within working hours
	if !sameDay(startTime, endTime) {
		return nil, errors.New("booking must be within a single calendar day")
	}
	// Disallow any bookings on Sundays
	if startTime.Weekday() == time.Sunday {
		return nil, errors.New("bookings are not allowed on Sundays")
	}
	// Allowed hours: from 08:00:00 (inclusive) to 22:00:00 (inclusive end)
	dayStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 8, 0, 0, 0, startTime.Location())
	dayEnd := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 22, 0, 0, 0, endTime.Location())
	if startTime.Before(dayStart) || endTime.After(dayEnd) {
		return nil, errors.New("bookings allowed only between 08:00 and 22:00 on non-Sundays")
	}

	hasConflict, err := u.repo.HasConflict(ctx, roomID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	if hasConflict {
		return nil, errors.New("room is already booked for this time")
	}

	if status == "" {
		status = StatusConfirmed
	}

	booking := &domain.Booking{
		RoomID:    roomID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		Purpose:   purpose,
		Status:    status,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.repo.Create(ctx, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

func (u *bookingUsecase) GetBookingByID(ctx context.Context, id uint) (*domain.Booking, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *bookingUsecase) ListUserBookings(ctx context.Context, userID uint) ([]domain.Booking, error) {
	return u.repo.ListByUserID(ctx, userID)
}

func (u *bookingUsecase) CancelBooking(ctx context.Context, id uint, userID uint) (*domain.Booking, error) {
	if err := u.repo.Cancel(ctx, id, userID); err != nil {
		return nil, err
	}
	return u.repo.GetByID(ctx, id)
}

func (u *bookingUsecase) UpdateBookingStatus(ctx context.Context, id uint, status string) (*domain.Booking, error) {
	status = normalizeStatus(strings.TrimSpace(status))
	if status == "" {
		return nil, errors.New("invalid booking status")
	}

	if err := u.repo.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}
	return u.repo.GetByID(ctx, id)
}

func normalizeStatus(status string) string {
	switch strings.ToUpper(status) {
	case "PENDING":
		return StatusPending
	case "CONFIRMED":
		return StatusConfirmed
	case "CANCELLED", "CANCELED":
		return StatusCancelled
	case "REJECTED":
		return StatusRejected
	default:
		return ""
	}
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
