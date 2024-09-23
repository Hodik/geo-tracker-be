package views

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/Hodik/geo-tracker-be/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetGPSDevices godoc
// @Summary Get all GPS devices
// @Description Get a list of all GPS devices for the currently authenticated user
// @Tags devices
// @Produce json
// @Success 200 {array} models.GPSDevice
// @Failure 500 {object} schemas.Error
// @Router /api/devices [get]
func GetGPSDevices(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var devices []models.GPSDevice
	result := db.Omit("password").Find(&devices)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(200, devices)
}

// CreateGPSDevice godoc
// @Summary Create a new GPS device
// @Description Create a new GPS device for the currently authenticated user
// @Tags devices
// @Accept json
// @Produce json
// @Param createGPSDevice body schemas.CreateGPSDevice true "Create GPS device"
// @Success 201 {object} models.GPSDevice
// @Failure 400 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/devices [post]
func CreateGPSDevice(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	user := c.MustGet("user").(*models.User)

	var device schemas.CreateGPSDevice

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

// UpdateGPSDevice godoc
// @Summary Update a GPS device
// @Description Update a GPS device for the currently authenticated user
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param updateGPSDevice body schemas.UpdateGPSDevice true "Update GPS device"
// @Success 200 {object} models.GPSDevice
// @Failure 400 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/devices/{id} [patch]
func UpdateGPSDevice(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var device schemas.UpdateGPSDevice

	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	var deviceModel models.GPSDevice
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

// GetGPSDevice godoc
// @Summary Get a GPS device by ID
// @Description Get details of a GPS device by its ID, including its locations
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} models.GPSDevice
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/devices/{id} [get]
func GetGPSDevice(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	id := c.Param("id")

	var device models.GPSDevice
	result := db.Preload("Locations", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Where("id = ?", id).First(&device)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "device not found"})
		return
	}

	c.JSON(200, device)
}
