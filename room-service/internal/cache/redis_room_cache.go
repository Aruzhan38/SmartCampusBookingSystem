package cache

import (
	"context"
	"encoding/json"
	"time"

	"room-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

const roomsCacheKey = "rooms:list"

type RedisRoomCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRoomCache(addr string) *RedisRoomCache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisRoomCache{
		client: client,
		ttl:    5 * time.Minute,
	}
}

func (c *RedisRoomCache) GetRooms(ctx context.Context) ([]domain.Room, bool, error) {
	data, err := c.client.Get(ctx, roomsCacheKey).Result()
	if err == redis.Nil {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	var rooms []domain.Room
	if err := json.Unmarshal([]byte(data), &rooms); err != nil {
		return nil, false, err
	}

	return rooms, true, nil
}

func (c *RedisRoomCache) SetRooms(ctx context.Context, rooms []domain.Room) error {
	data, err := json.Marshal(rooms)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, roomsCacheKey, data, c.ttl).Err()
}

func (c *RedisRoomCache) DeleteRooms(ctx context.Context) error {
	return c.client.Del(ctx, roomsCacheKey).Err()
}
