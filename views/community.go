package views

import (
	"errors"
	"log"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/Hodik/geo-tracker-be/schemas"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateCommunity godoc
// @Summary Create a new community
// @Description Create a new community for the currently authenticated user
// @Tags communities
// @Accept json
// @Produce json
// @Param createCommunity body schemas.CreateCommunity true "Create community"
// @Success 201 {object} models.Community
// @Failure 400 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities [post]
func CreateCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	var schema schemas.CreateCommunity

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	comminuty, err := schema.ToCommunity(db)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Println("comminuty", comminuty)

	result := db.Create(&comminuty)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	if err := comminuty.AddMember(db, user, models.ADMIN); err != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(201, comminuty)
}

// UpdateCommunity godoc
// @Summary Update a community
// @Description Update a community for the currently authenticated user
// @Tags communities
// @Accept json
// @Produce json
// @Param id path string true "Community ID"
// @Param updateCommunity body schemas.UpdateCommunity true "Update community"
// @Success 200 {object} models.Community
// @Failure 400 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id} [patch]
func UpdateCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	var schema schemas.UpdateCommunity

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	err = schema.ToCommunity(community, user)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := db.Save(&community).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

// DeleteCommunity godoc
// @Summary Delete a community
// @Description Delete a community for the currently authenticated user
// @Tags communities
// @Produce json
// @Param id path string true "Community ID"
// @Success 204
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id} [delete]
func DeleteCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if !community.IsAdmin(user) {
		c.JSON(403, gin.H{"error": "only admin can delete a community"})
		return
	}

	if err := db.Delete(&community).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(204, community)
}

// GetCommunity godoc
// @Summary Get a community by ID
// @Description Get a community by its ID
// @Tags communities
// @Produce json
// @Param id path string true "Community ID"
// @Success 200 {object} models.Community
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id} [get]
func GetCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if !*community.AppearsInSearch && !community.IsMember(user) {
		c.JSON(403, gin.H{"error": "not allowed to get community details"})
		return
	}

	c.JSON(200, community)
}

// GetCommunities godoc
// @Summary Get all communities
// @Description Get all communities that appear in search
// @Tags communities
// @Produce json
// @Success 200 {array} models.Community
// @Failure 500 {object} schemas.Error
// @Router /api/communities [get]
func GetCommunities(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var communities []models.Community
	result := db.Select("communities.created_at, communities.deleted_at, communities.updated_at, communities.id, communities.name, communities.description, type, appears_in_search, include_external_events, allow_read_only_members_add_events").Where("deleted_at IS NULL AND appears_in_search = true").Find(&communities)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(200, communities)
}

// JoinCommunity godoc
// @Summary Join a community
// @Description Join a public community for the currently authenticated user
// @Tags communities
// @Produce json
// @Param id path string true "Community ID"
// @Success 200 {object} models.Community
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/join [post]
func JoinCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if community.Type != models.PUBLIC {
		c.JSON(403, gin.H{"error": "can only join public communities"})
		return
	}

	if community.IsMember(user) {
		c.JSON(403, gin.H{"error": "already a member of community"})
		return
	}

	if err := community.AddMember(db, user, models.READ_ONLY); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

// LeaveCommunity godoc
// @Summary Leave a community
// @Description Leave a community for the currently authenticated user
// @Tags communities
// @Produce json
// @Param id path string true "Community ID"
// @Success 204
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/leave [post]
func LeaveCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if !community.IsMember(user) {
		c.JSON(403, gin.H{"error": "not a member of this community"})
		return
	}

	if community.IsAdmin(user) && community.AdminsCount() == 1 {
		c.JSON(400, gin.H{"error": "last admin can't leave community"})
		return
	}

	if err := community.RemoveMember(db, user); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

// CreateCommunityInvite godoc
// @Summary Create a community invite
// @Description Create a new community invite for a user
// @Tags community-invites
// @Accept json
// @Produce json
// @Param createCommunityInvite body schemas.CreateCommunityInvite true "Create community invite"
// @Success 200 {object} models.CommunityInvite
// @Failure 400 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/community-invites [post]
func CreateCommunityInvite(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	var schema schemas.CreateCommunityInvite

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ci, err := schema.ToCommunityInvite(db, user)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result := db.Create(&ci)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(200, ci)
}

// UpdateCommunityInvite godoc
// @Summary Update a community invite
// @Description Update the status of a community invite
// @Tags community-invites
// @Accept json
// @Produce json
// @Param id path string true "Community Invite ID"
// @Param updateCommunityInvite body schemas.UpdateCommunityInvite true "Update community invite"
// @Success 200 {object} models.CommunityInvite
// @Failure 400 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/community-invites/{id} [patch]
func UpdateCommunityInvite(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	var schema schemas.UpdateCommunityInvite

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	var ci models.CommunityInvite
	result := db.Where("id = ?", id).First(&ci)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community invite not found"})
		return
	}

	if err := schema.ToCommunityInvite(&ci, user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&ci).Error; err != nil {
			return err
		}

		if *ci.Accepted {
			var comminuty *models.Community
			if err := db.Where("id = ?", ci.CommunityID).First(&comminuty).Error; err != nil {
				return err
			}
			if err := comminuty.AddMember(tx, user, ci.Role); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, ci)
}

