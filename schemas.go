package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Hodik/geo-tracker-be/auth"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	TrackingDevices *[]uuid.UUID `json:"tracking_devices"`
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
	Name            string         `json:"name" binding:"required"`
	Description     *string        `json:"description"`
	Type            *CommunityType `json:"type"`
	AppearsInSearch *bool          `json:"appears_in_search"`
	AdminID         *uuid.UUID     `json:"admin_id"`
	PolygonArea     string         `json:"polygon_area" binding:"required"`
	TrackingDevices *[]uuid.UUID   `json:"tracking_devices"`
}

type UpdateCommunity struct {
	Name            *string        `json:"name"`
	Description     *string        `json:"description"`
	Type            *CommunityType `json:"type"`
	AppearsInSearch *bool          `json:"appears_in_search"`
	PolygonArea     *string        `json:"polygon_area"`
	Members         *[]uuid.UUID   `json:"members"`
	Events          *[]uuid.UUID   `json:"events"`
	TrackingDevices *[]uuid.UUID   `json:"tracking_devices"`
}

type CreateCommunityInvite struct {
	CommunityID uuid.UUID `json:"community_id" binding:"required"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`
}

type UpdateCommunityInvite struct {
	Accepted bool `json:"accepted" binding:"required"`
}

func (c *CreateCommunityInvite) ToCommunityInvite(creator *auth.User) (*CommunityInvite, error) {

	var community Community
	result := db.Where("communities.id = ?", c.CommunityID).Preload("Members").First(&community)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("community doesn't exist")
	}

	if result.Error != nil {
		return nil, errors.New(result.Error.Error())
	}

	var user auth.User

	result = db.Where("users.id = ?", c.UserID).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("user doesn't exist")
	}

	if result.Error != nil {
		return nil, errors.New(result.Error.Error())
	}

	if creator.ID != community.AdminID {
		return nil, errors.New("only community's admin can invite users")
	}

	if community.IsMember(&user) {
		return nil, errors.New("user is already a member of community")
	}

	return &CommunityInvite{UserID: c.UserID, CommunityID: c.CommunityID, CreatorID: creator.ID}, nil

}

func (u *UpdateCommunityInvite) ToCommunityInvite(existing *CommunityInvite, user *auth.User) error {
	if user.ID != existing.UserID {
		return errors.New("only invited user can accept")
	}

	existing.Accepted = &u.Accepted
	return nil
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

	var AdminId uuid.UUID
	if c.AdminID != nil {
		AdminId = *c.AdminID
	} else {
		AdminId = creator.ID
	}

	var Type CommunityType = PUBLIC
	if c.Type != nil {
		Type = *c.Type
	}

	var AppearsInSearch bool
	if c.AppearsInSearch != nil {
		AppearsInSearch = *c.AppearsInSearch
	} else {
		if Type == PUBLIC {
			AppearsInSearch = true
		} else if Type == PRIVATE {
			AppearsInSearch = false
		}
	}

	var devices []GPSDevice
	if c.TrackingDevices != nil {
		if result := db.Where("id IN ?", *c.TrackingDevices).Find(&devices); result.Error != nil {
			return nil, result.Error
		}
	}

	return &Community{Name: c.Name, Description: c.Description, Type: Type, AdminID: AdminId, PolygonArea: c.PolygonArea, Members: []auth.User{*creator}, AppearsInSearch: AppearsInSearch, TrackingDevices: devices}, nil
}

func (u *UpdateCommunity) ToCommunity(existing *Community, user *auth.User) error {
	if user.ID != existing.AdminID {
		return errors.New("only admin can modify community")
	}

	if u.PolygonArea != nil {
		if err := ValidatePolygonWKT(*u.PolygonArea); err != nil {
			return err
		}
		existing.PolygonArea = *u.PolygonArea
	}

	if u.Name != nil {
		existing.Name = *u.Name
	}

	if u.Description != nil {
		existing.Description = u.Description
	}

	if u.AppearsInSearch != nil {
		existing.AppearsInSearch = *u.AppearsInSearch
	}

	if u.Type != nil {
		if *u.Type != PUBLIC && *u.Type != PRIVATE {
			return errors.New("type should be private or public")
		}

		existing.Type = *u.Type
	}

	if u.Members != nil {
		var members []auth.User
		if result := db.Where("id IN ?", *u.Members).Find(&members); result.Error != nil {
			return result.Error
		}
		existing.Members = members
	}

	if u.Events != nil {
		var events []Event
		if result := db.Where("id IN ?", *u.Events).Find(&events); result.Error != nil {
			return result.Error
		}
		existing.Events = events
	}

	if u.TrackingDevices != nil {
		var devices []GPSDevice

		if result := db.Where("id IN ?", *u.TrackingDevices).Find(&devices); result.Error != nil {
			return result.Error
		}
		existing.TrackingDevices = devices
	}

	return nil
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
func (u *UpdateMe) ToUser(existing *auth.User) {
	if u.User == nil {
		return
	}

	if u.User.Name != nil {
		existing.Name = u.User.Name
	}
}

func (u *UpdateMe) ToSettings(existing *UserSettings) error {
	if u.Settings == nil {
		return nil
	}

	if u.Settings.TrackingDevices != nil {
		if err := db.Model(&existing).Association("TrackingDevices").Clear(); err != nil {
			return err
		}

		var devices []GPSDevice
		if result := db.Where("id IN ?", *u.Settings.TrackingDevices).Find(&devices); result.Error != nil {
			return result.Error
		}
		existing.TrackingDevices = devices
	}

	return nil
}

func ToUserProfile(u *auth.User, settings *UserSettings) *UserProfile {
	return &UserProfile{User: u, Settings: settings.ToUserSettingsOut()}
}

func (settings *UserSettings) ToUserSettingsOut() *UserSettingsOut {
	return &UserSettingsOut{TrackingDevices: settings.TrackingDevices}
}

func (c *CreateGPSDevice) ToGPSDevice(creator *auth.User) *GPSDevice {
	if c.Imei == nil {
		defaultTracking := false
		c.Tracking = &defaultTracking
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

func (u *UpdateGPSDevice) ToGPSDevice(existing *GPSDevice) {
	if u.Number != nil {
		existing.Number = u.Number
	}

	if u.Imei != nil {
		existing.Imei = u.Imei
	}

	if u.Tracking != nil {
		existing.Tracking = u.Tracking
	}
}
