package main

import (
	"errors"
	"time"

	"github.com/Hodik/geo-tracker-be/auth"
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
	Imei        string        `gorm:"unique;not null;index" json:"imei"`
	Password    string        `gorm:"not null" json:"password"`
	Tracking    *bool         `gorm:"default:true;not null" json:"tracking"`
	APICookie   *string       `json:"api_cookie"`
	Number      *string       `gorm:"unique;index" json:"number"`
	Locations   []GPSLocation `gorm:"foreignKey:DeviceID" json:"locations"`
	CreatedByID *uint         `gorm:"index" json:"created_by"`
	CreatedBY   *auth.User    `json:"-"`
}

type GPSLocation struct {
	Base
	Latitude  float64    `gorm:"not null" json:"latitude"`
	Longitude float64    `gorm:"not null" json:"longitude"`
	DeviceID  uint       `gorm:"index" json:"device_id"`
	Device    *GPSDevice `json:"-"`
}

type UserSettings struct {
	Base
	User   auth.User
	UserID uint

	TrackingDevices []GPSDevice `gorm:"many2many:user_tracking"`
}

type Config struct {
	PollInterval uint8  `gorm:"default:30" json:"poll_interval"`
	Dummy        string `gorm:"unique;default:'singleton'" json:"-"`
}

func GetUserSettings(user *auth.User) (*UserSettings, error) {
	var userSettings UserSettings
	settingsResult := db.Preload("TrackingDevices").First(&userSettings)

	if errors.Is(settingsResult.Error, gorm.ErrRecordNotFound) {
		userSettings = UserSettings{UserID: user.ID}
		createResult := db.Create(&userSettings)

		if createResult.Error != nil {
			return nil, createResult.Error
		}
	}

	return &userSettings, nil
}
