package repository

import (
	"context"
	"room-service/internal/domain"

	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(ctx context.Context, room *domain.Room) error
	GetByID(ctx context.Context, id uint) (*domain.Room, error)
	List(ctx context.Context) ([]domain.Room, error)
	Update(ctx context.Context, room *domain.Room) error
	SearchByCapacity(ctx context.Context, minCapacity int32) ([]domain.Room, error)
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(ctx context.Context, room *domain.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}

func (r *roomRepository) GetByID(ctx context.Context, id uint) (*domain.Room, error) {
	var room domain.Room
	if err := r.db.WithContext(ctx).First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) List(ctx context.Context) ([]domain.Room, error) {
	var rooms []domain.Room
	err := r.db.WithContext(ctx).Order("id asc").Find(&rooms).Error
	return rooms, err
}

func (r *roomRepository) Update(ctx context.Context, room *domain.Room) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Room{}).
		Where("id = ?", room.ID).
		Updates(map[string]interface{}{
			"building_id": room.BuildingID,
			"number":      room.Number,
			"capacity":    room.Capacity,
			"description": room.Description,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *roomRepository) SearchByCapacity(ctx context.Context, minCapacity int32) ([]domain.Room, error) {
	var rooms []domain.Room
	err := r.db.WithContext(ctx).
		Where("capacity >= ?", minCapacity).
		Order("capacity asc").
		Find(&rooms).Error
	return rooms, err
}
