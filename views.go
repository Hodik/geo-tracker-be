package main

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getGPSDevices(c *gin.Context) {
	var devices []GPSDevice
	db.Omit("password").Find(&devices)

	c.JSON(200, devices)
}

func createGPSDevice(c *gin.Context) {
	var device CreateGPSDevice

	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	deviceModel := device.ToGPSDevice()

	db.Create(&deviceModel)

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

	db.Save(&deviceModel)

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
