package main

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

	result := db.Save(user)

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

func userTrackDevice(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)

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

	userSettings, err := GetUserSettings(user)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if userSettings.IsDeviceTracked(&device) {
		c.JSON(400, gin.H{"error": "user is already tracking this device"})
		return
	}

	if err := db.Model(userSettings).Association("TrackingDevices").Append(&device); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, ToUserProfile(user, userSettings))
}

func userUntrackDevice(c *gin.Context) {
	user := c.MustGet("user").(*auth.User)

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

	userSettings, err := GetUserSettings(user)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&userSettings).Association("TrackingDevices").Delete(&device); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, ToUserProfile(user, userSettings))
}
