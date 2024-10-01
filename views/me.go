package views

import (
	"errors"
	"strconv"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/Hodik/geo-tracker-be/schemas"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
func CreateMyAreaOfInterest(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

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

	if err := aoiModel.Create(db); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(user).Association("AreasOfInterest").Append(aoiModel); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, aoiModel)
}

// DeleteAreaOfInterest godoc
// @Summary Delete an area of interest
// @Description Delete an area of interest for the currently authenticated user
// @Tags me
// @Produce json
// @Param area_of_interest_id path string true "Area of interest ID"
// @Success 204
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /me/areas-of-interest/{area_of_interest_id} [delete]
func DeleteMyAreaOfInterest(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	areaOfInterestID := c.Param("area_of_interest_id")

	var count int64
	if err := db.Table("user_areas_of_interest").Where("area_of_interest_id = ? AND user_id = ?", areaOfInterestID, user.ID).Count(&count).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch area of interest"})
		return
	}

	if count == 0 {
		c.JSON(404, gin.H{"error": "Area of interest not found or doesn't belong to the user"})
		return
	}

	if err := db.Delete(&models.AreaOfInterest{Base: models.Base{ID: uuid.MustParse(areaOfInterestID)}}).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete area of interest"})
		return
	}

	c.Status(204)
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
	result := db.Model(&aois).Select("area_of_interests.id, ST_AsText(area_of_interests.polygon_area) as polygon_area, area_of_interests.created_at, area_of_interests.updated_at, area_of_interests.deleted_at").
		Joins("JOIN user_areas_of_interest ON user_areas_of_interest.area_of_interest_id = area_of_interests.id").
		Where("user_areas_of_interest.user_id = ?", user.ID).
		Preload("Events", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Find(&aois)

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

func MyFeed(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	db := c.MustGet("db").(*gorm.DB)

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	query := `
		SELECT events.*
		FROM events
		LEFT JOIN event_areas_of_interest eaoi ON eaoi.event_id = events.id
		LEFT JOIN user_areas_of_interest uaoi ON uaoi.area_of_interest_id = eaoi.area_of_interest_id
		WHERE uaoi.user_id = ?
		UNION
		SELECT events.*
		FROM events
		WHERE events.created_by_id = ?
		ORDER BY created_at DESC
	`

	countQuery := `SELECT COUNT(*) FROM (` + query + `)`
	paginatedQuery := query + `LIMIT ? OFFSET ?`

	var total int64
	if err := db.Raw(countQuery, user.ID, user.ID).Count(&total).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var events []models.Event
	if err := db.Raw(paginatedQuery, user.ID, user.ID, pageSize, (page-1)*pageSize).Scan(&events).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	paginated := schemas.Paginated{Page: page, PageSize: pageSize, Total: int(total), Items: events}
	c.JSON(200, paginated)
}