// DeleteCommunityInvite godoc
// @Summary Delete a community invite
// @Description Delete a community invite by its ID
// @Tags community-invites
// @Produce json
// @Param id path string true "Community Invite ID"
// @Success 204
// @Failure 400 {object} schemas.Error
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/community-invites/{id} [delete]
func DeleteCommunityInvite(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var ci models.CommunityInvite
	result := db.Where("community_invites.id = ?", id).Joins("Community").First(&ci)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community invite not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	if ci.Accepted != nil {
		c.JSON(400, gin.H{"error": "cannot delete accepted or declined invite"})
		return
	}

	if err := ci.Community.LoadMembers(db); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if user.ID != ci.CreatorID && !ci.Community.IsAdmin(user) {
		c.JSON(403, gin.H{"error": "only creator of community invite or community admin can delete invite"})
		return
	}

	log.Println(ci)
	if err := db.Unscoped().Delete(&ci).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(204)
}

// GetCommunityInvitesUser godoc
// @Summary Get community invites for user
// @Description Get all community invites for the currently authenticated user
// @Tags community-invites
// @Produce json
// @Success 200 {array} models.CommunityInvite
// @Failure 500 {object} schemas.Error
// @Router /api/me/community-invites [get]
func GetCommunityInvitesUser(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	var communityInvites []models.CommunityInvite
	if err := db.Where("user_id = ? AND accepted is NULL", user.ID).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).Find(&communityInvites).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, communityInvites)
}

// GetCommunityInvitesCommunity godoc
// @Summary Get community invites for a community
// @Description Get all community invites for a specific community
// @Tags community-invites
// @Produce json
// @Param id path string true "Community ID"
// @Success 200 {array} models.CommunityInvite
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/invites [get]
func GetCommunityInvitesCommunity(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var communityInvites []models.CommunityInvite
	if err := db.Where("community_id = ? AND accepted is NULL", id).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).Find(&communityInvites).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, communityInvites)
}

// CommunityRemoveMember godoc
// @Summary Remove a member from a community
// @Description Remove a member from a community by an admin
// @Tags communities
// @Accept json
// @Produce json
// @Param id path string true "Community ID"
// @Param removeMember body schemas.AddMember true "Remove member"
// @Success 200 {object} models.Community
// @Failure 400 {object} schemas.Error
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/remove-member [post]
func CommunityRemoveMember(c *gin.Context) {
	reqUser := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if !community.IsAdmin(reqUser) {
		c.JSON(403, gin.H{"error": "only admin can remove members to community"})
		return
	}

	var schema schemas.AddMember

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	result := db.Where("id = ?", schema.UserID).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	log.Println("USER", user)
	log.Println("community", community)
	if err := community.RemoveMember(db, &user); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := community.LoadMembers(db); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

// CommunityAddEvent godoc
// @Summary Add an event to a community
// @Description Add an event to a community by an admin
// @Tags communities
// @Accept json
// @Produce json
// @Param id path string true "Community ID"
// @Param addEvent body schemas.AddEvent true "Add event"
// @Success 200 {object} models.Community
// @Failure 400 {object} schemas.Error
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/add-event [post]
func CommunityAddEvent(c *gin.Context) {
	reqUser := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if !community.IsAdmin(reqUser) {
		c.JSON(403, gin.H{"error": "only admin can add events to community"})
		return
	}

	var schema schemas.AddEvent

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var event models.Event
	result := db.Where("id = ?", schema.EventID).First(&event)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "event not found"})
		return
	}

	if community.IsEventLinked(&event) {
		c.JSON(400, gin.H{"error": "event is already linked to community"})
		return
	}

	if err := db.Model(&community).Association("Events").Append(&event); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

// CommunityRemoveEvent godoc
// @Summary Remove an event from a community
// @Description Remove an event from a community by an admin
// @Tags communities
// @Accept json
// @Produce json
// @Param id path string true "Community ID"
// @Param removeEvent body schemas.AddEvent true "Remove event"
// @Success 200 {object} models.Community
// @Failure 400 {object} schemas.Error
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/remove-event [post]
func CommunityRemoveEvent(c *gin.Context) {
	reqUser := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if !community.IsAdmin(reqUser) {
		c.JSON(403, gin.H{"error": "only admin can remove events to community"})
		return
	}

	var schema schemas.AddEvent

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var event models.Event
	result := db.Where("id = ?", schema.EventID).First(&event)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "event not found"})
		return
	}

	if err := db.Model(&community).Association("Events").Delete(&event); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

