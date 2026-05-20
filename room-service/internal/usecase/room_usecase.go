package usecase

import (
	"context"
	"log"

	"room-service/internal/domain"
	"room-service/internal/repository"
)

type RoomCache interface {
	GetRooms(ctx context.Context) ([]domain.Room, bool, error)
	SetRooms(ctx context.Context, rooms []domain.Room) error
	DeleteRooms(ctx context.Context) error
}

type RoomUsecase interface {
	CreateRoom(ctx context.Context, room *domain.Room) error
	GetRoomByID(ctx context.Context, id uint) (*domain.Room, error)
	ListRooms(ctx context.Context) ([]domain.Room, error)
	UpdateRoom(ctx context.Context, room *domain.Room) error
	SearchRoomsByCapacity(ctx context.Context, minCapacity int32) ([]domain.Room, error)
}

type roomUsecase struct {
	repo  repository.RoomRepository
	cache RoomCache
}

func NewRoomUsecase(repo repository.RoomRepository, cacheOpt ...RoomCache) RoomUsecase {
	var cache RoomCache
	if len(cacheOpt) > 0 {
		cache = cacheOpt[0]
	}

	return &roomUsecase{
		repo:  repo,
		cache: cache,
	}
}

func (u *roomUsecase) CreateRoom(ctx context.Context, room *domain.Room) error {
	if err := u.repo.Create(ctx, room); err != nil {
		return err
	}

	if u.cache != nil {
		if err := u.cache.DeleteRooms(ctx); err != nil {
			log.Println("failed to invalidate rooms cache:", err)
		} else {
			log.Println("rooms cache invalidated")
		}
	}

	return nil
}

func (u *roomUsecase) GetRoomByID(ctx context.Context, id uint) (*domain.Room, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *roomUsecase) ListRooms(ctx context.Context) ([]domain.Room, error) {
	if u.cache != nil {
		rooms, found, err := u.cache.GetRooms(ctx)
		if err != nil {
			log.Println("failed to get rooms from cache:", err)
		}

		if found {
			log.Println("rooms returned from Redis cache")
			return rooms, nil
		}
	}

	rooms, err := u.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	if u.cache != nil {
		if err := u.cache.SetRooms(ctx, rooms); err != nil {
			log.Println("failed to save rooms to cache:", err)
		} else {
			log.Println("rooms saved to Redis cache")
		}
	}

	return rooms, nil
}

func (u *roomUsecase) UpdateRoom(ctx context.Context, room *domain.Room) error {
	if err := u.repo.Update(ctx, room); err != nil {
		return err
	}

	if u.cache != nil {
		if err := u.cache.DeleteRooms(ctx); err != nil {
			log.Println("failed to invalidate rooms cache:", err)
		} else {
			log.Println("rooms cache invalidated")
		}
	}

	return nil
}

func (u *roomUsecase) SearchRoomsByCapacity(ctx context.Context, minCapacity int32) ([]domain.Room, error) {
	return u.repo.SearchByCapacity(ctx, minCapacity)
}
