package models

import (
	"github.com/google/uuid"
)

type Event struct {
	Base
	Title        string    `gorm:"not null" json:"title"`
	Description  string    `json:"description"`
	Type         string    `json:"type"`
	Status       string    `json:"string"`
	IsPublic     bool      `gorm:"default:true;not null" json:"is_public"`
	Latitude     float64   `gorm:"not null" json:"latitude"`
	Longitude    float64   `gorm:"not null" json:"longitude"`
	CreatedBy    *User     `json:"-"`
	CreatedByID  uuid.UUID `gorm:"not null" json:"created_by_id"`
	LinkedEvents []Event   `gorm:"many2many:event_linked"`
}

type Comment struct {
	Base
	Content     string    `gorm:"not null" json:"content"`
	CreatedBy   *User     `json:"-"`
	CreatedByID uuid.UUID `gorm:"not null" json:"created_by_id"`
	Event       Event     `json:"-"`
	EventID     uuid.UUID `gorm:"not null" json:"event_id"`
}
