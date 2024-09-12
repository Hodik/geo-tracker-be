package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Hodik/geo-tracker-be/auth"
)

type CreateAreaOfInterest struct {
	PolygonArea string `json:"polygon_area" binding:"required"`
}

type CreateGPSDevice struct {
	Imei     *string `json:"imei"`
	Password *string `json:"password"`

	Number   *string `json:"number"`
	Tracking *bool   `json:"tracking"`
}

type UpdateGPSDevice struct {
	Number   *string `json:"number"`
	Imei     *string `json:"imei"`
	Tracking *bool   `json:"tracking"`
}

type UpdateUser struct {
	Name *string `json:"name"`
}

type UpdateSettings struct {
	TrackingDevices *[]uint `json:"tracking_devices"`
}

type UpdateMe struct {
	User     *UpdateUser     `json:"user"`
	Settings *UpdateSettings `json:"settings"`
}

type UserSettingsOut struct {
	TrackingDevices []GPSDevice `json:"tracking_devices"`
}

type UserProfile struct {
	User     *auth.User       `json:"user"`
	Settings *UserSettingsOut `json:"settings"`
}

type CreateCommunity struct {
	Name        string         `json:"name" binding:"required"`
	Description *string        `json:"description"`
	Type        *CommunityType `json:"type"`
	AdminID     *uint          `json:"admin_id"`
	PolygonArea string         `json:"polygon_area" binding:"required"`
}

type UpdateCommunity struct {
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	Type        *CommunityType `json:"type"`
	PolygonArea *string        `json:"polygon_area"`
	Members     *[]uint        `json:"members"`
	Events      *[]uint        `json:"events"`
}

func (c *CreateCommunity) ToCommunity(creator *auth.User) (*Community, error) {
	if err := ValidatePolygonWKT(c.PolygonArea); err != nil {
		return nil, err
	}

	if c.Type != nil {
		if *c.Type != PUBLIC && *c.Type != PRIVATE {
			return nil, errors.New("type should be private or public")
		}
	}

	var AdminId uint
	if c.AdminID != nil {
		AdminId = *c.AdminID
	} else {
		AdminId = creator.ID
	}

	var Type CommunityType = PUBLIC
	if c.Type != nil {
		Type = *c.Type
	}

	return &Community{Name: c.Name, Description: c.Description, Type: Type, AdminID: AdminId, PolygonArea: c.PolygonArea}, nil
}

func (u *UpdateCommunity) ToCommunity(existing *Community, user *auth.User) (*Community, error) {
	if user.ID != existing.AdminID {
		return nil, errors.New("only admin can modify community")
	}

	if u.PolygonArea != nil {
		if err := ValidatePolygonWKT(*u.PolygonArea); err != nil {
			return nil, err
		}
		existing.PolygonArea = *u.PolygonArea
	}

	if u.Name != nil {
		existing.Name = *u.Name
	}

	if u.Description != nil {
		existing.Description = u.Description
	}

	if u.Type != nil {
		if *u.Type != PUBLIC && *u.Type != PRIVATE {
			return nil, errors.New("type should be private or public")
		}

		existing.Type = *u.Type
	}

	if u.Members != nil {
		var members []auth.User
		if result := db.Where("id IN ?", *u.Members).Find(&members); result.Error != nil {
			return nil, result.Error
		}
		existing.Members = members
	}

	if u.Events != nil {
		var events []Event
		if result := db.Where("id IN ?", *u.Events).Find(&events); result.Error != nil {
			return nil, result.Error
		}
		existing.Events = events
	}

	return existing, nil
}

func (c *CreateAreaOfInterest) ToAreaOfInterest(creator *auth.User) (*AreaOfInterest, error) {
	if err := ValidatePolygonWKT(c.PolygonArea); err != nil {
		return nil, err
	}

	return &AreaOfInterest{UserID: creator.ID, PolygonArea: c.PolygonArea}, nil
}

func ValidatePolygonWKT(polygon string) error {
	// Basic format check (Starts with POLYGON and ends with closing parentheses)
	if !strings.HasPrefix(polygon, "POLYGON((") || !strings.HasSuffix(polygon, "))") {
		return errors.New("invalid polygon format: must start with 'POLYGON((' and end with '))'")
	}

	// Extract the coordinates part
	coordPart := strings.TrimPrefix(polygon, "POLYGON((")
	coordPart = strings.TrimSuffix(coordPart, "))")

	// Split the coordinates by comma
	points := strings.Split(coordPart, ",")

	if len(points) < 4 { // A valid polygon must have at least 4 points (including the closure point)
		return errors.New("invalid polygon: must have at least 4 points")
	}

	// Check if the first point matches the last (closure)
	if strings.TrimSpace(points[0]) != strings.TrimSpace(points[len(points)-1]) {
		return errors.New("invalid polygon: first and last points must be the same (closed polygon)")
	}

	// Validate each point's format
	coordRegex := regexp.MustCompile(`^[-+]?[0-9]*\.?[0-9]+ [-+]?[0-9]*\.?[0-9]+$`)
	for _, point := range points {
		point = strings.TrimSpace(point)
		if !coordRegex.MatchString(point) {
			return fmt.Errorf("invalid coordinate format: '%s'", point)
		}
	}

	return nil
}
func (u *UpdateMe) ToUser(existing *auth.User) *auth.User {
	if u.User == nil {
		return existing
	}

	if u.User.Name != nil {
		existing.Name = u.User.Name
	}

	return existing
}

func (u *UpdateMe) ToSettings(existing *UserSettings) (*UserSettings, error) {
	if u.Settings == nil {
		return existing, nil
	}

	if u.Settings.TrackingDevices != nil {
		var devices []GPSDevice
		if result := db.Where("id IN ?", *u.Settings.TrackingDevices).Find(&devices); result.Error != nil {
			return nil, result.Error
		}
		existing.TrackingDevices = devices
	}

	return existing, nil
}

func ToUserProfile(u *auth.User, settings *UserSettings) *UserProfile {
	return &UserProfile{User: u, Settings: settings.ToUserSettingsOut()}
}

func (settings *UserSettings) ToUserSettingsOut() *UserSettingsOut {
	return &UserSettingsOut{TrackingDevices: settings.TrackingDevices}
}

func (c *CreateGPSDevice) ToGPSDevice(creator *auth.User) *GPSDevice {
	if c.Imei == nil {
		*c.Tracking = false
	}

	d := &GPSDevice{
		Number:      c.Number,
		Imei:        c.Imei,
		Password:    c.Password,
		Tracking:    c.Tracking,
		CreatedByID: &creator.ID,
	}
	return d
}

func (u *UpdateGPSDevice) ToGPSDevice(existing *GPSDevice) *GPSDevice {
	if u.Number != nil {
		existing.Number = u.Number
	}

	if u.Imei != nil {
		existing.Imei = u.Imei
	}

	if u.Tracking != nil {
		existing.Tracking = u.Tracking
	}

	return existing
}
