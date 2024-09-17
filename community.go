package main

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func createCommunity(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)
	var schema CreateCommunity

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	comminuty, err := schema.ToCommunity(user)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result := db.Create(&comminuty)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(201, comminuty)
}

func updateCommunity(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)
	var schema UpdateCommunity

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	var community Community
	err := community.Fetch(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	err = schema.ToCommunity(&community, user)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := db.Save(&community).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, community)
}

func getCommunity(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)
	id := c.Param("id")

	var community Community
	err := community.Fetch(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if !community.AppearsInSearch && !community.IsMember(user) && community.AdminID != user.ID {
		c.JSON(403, gin.H{"error": "not allowed to get community details"})
		return
	}

	c.JSON(200, community)
}

func getCommunities(c *gin.Context) {
	var communities []Community
	result := db.Select("communities.created_at, communities.deleted_at, communities.updated_at, communities.id, communities.name, communities.description, ST_AsText(polygon_area) AS polygon_area, type, admin_id").Find(&communities)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(200, communities)
}

func joinCommunity(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)
	id := c.Param("id")

	var community Community
	err := community.Fetch(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if community.Type != PUBLIC {
		c.JSON(403, gin.H{"error": "can only join public communities"})
		return
	}

	if community.IsMember(user) {
		c.JSON(403, gin.H{"error": "already a member of community"})
		return
	}

	if err := db.Model(&community).Association("Members").Append(user); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := db.Save(&community).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

func createCommunityInvite(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)

	var schema CreateCommunityInvite

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ci, err := schema.ToCommunityInvite(user)

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

func updateCommunityInvite(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)

	var schema UpdateCommunityInvite

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	var ci CommunityInvite
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
			var comminuty *Community
			if err := db.Model(&comminuty).Association("Members").Append(user); err != nil {
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

func deleteCommunityInvite(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)

	id := c.Param("id")

	var ci CommunityInvite
	result := db.Where("id = ?", id).Joins("Community").First(&ci)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community invite not found"})
		return
	}

	if ci.Accepted != nil {
		c.JSON(400, gin.H{"error": "cannot delete accepted or declined invite"})
		return
	}

	if user.ID != ci.CreatorID && user.ID != ci.Community.AdminID {
		c.JSON(403, gin.H{"error": "only creator of community invite or community admin cab delete invite"})
		return
	}

	if err := db.Delete(&ci).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(204)
}

func getCommunityInvitesUser(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)

	var communityInvites []CommunityInvite
	if err := db.Where("user_id = ?", user.ID).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).Find(&communityInvites).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, communityInvites)
}

func getCommunityInvitesCommunity(c *gin.Context) {
	id := c.Param("id")

	var communityInvites []CommunityInvite
	if err := db.Where("community_id = ?", id).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).Find(&communityInvites).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, communityInvites)
}

func communityAddMember(c *gin.Context) {
	reqUser := c.MustGet("user").(*auth.User)

	id := c.Param("id")

	var community Community
	err := community.Fetch(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if reqUser.ID != community.AdminID {
		c.JSON(403, gin.H{"error": "only admin can add members to community"})
		return
	}

	var schema AddMember

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user auth.User
	result := db.Where("id = ?", schema.UserID).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	if err := db.Model(&community).Association("Members").Append(&user); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

func communityRemoveMember(c *gin.Context) {
	reqUser := c.MustGet("user").(*auth.User)

	id := c.Param("id")

	var community Community
	err := community.Fetch(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if reqUser.ID != community.AdminID {
		c.JSON(403, gin.H{"error": "only admin can remove members to community"})
		return
	}

	var schema AddMember

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user auth.User
	result := db.Where("id = ?", schema.UserID).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	if err := db.Model(&community).Association("Members").Delete(&user); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

func communityAddEvent(c *gin.Context) {
	reqUser := c.MustGet("user").(*auth.User)

	id := c.Param("id")

	var community Community
	err := community.Fetch(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if reqUser.ID != community.AdminID {
		c.JSON(403, gin.H{"error": "only admin can add events to community"})
		return
	}

	var schema AddEvent

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var event Event
	result := db.Where("id = ?", schema.EventID).First(&event)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "event not found"})
		return
	}

	if err := db.Model(&community).Association("Events").Append(&event); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

func communityRemoveEvent(c *gin.Context) {
	reqUser := c.MustGet("user").(*auth.User)

	id := c.Param("id")

	var community Community
	err := community.Fetch(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if reqUser.ID != community.AdminID {
		c.JSON(403, gin.H{"error": "only admin can remove events to community"})
		return
	}

	var schema AddEvent

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var event Event
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

func communityTrackDevice(c *gin.Context) {
	reqUser := c.MustGet("user").(*auth.User)

	id := c.Param("id")

	var community Community
	err := community.Fetch(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if reqUser.ID != community.AdminID {
		c.JSON(403, gin.H{"error": "only admin can add devices to community"})
		return
	}

	var schema TrackDevice

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var device GPSDevice
	result := db.Where("id = ?", schema.DeviceID).First(&device)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "device not found"})
		return
	}

	if err := db.Model(&community).Association("TrackingDevices").Append(&device); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, community)
}

func communityUntrackDevice(c *gin.Context) {
	reqUser := c.MustGet("user").(*auth.User)

	id := c.Param("id")

	var community Community
	err := community.Fetch(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if reqUser.ID != community.AdminID {
		c.JSON(403, gin.H{"error": "only admin can remove devices to community"})
		return
	}

	var schema TrackDevice

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var device GPSDevice
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
