package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Community struct {
	Base
	Name            string        `gorm:"not null;unique" json:"name"`
	Description     *string       `json:"description"`
	Type            CommunityType `gorm:"type:community_type;default:'public'" json:"type"`
	AppearsInSearch bool          `gorm:"not null;default:true" json:"appears_in_search"`
	Admin           *User         `json:"admin"`
	AdminID         uuid.UUID     `gorm:"not null" json:"admin_id"`
	Members         []User        `gorm:"many2many:community_members" json:"members"`
	Events          []Event       `gorm:"many2many:event_communities" json:"events"`
	PolygonArea     string        `gorm:"type:GEOMETRY(POLYGON,4326);not null" json:"polygon_area"`
	TrackingDevices []GPSDevice   `gorm:"many2many:community_tracking" json:"tracking_devices"`
}

type CommunityInvite struct {
	Base
	Community   Community `json:"-"`
	CommunityID uuid.UUID `gorm:"not null" json:"community_id"`
	User        User      `json:"-"`
	UserID      uuid.UUID `gorm:"not null" json:"user_id"`
	Accepted    *bool     `json:"accepted"`
	Creator     User      `json:"-"`
	CreatorID   uuid.UUID `gorm:"not null" json:"creator_id"`
}

type CommunityType string

const (
	PUBLIC  CommunityType = "public"
	PRIVATE CommunityType = "private"
)

func (ct *CommunityType) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*ct = CommunityType(v) // value is already a string
	case []byte:
		*ct = CommunityType(string(v)) // value is a byte slice, convert to string
	default:
		return fmt.Errorf("unsupported scan type for CommunityType: %T", value)
	}
	return nil
}

func (ct CommunityType) Value() (driver.Value, error) {
	return string(ct), nil
}

func (c *Community) IsMember(user *User) bool {
	for _, member := range c.Members {
		if member.ID == user.ID {
			return true
		}
	}

	return false
}

func (c *Community) IsDeviceTracked(device *GPSDevice) bool {
	for _, d := range c.TrackingDevices {
		if d.ID == device.ID {
			return true
		}
	}

	return false
}

func (c *Community) IsEventLinked(event *Event) bool {
	for _, e := range c.Events {
		if e.ID == event.ID {
			return true
		}
	}

	return false
}

func (community *Community) Fetch(db *gorm.DB, id string) error {
	return db.Select("communities.created_at, communities.deleted_at, communities.updated_at, communities.id, communities.name, communities.description, ST_AsText(polygon_area) AS polygon_area, type, admin_id").Preload("Members", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Preload("Events", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Preload("TrackingDevices", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Joins("Admin").Where("communities.id = ?", id).First(&community).Error
}
