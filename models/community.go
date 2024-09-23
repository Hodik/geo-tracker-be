package models

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommunityMember struct {
	CommunityID uuid.UUID  `gorm:"primaryKey" json:"-"`
	Community   Community  `json:"-"`
	UserID      uuid.UUID  `gorm:"primaryKey" json:"-"`
	User        User       `json:"user"`
	Role        MemberRole `gorm:"type:member_role;default:'read_only'" json:"role"`
}

type Community struct {
	Base
	Name                          string            `gorm:"not null;unique" json:"name"`
	Description                   *string           `json:"description"`
	Type                          CommunityType     `gorm:"type:community_type;default:'public'" json:"type"`
	AppearsInSearch               bool              `gorm:"not null;default:true" json:"appears_in_search"`
	IncludeExternalEvents         bool              `gorm:"not null;default:true" json:"include_external_events"`
	AllowReadOnlyMembersAddEvents bool              `gorm:"not null;default:true" json:"allow_read_only_members_add_events"`
	Members                       []CommunityMember `json:"members"`
	Events                        []Event           `gorm:"many2many:event_communities" json:"events"`
	PolygonArea                   string            `gorm:"type:GEOMETRY(POLYGON,4326);not null" json:"polygon_area"`
	TrackingDevices               []GPSDevice       `gorm:"many2many:community_tracking" json:"tracking_devices"`
}

type CommunityInvite struct {
	Base
	Community   Community  `json:"-"`
	CommunityID uuid.UUID  `gorm:"not null" json:"community_id"`
	User        User       `json:"-"`
	UserID      uuid.UUID  `gorm:"not null" json:"user_id"`
	Accepted    *bool      `json:"accepted"`
	Creator     User       `json:"-"`
	CreatorID   uuid.UUID  `gorm:"not null" json:"creator_id"`
	Role        MemberRole `gorm:"type:member_role;default:'read_only'" json:"role"`
}

type CommunityType string
type MemberRole string

const (
	PUBLIC  CommunityType = "public"
	PRIVATE CommunityType = "private"
)

const (
	ADMIN     MemberRole = "admin"
	READ_ONLY MemberRole = "read_only"
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

func (mr *MemberRole) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*mr = MemberRole(v) // value is already a string
	case []byte:
		*mr = MemberRole(string(v)) // value is a byte slice, convert to string
	default:
		return fmt.Errorf("unsupported scan type for MemberRole: %T", value)
	}
	return nil
}

func (ct CommunityType) Value() (driver.Value, error) {
	return string(ct), nil
}

func (mr MemberRole) Value() (driver.Value, error) {
	return string(mr), nil
}

func (c *Community) LoadMembers(db *gorm.DB) error {
	var communityMembers []CommunityMember
	if err := db.Where("community_id = ?", c.ID).Joins("User").
		Find(&communityMembers).Error; err != nil {
		return nil
	}
	c.Members = communityMembers
	return nil
}

func (c *Community) IsMember(user *User) bool {
	for _, member := range c.Members {
		if member.UserID == user.ID {
			return true
		}
	}

	return false
}

func (c *Community) IsAdmin(user *User) bool {
	for _, member := range c.Members {
		if member.UserID == user.ID && member.Role == ADMIN {
			return true
		}
	}

	return false
}

func (c *Community) AdminsCount() int {
	i := 0
	for _, member := range c.Members {
		if member.Role == ADMIN {
			i++
		}
	}

	return i
}

func (c *Community) AddMember(db *gorm.DB, user *User, role MemberRole) error {
	newMember := CommunityMember{User: *user, Community: *c, Role: role}
	if err := db.Create(&newMember).Error; err != nil {
		return err
	}

	if c.Members != nil {
		c.Members = append(c.Members, newMember)
	} else {
		return c.LoadMembers(db)
	}
	return nil
}

func (c *Community) RemoveMember(db *gorm.DB, user *User) error {
	if err := db.Unscoped().Where("user_id = ? AND community_id = ?", user.ID, c.ID).Delete(&CommunityMember{}).Error; err != nil {
		return err
	}

	return c.LoadMembers(db) // reload members
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
	if err := db.Select("communities.created_at, communities.deleted_at, communities.updated_at, communities.id, communities.name, communities.description, ST_AsText(polygon_area) AS polygon_area, type").Preload("Events", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Preload("TrackingDevices", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Where("communities.id = ?", id).First(&community).Error; err != nil {
		return nil
	}

	return community.LoadMembers(db)
}

func ValidateCommunityType(ct string) error {
	if ct != string(PUBLIC) && ct != string(PRIVATE) {
		return errors.New("invalid community type")
	}

	return nil
}
