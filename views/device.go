package views

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/Hodik/geo-tracker-be/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
