package models

import (
	"time"
)

// UserDevice represents a user's subscribed device for push notifications
type UserDevice struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index;not null" json:"user_id"`         
	PlayerID  string    `gorm:"uniqueIndex;not null" json:"player_id"` 
	Platform  string    `gorm:"type:varchar(50)" json:"platform"`      // web, ios, android
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
