package models

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Base
	Email string `gorm:"unique;not null;index" json:"email"`

	Name          *string `json:"name"`
	EmailVerified *bool   `json:"email_verified"`

	AreasOfInterest []*AreaOfInterest `gorm:"many2many:user_areas_of_interest" json:"areas_of_interest"`
}

type UserSettings struct {
	Base
	User   User      `json:"-"`
	UserID uuid.UUID `gorm:"not null;index" json:"-"`

	TrackingDevices []*GPSDevice `gorm:"many2many:user_tracking" json:"tracking_devices"`
}

type Notification struct {
	Base
	Message string    `gorm:"not null" json:"message"`
	User    *User     `json:"-"`
	UserID  uuid.UUID `gorm:"not null;index" json:"user_id"`
	Event   *Event    `json:"-"`
	EventID uuid.UUID `gorm:"not null;index" json:"event_id"`
	IsRead  bool      `gorm:"not null;default:false"`
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

func (u *User) FetchCommunities(db *gorm.DB) ([]Community, error) {
	var communityMembers []CommunityMember
	if err := db.Where("user_id = ?", u.ID).Joins("Community").
		Find(&communityMembers).Error; err != nil {
		return nil, err
	}

	var communities []Community
	for _, cm := range communityMembers {
		communities = append(communities, cm.Community)
	}

	return communities, nil
}
