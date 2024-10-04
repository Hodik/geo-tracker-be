package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"log"

	"github.com/Hodik/geo-tracker-be/dbconn"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Event struct {
	Base
	Title              string            `gorm:"not null" json:"title"`
	Description        string            `json:"description"`
	Type               EventType         `gorm:"type:event_type;not null;default:'other'" json:"type"`
	Status             EventStatus       `gorm:"type:event_status;not null;default:'open'" json:"status"`
	IsPublic           *bool             `gorm:"default:true;not null" json:"is_public"`
	DeviceID           *uuid.UUID        `gorm:"index" json:"device_id"`
	Device             *GPSDevice        `json:"device"`
	Latitude           float64           `gorm:"not null" json:"latitude"`
	Longitude          float64           `gorm:"not null" json:"longitude"`
	CreatedBy          *User             `json:"-"`
	CreatedByID        uuid.UUID         `gorm:"not null;index" json:"created_by_id"`
	LinkedEvents       []*Event          `gorm:"many2many:event_linked" json:"-"`
	Communities        []*Community      `gorm:"many2many:event_communities" json:"-"`
	AreasOfInterest    []*AreaOfInterest `gorm:"many2many:event_areas_of_interest" json:"-"`
	Comments           []*Comment        `json:"comments"`
	MediaFiles         []*MediaFile      `gorm:"many2many:event_media_files" json:"-"`
	MediaPresignedUrls []*PresignedUrl   `gorm:"-" json:"media_presigned_urls"`
}

type Comment struct {
	Base
	Content     string    `gorm:"not null" json:"content"`
	CreatedBy   *User     `json:"-"`
	CreatedByID uuid.UUID `gorm:"not null;index" json:"created_by_id"`
	Event       Event     `json:"-"`
	EventID     uuid.UUID `gorm:"not null;index" json:"event_id"`
}

type EventType string
type EventStatus string

const (
	EventTypeRobbery  EventType = "robbery"
	EventTypeLost     EventType = "lost"
	EventTypeAccident EventType = "accident"
	EventTypeOther    EventType = "other"
)

const (
	EventStatusOpen     EventStatus = "open"
	EventStatusResolved EventStatus = "resolved"
	EventStatusClosed   EventStatus = "closed"
)

func ValidateEventType(et string) error {
	if et != string(EventTypeRobbery) && et != string(EventTypeLost) && et != string(EventTypeAccident) && et != string(EventTypeOther) {
		return errors.New("invalid event type")
	}

	return nil
}

func ValidateEventStatus(es string) error {
	if es != string(EventStatusOpen) && es != string(EventStatusResolved) && es != string(EventStatusClosed) {
		return errors.New("invalid event status")
	}

	return nil
}

func (et EventType) Value() (driver.Value, error) {
	return string(et), nil
}

func (es EventStatus) Value() (driver.Value, error) {
	return string(es), nil
}

func (et *EventType) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*et = EventType(v)
	case []byte:
		*et = EventType(string(v))
	default:
		return fmt.Errorf("unsupported scan type for EventType: %T", value)
	}
	return nil
}

func (es *EventStatus) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*es = EventStatus(v)
	case []byte:
		*es = EventStatus(string(v))
	default:
		return fmt.Errorf("unsupported scan type for EventStatus: %T", value)
	}
	return nil
}

func (e *Event) HasAccess(db *gorm.DB, user *User, writer bool) (bool, error) {

	if *e.IsPublic {
		return true, nil
	}

	if e.CreatedByID == user.ID {
		return true, nil
	}

	var count int64
	tx := db.Table("event_communities").Joins("INNER JOIN community_members ON event_communities.community_id = community_members.community_id")
	if !writer {
		tx = tx.Where("event_communities.event_id = ? AND community_members.user_id = ?", e.ID, user.ID)
	} else if writer {
		tx = tx.Where("event_communities.event_id = ? AND community_members.user_id = ? AND community_members.role = ?", e.ID, user.ID, "admin")
	}

	if err := tx.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (e *Event) AfterSave(db *gorm.DB) (err error) {
	if *e.IsPublic {
		go e.Populate()
	}
	return nil
}

func (e *Event) Populate() (err error) {
	db := dbconn.GetDB()
	if err := e.PopulateToAreasOfInterest(db); err != nil {
		log.Println("Error populating event to areas of interest:", err)
	}
	return nil
}

func (e *Event) PopulateToAreasOfInterest(db *gorm.DB) (err error) {
	var areasOfInterest []AreaOfInterest
	if err := db.Where("ST_Intersects(polygon_area, " + e.GetPoint() + ")").Find(&areasOfInterest).Error; err != nil {
		return err
	}

	if len(areasOfInterest) == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, aoi := range areasOfInterest {
			if err := tx.Session(&gorm.Session{SkipHooks: true}).Clauses(clause.OnConflict{DoNothing: true}).Model(&aoi).Association("Events").Append(e); err != nil {
				return err
			}
		}
		return nil
	})
}

func (e *Event) GetPoint() string {
	return fmt.Sprintf("ST_SetSRID(ST_MakePoint(%f, %f), 4326)", e.Longitude, e.Latitude)
}

func (c *Event) GetMediaKey(filename string) string {
	return fmt.Sprintf("events/%s/%s", c.ID, filename)
}
