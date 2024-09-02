package auth

import (
	"time"

	"gorm.io/gorm"
)

type Base struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type User struct {
	Base
	Email string `gorm:"unique;not null" json:"email"`

	Name          *string `json:"name"`
	EmailVerified *bool   `json:"email_verified"`
}
