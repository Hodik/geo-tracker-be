package schemas

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateCommunity struct {
	Name                          string                `json:"name" binding:"required"`
	Description                   *string               `json:"description"`
	Type                          *models.CommunityType `json:"type"`
	AppearsInSearch               *bool                 `json:"appears_in_search"`
	IncludeExternalEvents         *bool                 `json:"include_external_events"`
	AllowReadOnlyMembersAddEvents *bool                 `json:"allow_read_only_members_add_events"`
	TrackingDevices               *[]uuid.UUID          `json:"tracking_devices"`
}

type UpdateCommunity struct {
	Name                          *string               `json:"name"`
	Description                   *string               `json:"description"`
	Type                          *models.CommunityType `json:"type"`
	AppearsInSearch               *bool                 `json:"appears_in_search"`
	IncludeExternalEvents         *bool                 `json:"include_external_events"`
	AllowReadOnlyMembersAddEvents *bool                 `json:"allow_read_only_members_add_events"`
	PolygonArea                   *string               `json:"polygon_area"`
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

	if c.Type != nil {
		if err := models.ValidateCommunityType(string(*c.Type)); err != nil {
			return nil, err
		}
	}

	communityType := models.PUBLIC
	if c.Type != nil {
		communityType = *c.Type
	}

	// Set Flags Based on Type
	IncludeExternalEvents := c.IncludeExternalEvents != nil && *c.IncludeExternalEvents || communityType == models.PUBLIC
	AllowReadOnlyMembersAddEvents := c.AllowReadOnlyMembersAddEvents != nil && *c.AllowReadOnlyMembersAddEvents || communityType == models.PUBLIC
	AppearsInSearch := c.AppearsInSearch != nil && *c.AppearsInSearch || communityType == models.PUBLIC

	var devices []*models.GPSDevice
	if c.TrackingDevices != nil {
		if result := db.Where("id IN ?", *c.TrackingDevices).Find(&devices); result.Error != nil {
			return nil, result.Error
		}
	}

	return &models.Community{
			Name:                          c.Name,
			Description:                   c.Description,
			Type:                          communityType,
			Members:                       []*models.CommunityMember{},
			AppearsInSearch:               &AppearsInSearch,
			AllowReadOnlyMembersAddEvents: &AllowReadOnlyMembersAddEvents,
			IncludeExternalEvents:         &IncludeExternalEvents,
			TrackingDevices:               devices,
		},
		nil
}

func (u *UpdateCommunity) ToCommunity(existing *models.Community, user *models.User) error {
	if !existing.IsAdmin(user) {
		return errors.New("only admin can modify community")
	}

	if u.Name != nil {
		existing.Name = *u.Name
	}

	if u.Description != nil {
		existing.Description = u.Description
	}

	if u.AppearsInSearch != nil {
		existing.AppearsInSearch = u.AppearsInSearch
	}

	if u.IncludeExternalEvents != nil {
		existing.IncludeExternalEvents = u.IncludeExternalEvents
	}

	if u.AllowReadOnlyMembersAddEvents != nil {
		existing.AllowReadOnlyMembersAddEvents = u.AllowReadOnlyMembersAddEvents
	}

	if u.Type != nil {
		if err := models.ValidateCommunityType(string(*u.Type)); err != nil {
			return err
		}

		existing.Type = *u.Type
	}

	return nil
}
