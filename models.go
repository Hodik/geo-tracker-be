package main

import (
	"database/sql/driver"
	"errors"
	"log"
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
	Imei        *string       `gorm:"unique;index" json:"imei"`
	Password    *string       `json:"password"`
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

type Migration struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Status bool   `gorm:"default:true" json:"status"`
	Dummy  string `gorm:"unique;default:'singleton'" json:"-"`
}

type Event struct {
	Base
	Title        string    `gorm:"not null" json:"title"`
	Description  string    `json:"description"`
	Type         string    `json:"type"`
	Status       string    `json:"string"`
	IsPublic     bool      `gorm:"default:true;not null" json:"is_public"`
	Latitude     float64   `gorm:"not null" json:"latitude"`
	Longitude    float64   `gorm:"not null" json:"longitude"`
	CreatedBy    auth.User `json:"-"`
	CreatedByID  uint      `gorm:"not null" json:"created_by_id"`
	LinkedEvents []Event   `gorm:"many2many:event_linked"`
}

type Comment struct {
	Base
	Content     string    `gorm:"not null" json:"content"`
	CreatedBy   auth.User `json:"-"`
	CreatedByID uint      `gorm:"not null" json:"created_by_id"`
	Event       Event     `json:"-"`
	EventID     uint      `gorm:"not null" json:"event_id"`
}

type Notification struct {
	Base
	Message string    `gorm:"not null" json:"message"`
	User    auth.User `json:"-"`
	UserID  uint      `gorm:"not null" json:"user_id"`
	Event   Event     `json:"-"`
	EventID uint      `gorm:"not null" json:"event_id"`
	IsRead  bool      `gorm:"not null;default:false"`
}

type Community struct {
	Base
	Name        string        `gorm:"not null;unique" json:"name"`
	Description string        `json:"description"`
	Type        CommunityType `gorm:"type:community_type;default:'public'"`
	Admin       auth.User     `json:"-"`
	AdminID     uint          `gorm:"not null" json:"admin_id"`
	Members     []auth.User   `gorm:"many2many:community_members" json:"-"`
	Events      []Event       `gorm:"many2many:event_communities" json:"-"`
	PolygonArea string        `gorm:"type:GEOMETRY(POLYGON,4326);not null" json:"polygon_area"`
}

type AreaOfInterest struct {
	Base
	User        auth.User `json:"-"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	PolygonArea string    `gorm:"type:GEOMETRY(POLYGON,4326);not null" json:"polygon_area"`
}

type CommunityType string

const (
	PUBLIC  CommunityType = "public"
	PRIVATE CommunityType = "private"
)

func (ct *CommunityType) Scan(value interface{}) error {
	*ct = CommunityType(value.([]byte))
	return nil
}

func (ct CommunityType) Value() (driver.Value, error) {
	return string(ct), nil
}

func CreateCommunityTypeEnum() error {
	result := db.Exec("SELECT 1 FROM pg_type WHERE typname = 'community_type';")

	switch {
	case result.RowsAffected == 0:
		if err := db.Exec("CREATE TYPE community_type AS ENUM ('public', 'private');").Error; err != nil {
			log.Println("Error creating community_type")
			return err
		}

		return nil
	case result.Error != nil:
		return result.Error

	default:
		return nil
	}
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
