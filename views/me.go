package views

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/Hodik/geo-tracker-be/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetMe(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	userSettings, err := models.GetUserSettings(db, user)

	if err != nil {
		c.JSON(500, gin.H{"error": err})
	}

	c.JSON(200, schemas.ToUserProfile(user, userSettings))
}

func UpdateMe(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	var updateMe schemas.UpdateMe

	if err := c.ShouldBindJSON(&updateMe); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userSettings, err := models.GetUserSettings(db, user)

	if err != nil {
		c.JSON(500, gin.H{"error": err})
	}

	updateMe.ToUser(user)

	result := db.Save(user)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(200, schemas.ToUserProfile(user, userSettings))
}

func CreateAreaOfInterest(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	var schema schemas.CreateAreaOfInterest

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

func GetMyAreasOfInterest(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	var aois []models.AreaOfInterest
	result := db.Raw("SELECT id, user_id, ST_AsText(polygon_area) as polygon_area, created_at, updated_at, deleted_at FROM area_of_interests WHERE user_id = ?", user.ID).Scan(&aois)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error})
	}
	c.JSON(200, aois)
}

func UserTrackDevice(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

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

	userSettings, err := models.GetUserSettings(db, user)

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

	c.JSON(200, schemas.ToUserProfile(user, userSettings))
}

func UserUntrackDevice(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

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

	userSettings, err := models.GetUserSettings(db, user)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&userSettings).Association("TrackingDevices").Delete(&device); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, schemas.ToUserProfile(user, userSettings))
}

func GetMyCommunities(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	var communities []models.Community

	result := db.
		Table("communities").
		Joins("JOIN community_members ON community_members.community_id = communities.id").
		Where("community_members.user_id = ?", user.ID).
		Select("communities.created_at, communities.deleted_at, communities.updated_at, communities.id, communities.name, communities.description, ST_AsText(polygon_area) AS polygon_area, type, admin_id").
		Group("communities.id").
		Find(&communities)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(200, communities)
}
