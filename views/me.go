package views

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/Hodik/geo-tracker-be/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetMe godoc
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags me
// @Produce json
// @Success 200 {object} schemas.UserProfile
// @Failure 500 {object} schemas.Error
// @Router /me [get]
func GetMe(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	userSettings, err := models.GetUserSettings(db, user)

	if err != nil {
		c.JSON(500, gin.H{"error": err})
	}

	c.JSON(200, schemas.ToUserProfile(user, userSettings))
}

// UpdateMe godoc
// @Summary Update current user profile
// @Description Update the profile of the currently authenticated user
// @Tags me
// @Accept json
// @Produce json
// @Param updateMe body schemas.UpdateMe true "Update user profile"
// @Success 200 {object} schemas.UserProfile
// @Failure 400 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /me [patch]
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

// CreateAreaOfInterest godoc
// @Summary Create a new area of interest
// @Description Create a new area of interest for the currently authenticated user
// @Tags me
// @Accept json
// @Produce json
// @Param createAreaOfInterest body schemas.CreateAreaOfInterest true "Create area of interest"
// @Success 201 {object} models.AreaOfInterest
// @Failure 400 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /me/areas-of-interest [post]
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

// GetMyAreasOfInterest godoc
// @Summary Get areas of interest
// @Description Get all areas of interest for the currently authenticated user
// @Tags me
// @Produce json
// @Success 200 {array} models.AreaOfInterest
// @Failure 500 {object} schemas.Error
// @Router /me/areas-of-interest [get]
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

// UserTrackDevice godoc
// @Summary Track a device
// @Description Track a GPS device for the currently authenticated user
// @Tags me
// @Accept json
// @Produce json
// @Param trackDevice body schemas.TrackDevice true "Track device"
// @Success 200 {object} schemas.UserProfile
// @Failure 400 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /me/track-device [post]
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

// UserUntrackDevice godoc
// @Summary Untrack a device
// @Description Untrack a GPS device for the currently authenticated user
// @Tags me
// @Accept json
// @Produce json
// @Param trackDevice body schemas.TrackDevice true "Untrack device"
// @Success 200 {object} schemas.UserProfile
// @Failure 400 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /me/untrack-device [post]
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

// GetMyCommunities godoc
// @Summary Get user communities
// @Description Get all communities the currently authenticated user is a member of
// @Tags me
// @Produce json
// @Success 200 {array} models.Community
// @Failure 500 {object} schemas.Error
// @Router /me/communities [get]
func GetMyCommunities(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	communities, err := user.FetchCommunities(db)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, communities)
}
