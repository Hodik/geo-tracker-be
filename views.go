package main

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func getGPSDevices(c *gin.Context) {
	var devices []GPSDevice
	result := db.Omit("password").Find(&devices)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(200, devices)
}

func createGPSDevice(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)
	var device CreateGPSDevice

	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	deviceModel := device.ToGPSDevice(user)
	result := db.Create(&deviceModel)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(201, deviceModel)
}

func updateGPSDevice(c *gin.Context) {
	var device UpdateGPSDevice

	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	var deviceModel GPSDevice
	result := db.Where("id = ?", id).First(&deviceModel)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "device not found"})
		return
	}

	device.ToGPSDevice(&deviceModel)

	result = db.Save(&deviceModel)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(200, deviceModel)
}

func getGPSDevice(c *gin.Context) {
	id := c.Param("id")

	var device GPSDevice
	result := db.Preload("Locations", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Where("id = ?", id).First(&device)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "device not found"})
		return
	}

	c.JSON(200, device)
}

func getMe(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)

	userSettings, err := GetUserSettings(user)

	if err != nil {
		c.JSON(500, gin.H{"error": err})
	}

	c.JSON(200, ToUserProfile(user, userSettings))
}

func updateMe(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)

	var updateMe UpdateMe

	if err := c.ShouldBindJSON(&updateMe); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userSettings, err := GetUserSettings(user)

	if err != nil {
		c.JSON(500, gin.H{"error": err})
	}

	updateMe.ToUser(user)
	if err = updateMe.ToSettings(userSettings); err != nil {
		c.JSON(500, gin.H{"error": err})
	}

	result := db.Save(user)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	result = db.Save(userSettings)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(200, ToUserProfile(user, userSettings))
}

func createAreaOfInterest(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)
	var schema CreateAreaOfInterest

	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	aoiModel, err := schema.ToAreaOfInterest(user)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result := db.Create(&aoiModel)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(201, aoiModel)
}

func getMyAreasOfInterest(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)
	var aois []AreaOfInterest
	result := db.Raw("SELECT id, user_id, ST_AsText(polygon_area) as polygon_area, created_at, updated_at, deleted_at FROM area_of_interests WHERE user_id = ?", user.ID).Scan(&aois)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error})
	}
	c.JSON(200, aois)
}

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
	result := db.Where("id = ?", id).First(&community)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	err := schema.ToCommunity(&community, user)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result = db.Save(&community)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(201, community)
}

func getCommunity(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)
	id := c.Param("id")

	var community Community
	result := db.Select("communities.created_at, communities.deleted_at, communities.updated_at, communities.id, communities.name, communities.description, ST_AsText(polygon_area) AS polygon_area, type, admin_id").Preload("Members", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Preload("Events", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Preload("TrackingDevices", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Joins("Admin").Where("communities.id = ?", id).First(&community)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
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
	result := db.Select("communities.created_at, communities.deleted_at, communities.updated_at, communities.id, communities.name, communities.description, ST_AsText(polygon_area) AS polygon_area, type, admin_id").Preload("Members", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Where("communities.id = ?", id).First(&community)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
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

	community.Members = append(community.Members, *user)
	result = db.Save(community)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
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
		if err := tx.Save(ci).Error; err != nil {
			return err
		}

		if *ci.Accepted {
			var comminuty *Community
			if err := db.Where("id = ?", ci.CommunityID).Preload("Members").First(comminuty).Error; err != nil {
				return err
			}

			comminuty.Members = append(comminuty.Members, *user)
			if err := db.Save(comminuty).Error; err != nil {
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
