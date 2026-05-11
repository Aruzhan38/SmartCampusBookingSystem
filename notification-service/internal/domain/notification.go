package domain

import "time"

type Notification struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID    uint      `gorm:"column:user_id" json:"user_id"`
	Message   string    `gorm:"column:message" json:"message"`
	Type      string    `gorm:"column:type" json:"type"`
	IsRead    bool      `gorm:"column:is_read" json:"is_read"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
