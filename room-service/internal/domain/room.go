package domain

import "time"

type Room struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Number      string    `json:"number"`
	Capacity    int32     `json:"capacity"`
	BuildingID  uint      `json:"building_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
