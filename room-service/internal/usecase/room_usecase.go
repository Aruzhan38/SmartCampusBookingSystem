package usecase

import (
	"context"
	"room-service/internal/domain"
	"room-service/internal/repository"
)

type RoomUsecase interface {
	CreateRoom(ctx context.Context, room *domain.Room) error
	GetRoomByID(ctx context.Context, id uint) (*domain.Room, error)
	ListRooms(ctx context.Context) ([]domain.Room, error)
	UpdateRoom(ctx context.Context, room *domain.Room) error
	SearchRoomsByCapacity(ctx context.Context, minCapacity int32) ([]domain.Room, error)
}

type roomUsecase struct {
	repo repository.RoomRepository
}

func NewRoomUsecase(repo repository.RoomRepository) RoomUsecase {
	return &roomUsecase{repo: repo}
}

func (u *roomUsecase) CreateRoom(ctx context.Context, room *domain.Room) error {
	return u.repo.Create(ctx, room)
}

func (u *roomUsecase) GetRoomByID(ctx context.Context, id uint) (*domain.Room, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *roomUsecase) ListRooms(ctx context.Context) ([]domain.Room, error) {
	return u.repo.List(ctx)
}

func (u *roomUsecase) UpdateRoom(ctx context.Context, room *domain.Room) error {
	return u.repo.Update(ctx, room)
}

func (u *roomUsecase) SearchRoomsByCapacity(ctx context.Context, minCapacity int32) ([]domain.Room, error) {
	return u.repo.SearchByCapacity(ctx, minCapacity)
}
