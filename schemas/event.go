package schemas

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateEvent struct {
	Title       string              `json:"title" binding:"required"`
	Description string              `json:"description" binding:"required"`
	IsPublic    bool                `json:"is_public"`
	Type        *models.EventType   `json:"type"`
	Status      *models.EventStatus `json:"status"`
	Latitude    *float64            `json:"latitude"`
	Longitude   *float64            `json:"longitude"`
	DeviceID    *uuid.UUID          `json:"device_id"`
	Communities *[]uuid.UUID        `json:"communities"`
}

type UpdateEvent struct {
	Title       *string             `json:"title"`
	Description *string             `json:"description"`
	Status      *models.EventStatus `json:"status"`
	Latitude    *float64            `json:"latitude"`
	Longitude   *float64            `json:"longitude"`
}

type CreateComment struct {
	Content string `json:"content" binding:"required"`
}

type UpdateComment struct {
	Content *string `json:"content"`
}

func (c *CreateComment) ToComment(user *models.User, event *models.Event) *models.Comment {
	return &models.Comment{Content: c.Content, CreatedBy: user, Event: *event, CreatedByID: user.ID, EventID: event.ID}
}

func (u *UpdateComment) ToComment(user *models.User, existing *models.Comment) error {

	if user.ID != existing.CreatedByID {
		return errors.New("user does not have permission to modify this comment")
	}
	if u.Content != nil {
		existing.Content = *u.Content
	}
	return nil
}

func (c *CreateEvent) ToEvent(db *gorm.DB, user *models.User) (*models.Event, error) {
	var EventType models.EventType = models.EventTypeOther
	var EventStatus models.EventStatus = models.EventStatusOpen

	if c.Type != nil {
		if err := models.ValidateEventType(string(*c.Type)); err != nil {
			return nil, err
		}
		EventType = *c.Type
	}

	if c.Status != nil {
		if err := models.ValidateEventStatus(string(*c.Status)); err != nil {
			return nil, err
		}
		EventStatus = *c.Status
	}

	var Latitude float64
	var Longitude float64

	if c.Latitude != nil && c.Longitude != nil {
		Latitude = *c.Latitude
		Longitude = *c.Longitude
	} else if c.DeviceID != nil {
		var latestLocation models.GPSLocation

		if err := db.Where("device_id = ?", c.DeviceID).Order("created_at DESC").First(&latestLocation).Error; err != nil {
			return nil, err
		}

		Latitude = latestLocation.Latitude
		Longitude = latestLocation.Longitude
	} else {
		return nil, errors.New("latitude and longitude or device id is required")
	}

	var communities []*models.Community

	if c.Communities != nil {
		communityIDs := make([]string, len(*c.Communities))
		for i, id := range *c.Communities {
			communityIDs[i] = id.String()
		}
		if err := db.Where("id IN (?)", communityIDs).Find(&communities).Error; err != nil {
			return nil, err
		}

		for _, community := range communities {
			if err := community.LoadMembers(db); err != nil { // TODO: Optimize this to load members in one query
				return nil, err
			}

			if !community.CanAddEvent(user, c.IsPublic) {
				return nil, errors.New("cannot add event to community " + community.Name)
			}
		}
	}

	return &models.Event{
		CreatedByID: user.ID,
		DeviceID:    c.DeviceID,
		Latitude:    Latitude,
		Longitude:   Longitude,
		Title:       c.Title,
		Description: c.Description,
		IsPublic:    &c.IsPublic,
		Type:        EventType,
		Status:      EventStatus,
		Communities: communities,
	}, nil
}

func (u *UpdateEvent) ToEvent(db *gorm.DB, user *models.User, existing *models.Event) error {

	if u.Title != nil {
		existing.Title = *u.Title
	}

	if u.Description != nil {
		existing.Description = *u.Description
	}

	if u.Status != nil {
		if err := models.ValidateEventStatus(string(*u.Status)); err != nil {
			return err
		}
		existing.Status = *u.Status
	}

	if u.Latitude != nil && u.Longitude != nil {
		existing.Latitude = *u.Latitude
		existing.Longitude = *u.Longitude
	}

	return nil
}
