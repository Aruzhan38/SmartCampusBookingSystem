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
	StatusConfirmed = "Confirmed"
	StatusCancelled = "Cancelled"
	StatusRejected  = "Rejected"
)

type BookingUsecase interface {
	CreateBooking(ctx context.Context, userID, roomID uint, startTime, endTime time.Time, purpose string) (*domain.Booking, error)
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

func (u *bookingUsecase) CreateBooking(ctx context.Context, userID, roomID uint, startTime, endTime time.Time, purpose string) (*domain.Booking, error) {
	if !startTime.Before(endTime) {
		return nil, errors.New("start_time must be before end_time")
	}

	hasConflict, err := u.repo.HasConflict(ctx, roomID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	if hasConflict {
		return nil, errors.New("room is already booked for this time")
	}

	booking := &domain.Booking{
		RoomID:    roomID,
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		Purpose:   purpose,
		Status:    StatusConfirmed,
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
	case StatusConfirmed, StatusCancelled, StatusRejected:
		return true
	default:
		return false
	}
}
