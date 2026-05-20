package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"room-service/internal/domain"
)

type fakeRoomRepository struct {
	createCalled           bool
	updateCalled           bool
	listCalled             bool
	getByIDCalled          bool
	searchByCapacityCalled bool

	createError error
	updateError error
	getError    error
	listError   error
	searchError error

	rooms map[uint]*domain.Room
}

func (f *fakeRoomRepository) Create(ctx context.Context, room *domain.Room) error {
	f.createCalled = true

	if f.createError != nil {
		return f.createError
	}

	if f.rooms == nil {
		f.rooms = make(map[uint]*domain.Room)
	}

	if room.ID == 0 {
		room.ID = uint(len(f.rooms) + 1)
	}

	if room.CreatedAt.IsZero() {
		room.CreatedAt = time.Now().UTC()
	}

	f.rooms[room.ID] = room
	return nil
}

func (f *fakeRoomRepository) GetByID(ctx context.Context, id uint) (*domain.Room, error) {
	f.getByIDCalled = true

	if f.getError != nil {
		return nil, f.getError
	}

	room, ok := f.rooms[id]
	if !ok {
		return nil, errors.New("room not found")
	}

	return room, nil
}

func (f *fakeRoomRepository) List(ctx context.Context) ([]domain.Room, error) {
	f.listCalled = true

	if f.listError != nil {
		return nil, f.listError
	}

	rooms := make([]domain.Room, 0, len(f.rooms))
	for _, room := range f.rooms {
		rooms = append(rooms, *room)
	}

	return rooms, nil
}

func (f *fakeRoomRepository) Update(ctx context.Context, room *domain.Room) error {
	f.updateCalled = true

	if f.updateError != nil {
		return f.updateError
	}

	if f.rooms == nil {
		return errors.New("room not found")
	}

	existingRoom, ok := f.rooms[room.ID]
	if !ok {
		return errors.New("room not found")
	}

	existingRoom.BuildingID = room.BuildingID
	existingRoom.Number = room.Number
	existingRoom.Capacity = room.Capacity
	existingRoom.Description = room.Description

	return nil
}

func (f *fakeRoomRepository) SearchByCapacity(ctx context.Context, minCapacity int32) ([]domain.Room, error) {
	f.searchByCapacityCalled = true

	if f.searchError != nil {
		return nil, f.searchError
	}

	var result []domain.Room
	for _, room := range f.rooms {
		if room.Capacity >= minCapacity {
			result = append(result, *room)
		}
	}

	return result, nil
}

func TestCreateRoomSuccess(t *testing.T) {
	repo := &fakeRoomRepository{}
	uc := NewRoomUsecase(repo)

	room := &domain.Room{
		Number:      "A-101",
		Capacity:    30,
		BuildingID:  1,
		Description: "Lecture room",
	}

	err := uc.CreateRoom(context.Background(), room)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !repo.createCalled {
		t.Error("expected repository Create to be called")
	}

	if room.ID == 0 {
		t.Error("expected room ID to be assigned")
	}

	if repo.rooms[room.ID].Number != "A-101" {
		t.Errorf("expected room number A-101, got %s", repo.rooms[room.ID].Number)
	}
}

func TestCreateRoomRepositoryError(t *testing.T) {
	repo := &fakeRoomRepository{
		createError: errors.New("database error"),
	}
	uc := NewRoomUsecase(repo)

	room := &domain.Room{
		Number:      "B-202",
		Capacity:    20,
		BuildingID:  2,
		Description: "Lab room",
	}

	err := uc.CreateRoom(context.Background(), room)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !repo.createCalled {
		t.Error("expected repository Create to be called")
	}
}

func TestGetRoomByIDSuccess(t *testing.T) {
	repo := &fakeRoomRepository{
		rooms: map[uint]*domain.Room{
			1: {
				ID:          1,
				Number:      "A-101",
				Capacity:    30,
				BuildingID:  1,
				Description: "Lecture room",
			},
		},
	}
	uc := NewRoomUsecase(repo)

	room, err := uc.GetRoomByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if room == nil {
		t.Fatal("expected room, got nil")
	}

	if room.ID != 1 {
		t.Errorf("expected room ID 1, got %d", room.ID)
	}

	if room.Number != "A-101" {
		t.Errorf("expected room number A-101, got %s", room.Number)
	}

	if !repo.getByIDCalled {
		t.Error("expected repository GetByID to be called")
	}
}

func TestListRoomsSuccess(t *testing.T) {
	repo := &fakeRoomRepository{
		rooms: map[uint]*domain.Room{
			1: {
				ID:          1,
				Number:      "A-101",
				Capacity:    30,
				BuildingID:  1,
				Description: "Lecture room",
			},
			2: {
				ID:          2,
				Number:      "B-202",
				Capacity:    15,
				BuildingID:  2,
				Description: "Lab room",
			},
		},
	}
	uc := NewRoomUsecase(repo)

	rooms, err := uc.ListRooms(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(rooms))
	}

	if !repo.listCalled {
		t.Error("expected repository List to be called")
	}
}

func TestUpdateRoomSuccess(t *testing.T) {
	repo := &fakeRoomRepository{
		rooms: map[uint]*domain.Room{
			1: {
				ID:          1,
				Number:      "A-101",
				Capacity:    30,
				BuildingID:  1,
				Description: "Lecture room",
			},
		},
	}
	uc := NewRoomUsecase(repo)

	updatedRoom := &domain.Room{
		ID:          1,
		Number:      "A-102",
		Capacity:    40,
		BuildingID:  1,
		Description: "Updated lecture room",
	}

	err := uc.UpdateRoom(context.Background(), updatedRoom)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !repo.updateCalled {
		t.Error("expected repository Update to be called")
	}

	room := repo.rooms[1]

	if room.Number != "A-102" {
		t.Errorf("expected room number A-102, got %s", room.Number)
	}

	if room.Capacity != 40 {
		t.Errorf("expected capacity 40, got %d", room.Capacity)
	}

	if room.Description != "Updated lecture room" {
		t.Errorf("expected updated description, got %s", room.Description)
	}
}

func TestSearchRoomsByCapacitySuccess(t *testing.T) {
	repo := &fakeRoomRepository{
		rooms: map[uint]*domain.Room{
			1: {
				ID:          1,
				Number:      "A-101",
				Capacity:    30,
				BuildingID:  1,
				Description: "Lecture room",
			},
			2: {
				ID:          2,
				Number:      "B-202",
				Capacity:    15,
				BuildingID:  2,
				Description: "Small room",
			},
			3: {
				ID:          3,
				Number:      "C-303",
				Capacity:    50,
				BuildingID:  3,
				Description: "Conference room",
			},
		},
	}
	uc := NewRoomUsecase(repo)

	rooms, err := uc.SearchRoomsByCapacity(context.Background(), 30)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rooms) != 2 {
		t.Errorf("expected 2 rooms with capacity >= 30, got %d", len(rooms))
	}

	if !repo.searchByCapacityCalled {
		t.Error("expected repository SearchByCapacity to be called")
	}
}
