package repository

import (
	"context"
	"time"

	"booking-service/internal/domain"

	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id uint) (*domain.Booking, error)
	ListByUserID(ctx context.Context, userID uint) ([]domain.Booking, error)
	Cancel(ctx context.Context, id uint, userID uint) error
	UpdateStatus(ctx context.Context, id uint, status string) error
	HasConflict(ctx context.Context, roomID uint, startTime time.Time, endTime time.Time) (bool, error)
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

func (r *bookingRepository) GetByID(ctx context.Context, id uint) (*domain.Booking, error) {
	var booking domain.Booking
	if err := r.db.WithContext(ctx).First(&booking, id).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) ListByUserID(ctx context.Context, userID uint) ([]domain.Booking, error) {
	var bookings []domain.Booking
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("start_time desc").
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) Cancel(ctx context.Context, id uint, userID uint) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Booking{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", "Cancelled")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *bookingRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Booking{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *bookingRepository) HasConflict(ctx context.Context, roomID uint, startTime time.Time, endTime time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Booking{}).
		Where("room_id = ? AND status <> ? AND start_time < ? AND end_time > ?", roomID, "Cancelled", endTime, startTime).
		Count(&count).Error
	return count > 0, err
}
