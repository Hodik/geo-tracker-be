package models

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Base
	Email string `gorm:"unique;not null" json:"email"`

	Name          *string `json:"name"`
	EmailVerified *bool   `json:"email_verified"`

	Communities []Community `gorm:"many2many:community_members" json:"communities"`
}

type UserSettings struct {
	Base
	User   User      `json:"-"`
	UserID uuid.UUID `json:"-"`

	TrackingDevices []GPSDevice `gorm:"many2many:user_tracking" json:"tracking_devices"`
}

type Notification struct {
	Base
	Message string    `gorm:"not null" json:"message"`
	User    *User     `json:"-"`
	UserID  uuid.UUID `gorm:"not null" json:"user_id"`
	Event   Event     `json:"-"`
	EventID uuid.UUID `gorm:"not null" json:"event_id"`
	IsRead  bool      `gorm:"not null;default:false"`
}

type AreaOfInterest struct {
	Base
	User        *User     `json:"-"`
	UserID      uuid.UUID `gorm:"not null" json:"user_id"`
	PolygonArea string    `gorm:"type:GEOMETRY(POLYGON,4326);not null" json:"polygon_area"`
}

func GetUserSettings(db *gorm.DB, user *User) (*UserSettings, error) {
	var userSettings UserSettings
	settingsResult := db.Preload("TrackingDevices").Where("user_id = ?", user.ID).First(&userSettings)

	if errors.Is(settingsResult.Error, gorm.ErrRecordNotFound) {
		userSettings = UserSettings{UserID: user.ID}
		createResult := db.Create(&userSettings)

		if createResult.Error != nil {
			return nil, createResult.Error
		}
	}

	return &userSettings, nil
}

func (us *UserSettings) IsDeviceTracked(device *GPSDevice) bool {
	for _, d := range us.TrackingDevices {
		if d.ID == device.ID {
			return true
		}
	}

	return false
}
