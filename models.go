package main

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

type GPSDevice struct {
	Base
	Imei      string        `gorm:"unique;not null;index" json:"imei"`
	Password  string        `gorm:"not null" json:"password"`
	APICookie *string       `json:"api_cookie"`
	Number    *string       `gorm:"unique;index" json:"number"`
	Locations []GPSLocation `gorm:"foreignKey:DeviceID" json:"locations"`
}

type GPSLocation struct {
	Base
	Latitude  float64 `gorm:"not null" json:"latitude"`
	Longitude float64 `gorm:"not null" json:"longitude"`
	DeviceID  uint    `gorm:"index" json:"device_id"`
	Device    *GPSDevice
}