// CommunityTrackDevice godoc
// @Summary Track a device in a community
// @Description Track a GPS device in a community by an admin
// @Tags communities
// @Accept json
// @Produce json
// @Param id path string true "Community ID"
// @Param trackDevice body schemas.TrackDevice true "Track device"
// @Success 200 {object} models.Community
// @Failure 400 {object} schemas.Error
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/track-device [post]
func CommunityTrackDevice(c *gin.Context) {
	reqUser := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if !community.IsAdmin(reqUser) {
		c.JSON(403, gin.H{"error": "only admin can add devices to community"})
		return
	}

	var schema schemas.TrackDevice

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var device models.GPSDevice
	result := db.Where("id = ?", schema.DeviceID).First(&device)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "device not found"})
		return
	}

	if community.IsDeviceTracked(&device) {
		c.JSON(400, gin.H{"error": "community is already tracking this device"})
		return
	}

	if err := db.Model(&community).Association("TrackingDevices").Append(&device); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

// CommunityUntrackDevice godoc
// @Summary Untrack a device in a community
// @Description Untrack a GPS device in a community by an admin
// @Tags communities
// @Accept json
// @Produce json
// @Param id path string true "Community ID"
// @Param untrackDevice body schemas.TrackDevice true "Untrack device"
// @Success 200 {object} models.Community
// @Failure 400 {object} schemas.Error
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/untrack-device [post]
func CommunityUntrackDevice(c *gin.Context) {
	reqUser := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if !community.IsAdmin(reqUser) {
		c.JSON(403, gin.H{"error": "only admin can remove devices to community"})
		return
	}

	var schema schemas.TrackDevice

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var device models.GPSDevice
	result := db.Where("id = ?", schema.DeviceID).First(&device)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "device not found"})
		return
	}

	if err := db.Model(&community).Association("TrackingDevices").Delete(&device); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

// CreateCommunityAreaOfInterest godoc
// @Summary Create an area of interest for a community
// @Description Create a new area of interest for a community by an admin
// @Tags communities
// @Accept json
// @Produce json
// @Param id path string true "Community ID"
// @Param createAreaOfInterest body schemas.CreateAreaOfInterest true "Create area of interest"
// @Success 201 {object} models.AreaOfInterest
// @Failure 400 {object} schemas.Error
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/areas-of-interest [post]
func CreateCommunityAreaOfInterest(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if !community.IsAdmin(user) {
		c.JSON(403, gin.H{"error": "only admin can add areas of interest to community"})
		return
	}

	var schema schemas.CreateAreaOfInterest

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	aoiModel, err := schema.ToAreaOfInterest()

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(&aoiModel).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&community).Association("AreasOfInterest").Append(aoiModel); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, aoiModel)
}

// DeleteCommunityAreaOfInterest godoc
// @Summary Delete an area of interest from a community
// @Description Delete an area of interest from a community by an admin
// @Tags communities
// @Accept json
// @Produce json
// @Param id path string true "Community ID"
// @Param area_of_interest_id path string true "Area of Interest ID"
// @Success 204
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/areas-of-interest/{area_of_interest_id} [delete]
func DeleteCommunityAreaOfInterest(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if !community.IsAdmin(user) {
		c.JSON(403, gin.H{"error": "only admin can remove areas of interest of community"})
		return
	}

	areaOfInterestID := c.Param("area_of_interest_id")

	var count int64
	if err := db.Table("community_areas_of_interest").Where("area_of_interest_id = ? AND community_id = ?", areaOfInterestID, community.ID).Count(&count).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch area of interest"})
		return
	}

	if count == 0 {
		c.JSON(404, gin.H{"error": "Area of interest not found or doesn't belong to the community"})
		return
	}

	if err := db.Delete(&models.AreaOfInterest{Base: models.Base{ID: uuid.MustParse(areaOfInterestID)}}).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete area of interest"})
		return
	}

	c.Status(204)
}

// GetCommunityAreasOfInterest godoc
// @Summary Get areas of interest for a community
// @Description Get areas of interest for a community
// @Tags communities
// @Produce json
// @Param id path string true "Community ID"
// @Success 200 {array} models.AreaOfInterest
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/communities/{id}/areas-of-interest [get]
func GetCommunityAreasOfInterest(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	community, err := GetCommunityFromParam(c, db)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if !*community.AppearsInSearch && !community.IsMember(user) {
		c.JSON(403, gin.H{"error": "not allowed to get community areas of interest"})
		return
	}

	var aois []models.AreaOfInterest
	result := db.Model(&aois).Select("area_of_interests.id, ST_AsText(area_of_interests.polygon_area) as polygon_area, area_of_interests.created_at, area_of_interests.updated_at, area_of_interests.deleted_at").
		Joins("JOIN community_areas_of_interest ON community_areas_of_interest.area_of_interest_id = area_of_interests.id").
		Where("community_areas_of_interest.community_id = ?", community.ID).
		Preload("Events", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Find(&aois)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error})
	}

	c.JSON(200, aois)
}
