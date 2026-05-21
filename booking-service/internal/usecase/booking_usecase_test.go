package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"booking-service/internal/domain"
	"booking-service/internal/messaging"
)

type fakeBookingRepository struct {
	createCalled     bool
	hasConflictValue bool
	hasConflictError error
	createError      error
	createdBooking   *domain.Booking
	bookings         map[uint]*domain.Booking
}

func (f *fakeBookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	f.createCalled = true
	if f.createError != nil {
		return f.createError
	}
	booking.ID = 1
	f.createdBooking = booking
	if f.bookings == nil {
		f.bookings = make(map[uint]*domain.Booking)
	}
	f.bookings[booking.ID] = booking
	return nil
}

func (f *fakeBookingRepository) CreateWithConflictCheck(ctx context.Context, booking *domain.Booking) error {
	hasConflict, err := f.HasConflict(ctx, booking.RoomID, booking.StartTime, booking.EndTime)
	if err != nil {
		return err
	}
	if hasConflict {
		return errors.New("room is already booked for this time")
	}
	return f.Create(ctx, booking)
}

func (f *fakeBookingRepository) GetByID(ctx context.Context, id uint) (*domain.Booking, error) {
	if booking, ok := f.bookings[id]; ok {
		return booking, nil
	}
	return nil, errors.New("booking not found")
}

func (f *fakeBookingRepository) ListByUserID(ctx context.Context, userID uint) ([]domain.Booking, error) {
	var result []domain.Booking
	for _, booking := range f.bookings {
		if booking.UserID == userID {
			result = append(result, *booking)
		}
	}
	return result, nil
}

func (f *fakeBookingRepository) ListAll(ctx context.Context) ([]domain.Booking, error) {
	var result []domain.Booking
	for _, booking := range f.bookings {
		result = append(result, *booking)
	}
	return result, nil
}

func (f *fakeBookingRepository) Cancel(ctx context.Context, id uint, userID uint) error {
	booking, ok := f.bookings[id]
	if !ok || booking.UserID != userID {
		return errors.New("booking not found")
	}
	booking.Status = StatusCancelled
	return nil
}

func (f *fakeBookingRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	booking, ok := f.bookings[id]
	if !ok {
		return errors.New("booking not found")
	}
	booking.Status = status
	return nil
}

func (f *fakeBookingRepository) HasConflict(ctx context.Context, roomID uint, startTime time.Time, endTime time.Time) (bool, error) {
	if f.hasConflictError != nil {
		return false, f.hasConflictError
	}
	return f.hasConflictValue, nil
}

type fakeNATSPublisher struct {
	publishCalled bool
	event         messaging.BookingCreatedEvent
	err           error
}

func (f *fakeNATSPublisher) PublishBookingCreated(event messaging.BookingCreatedEvent) error {
	f.publishCalled = true
	f.event = event
	return f.err
}

func (f *fakeNATSPublisher) PublishBookingStatusChanged(event messaging.BookingStatusChangedEvent) error {
	return f.err
}

func TestCreateBookingSuccess(t *testing.T) {
	repo := &fakeBookingRepository{}
	publisher := &fakeNATSPublisher{}

	uc := NewBookingUsecase(repo, publisher)

	startTime := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 20, 11, 0, 0, 0, time.UTC)

	booking, err := uc.CreateBooking(context.Background(), 10, 5, startTime, endTime, "Study session")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if booking == nil {
		t.Fatal("expected booking, got nil")
	}
	if booking.ID != 1 {
		t.Errorf("expected booking ID 1, got %d", booking.ID)
	}
	if booking.UserID != 10 {
		t.Errorf("expected user ID 10, got %d", booking.UserID)
	}
	if booking.RoomID != 5 {
		t.Errorf("expected room ID 5, got %d", booking.RoomID)
	}
	if !repo.createCalled {
		t.Error("expected repository Create to be called")
	}
	if !publisher.publishCalled {
		t.Error("expected NATS publisher to be called")
	}
	if publisher.event.UserID != booking.UserID {
		t.Errorf("expected event user ID %d, got %d", booking.UserID, publisher.event.UserID)
	}
	if publisher.event.BookingID != booking.ID {
		t.Errorf("expected event booking ID %d, got %d", booking.ID, publisher.event.BookingID)
	}
	if publisher.event.RoomID != booking.RoomID {
		t.Errorf("expected event room ID %d, got %d", booking.RoomID, publisher.event.RoomID)
	}
	if publisher.event.Type != "BOOKING_CREATED" {
		t.Errorf("expected event type BOOKING_CREATED, got %s", publisher.event.Type)
	}
}

func TestCreateBookingInvalidTime(t *testing.T) {
	repo := &fakeBookingRepository{}
	publisher := &fakeNATSPublisher{}

	uc := NewBookingUsecase(repo, publisher)

	startTime := time.Date(2026, 5, 20, 11, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)

	booking, err := uc.CreateBooking(context.Background(), 10, 5, startTime, endTime, "Invalid time test")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if booking != nil {
		t.Fatal("expected nil booking")
	}
	if repo.createCalled {
		t.Error("repository Create should not be called")
	}
	if publisher.publishCalled {
		t.Error("NATS publisher should not be called")
	}
}

func TestCreateBookingWithConflict(t *testing.T) {
	repo := &fakeBookingRepository{hasConflictValue: true}
	publisher := &fakeNATSPublisher{}

	uc := NewBookingUsecase(repo, publisher)

	startTime := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 20, 11, 0, 0, 0, time.UTC)

	booking, err := uc.CreateBooking(context.Background(), 10, 5, startTime, endTime, "Conflict test")

	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	if booking != nil {
		t.Fatal("expected nil booking")
	}
	if repo.createCalled {
		t.Error("repository Create should not be called when conflict exists")
	}
	if publisher.publishCalled {
		t.Error("NATS publisher should not be called when conflict exists")
	}
}

func TestUpdateBookingStatusSuccess(t *testing.T) {
	repo := &fakeBookingRepository{
		bookings: map[uint]*domain.Booking{
			1: {ID: 1, UserID: 10, RoomID: 5, Status: StatusConfirmed},
		},
	}

	uc := NewBookingUsecase(repo, nil)

	booking, err := uc.UpdateBookingStatus(context.Background(), 1, "cancelled")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if booking.Status != StatusCancelled {
		t.Errorf("expected status %s, got %s", StatusCancelled, booking.Status)
	}
}

func TestUpdateBookingStatusInvalid(t *testing.T) {
	repo := &fakeBookingRepository{
		bookings: map[uint]*domain.Booking{
			1: {ID: 1, UserID: 10, RoomID: 5, Status: StatusConfirmed},
		},
	}

	uc := NewBookingUsecase(repo, nil)

	booking, err := uc.UpdateBookingStatus(context.Background(), 1, "wrong-status")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if booking != nil {
		t.Fatal("expected nil booking")
	}
}
