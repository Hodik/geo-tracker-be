package views

import (
	"errors"
	"log"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/Hodik/geo-tracker-be/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

func UpdateCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	var schema schemas.UpdateCommunity

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

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

	c.JSON(200, community)
}

func DeleteCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
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

func GetCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if !community.AppearsInSearch && !community.IsMember(user) {
		c.JSON(403, gin.H{"error": "not allowed to get community details"})
		return
	}

	c.JSON(200, community)
}

func GetCommunities(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var communities []models.Community
	result := db.Select("communities.created_at, communities.deleted_at, communities.updated_at, communities.id, communities.name, communities.description, ST_AsText(polygon_area) AS polygon_area, type, appears_in_search").Where("deleted_at IS NULL AND appears_in_search = true").Find(&communities)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(200, communities)
}

func JoinCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
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

func LeaveCommunity(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
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

	if user.ID != ci.CreatorID && ci.Community.IsAdmin(user) {
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

func CommunityRemoveMember(c *gin.Context) {
	reqUser := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
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

func CommunityAddEvent(c *gin.Context) {
	reqUser := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
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

func CommunityRemoveEvent(c *gin.Context) {
	reqUser := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
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

func CommunityTrackDevice(c *gin.Context) {
	reqUser := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
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

func CommunityUntrackDevice(c *gin.Context) {
	reqUser := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
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
