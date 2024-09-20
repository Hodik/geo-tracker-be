package schemas

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Hodik/geo-tracker-be/models"
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

type UpdateMe struct {
	User *UpdateUser `json:"user"`
}

type UserSettingsOut struct {
	TrackingDevices []models.GPSDevice `json:"tracking_devices"`
}

type UserProfile struct {
	User     *models.User         `json:"user"`
	Settings *models.UserSettings `json:"settings"`
}

type CreateCommunity struct {
	Name            string                `json:"name" binding:"required"`
	Description     *string               `json:"description"`
	Type            *models.CommunityType `json:"type"`
	AppearsInSearch *bool                 `json:"appears_in_search"`
	PolygonArea     string                `json:"polygon_area" binding:"required"`
	TrackingDevices *[]uuid.UUID          `json:"tracking_devices"`
}

type UpdateCommunity struct {
	Name            *string               `json:"name"`
	Description     *string               `json:"description"`
	Type            *models.CommunityType `json:"type"`
	AppearsInSearch *bool                 `json:"appears_in_search"`
	PolygonArea     *string               `json:"polygon_area"`
}

type CreateCommunityInvite struct {
	CommunityID uuid.UUID         `json:"community_id" binding:"required"`
	UserID      uuid.UUID         `json:"user_id" binding:"required"`
	Role        models.MemberRole `json:"role"`
}

type UpdateCommunityInvite struct {
	Accepted bool `json:"accepted" binding:"required"`
}

type TrackDevice struct {
	DeviceID uuid.UUID `json:"device_id" binding:"required"`
}

type AddEvent struct {
	EventID uuid.UUID `json:"event_id" binding:"required"`
}

type AddMember struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

func (c *CreateCommunityInvite) ToCommunityInvite(db *gorm.DB, creator *models.User) (*models.CommunityInvite, error) {

	var community models.Community
	err := community.Fetch(db, c.CommunityID.String())

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("community doesn't exist")
	}

	if err != nil {
		return nil, errors.New(err.Error())
	}

	var user models.User

	result := db.Where("users.id = ?", c.UserID).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("user doesn't exist")
	}

	if result.Error != nil {
		return nil, errors.New(result.Error.Error())
	}

	if !community.IsAdmin(creator) {
		return nil, errors.New("only community's admin can invite users")
	}

	if community.IsMember(&user) {
		return nil, errors.New("user is already a member of community")
	}

	return &models.CommunityInvite{UserID: c.UserID, CommunityID: c.CommunityID, CreatorID: creator.ID, Role: c.Role}, nil

}

func (u *UpdateCommunityInvite) ToCommunityInvite(existing *models.CommunityInvite, user *models.User) error {
	if user.ID != existing.UserID {
		return errors.New("only invited user can accept")
	}

	existing.Accepted = &u.Accepted
	return nil
}
func (c *CreateCommunity) ToCommunity(db *gorm.DB) (*models.Community, error) {
	if err := ValidatePolygonWKT(c.PolygonArea); err != nil {
		return nil, err
	}

	if c.Type != nil {
		if *c.Type != models.PUBLIC && *c.Type != models.PRIVATE {
			return nil, errors.New("type should be private or public")
		}
	}

	var Type models.CommunityType = models.PUBLIC
	if c.Type != nil {
		Type = *c.Type
	}

	var AppearsInSearch bool
	if c.AppearsInSearch != nil {
		AppearsInSearch = *c.AppearsInSearch
	} else {
		if Type == models.PUBLIC {
			AppearsInSearch = true
		} else if Type == models.PRIVATE {
			AppearsInSearch = false
		}
	}

	var devices []models.GPSDevice
	if c.TrackingDevices != nil {
		if result := db.Where("id IN ?", *c.TrackingDevices).Find(&devices); result.Error != nil {
			return nil, result.Error
		}
	}

	return &models.Community{Name: c.Name, Description: c.Description, Type: Type, PolygonArea: c.PolygonArea, Members: []models.CommunityMember{}, AppearsInSearch: AppearsInSearch, TrackingDevices: devices}, nil
}

func (u *UpdateCommunity) ToCommunity(existing *models.Community, user *models.User) error {
	if !existing.IsAdmin(user) {
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
		if *u.Type != models.PUBLIC && *u.Type != models.PRIVATE {
			return errors.New("type should be private or public")
		}

		existing.Type = *u.Type
	}

	return nil
}

func (c *CreateAreaOfInterest) ToAreaOfInterest(creator *models.User) (*models.AreaOfInterest, error) {
	if err := ValidatePolygonWKT(c.PolygonArea); err != nil {
		return nil, err
	}

	return &models.AreaOfInterest{UserID: creator.ID, PolygonArea: c.PolygonArea}, nil
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
func (u *UpdateMe) ToUser(existing *models.User) {
	if u.User == nil {
		return
	}

	if u.User.Name != nil {
		existing.Name = u.User.Name
	}
}

func ToUserProfile(u *models.User, settings *models.UserSettings) *UserProfile {
	return &UserProfile{User: u, Settings: settings}
}

func (c *CreateGPSDevice) ToGPSDevice(creator *models.User) *models.GPSDevice {
	if c.Imei == nil {
		defaultTracking := false
		c.Tracking = &defaultTracking
	}

	d := &models.GPSDevice{
		Number:      c.Number,
		Imei:        c.Imei,
		Password:    c.Password,
		Tracking:    c.Tracking,
		CreatedByID: &creator.ID,
	}
	return d
}

func (u *UpdateGPSDevice) ToGPSDevice(existing *models.GPSDevice) {
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
