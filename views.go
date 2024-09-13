package main

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	result := db.First(&deviceModel, id)

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
	}).First(&device, id)

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

	user = updateMe.ToUser(user)
	userSettings, err = updateMe.ToSettings(userSettings)

	if err != nil {
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
	result := db.First(&community, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	comminuty, err := schema.ToCommunity(&community, user)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result = db.Save(&comminuty)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(201, comminuty)
}

func getCommunity(c *gin.Context) {
	id := c.Param("id")

	var community Community
	result := db.Select("communities.created_at, communities.deleted_at, communities.updated_at, communities.id, communities.name, communities.description, ST_AsText(polygon_area) AS polygon_area, type, admin_id").Preload("Members", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Preload("Events", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Preload("TrackingDevices", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Joins("Admin").First(&community, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "community not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
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
