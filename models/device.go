package models

import (
	"github.com/google/uuid"
)

type GPSDevice struct {
	Base
	Imei        *string       `gorm:"unique;index" json:"imei"`
	Password    *string       `json:"password"`
	Tracking    *bool         `gorm:"default:true;not null" json:"tracking"`
	APICookie   *string       `json:"api_cookie"`
	Number      *string       `gorm:"unique;index" json:"number"`
	Locations   []GPSLocation `gorm:"foreignKey:DeviceID" json:"locations"`
	CreatedByID *uuid.UUID    `gorm:"index" json:"created_by"`
	CreatedBY   *User         `json:"-"`
}

type GPSLocation struct {
	Base
	Latitude  float64    `gorm:"not null" json:"latitude"`
	Longitude float64    `gorm:"not null" json:"longitude"`
	DeviceID  uuid.UUID  `gorm:"index" json:"device_id"`
	Device    *GPSDevice `json:"-"`
}
