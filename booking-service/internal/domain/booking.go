package domain

import "time"

type Booking struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	RoomID    uint      `gorm:"column:room_id" json:"room_id"`
	UserID    uint      `gorm:"column:user_id" json:"user_id"`
	StartTime time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime   time.Time `gorm:"column:end_time" json:"end_time"`
	Purpose   string    `gorm:"column:purpose" json:"purpose"`
	Status    string    `gorm:"column:status" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Booking) TableName() string {
	return "bookings"
}
