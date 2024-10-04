package main

import (
	"errors"
	"log"
	"time"

	"github.com/Hodik/geo-tracker-be/config"
	"github.com/Hodik/geo-tracker-be/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SetAPICookie(db *gorm.DB, device *models.GPSDevice) {
	cookie := createSessionCookie()
	login(cookie, *device.Imei, *device.Password)
	device.APICookie = &cookie
	db.Save(&device)
}

func ReceiveDeviceLocation(db *gorm.DB, device *models.GPSDevice) (*models.GPSLocation, error) {

	if device.Imei == nil || device.Password == nil {
		return nil, errors.New("Non GPS device, cannot get location")
	}

	log.Default().Println("Polling device", device.ID)

	if device.APICookie == nil {
		SetAPICookie(db, device)
	}

	lat, lon, err := getLocation(*device.APICookie)

	if errors.Is(err, InvalidCookieError) {
		SetAPICookie(db, device)

		lat, lon, err = getLocation(*device.APICookie)

		if err != nil {
			return nil, err
		}
	}

	location := models.GPSLocation{
		Latitude:  lat,
		Longitude: lon,
		DeviceID:  device.ID,
	}

	result := db.Save(&location)
	if result.Error != nil {
		return nil, result.Error
	}

	return &location, nil
}

func CleanUpLocations(db *gorm.DB, device *models.GPSDevice) error {

	log.Default().Println("Cleaning up locations for device", device.ID)

	var recentLocations []models.GPSLocation
	if err := db.Where("device_id = ?", device.ID).Order("created_at desc").Limit(5).Find(&recentLocations).Error; err != nil {
		return err
	}

	var recentIDs []uuid.UUID
	for _, location := range recentLocations {
		recentIDs = append(recentIDs, location.ID)
	}

	if err := db.Where("device_id = ? AND id NOT IN ?", device.ID, recentIDs).Delete(&models.GPSLocation{}).Error; err != nil {
		return err
	}

	return nil
}

func PollDevices(db *gorm.DB) {
	log.Default().Println("Location server started")
	var devices []models.GPSDevice

	for {
		db.Where("tracking = ? AND imei IS NOT NULL AND password IS NOT NULL", true).Find(&devices)

		log.Default().Println("Pulling locations for ", len(devices), " devices")
		for _, device := range devices {
			go func() {
				_, err := ReceiveDeviceLocation(db, &device)
				if err != nil {
					log.Default().Println("Failed to receive device location", err)
				}
				err = CleanUpLocations(db, &device)

				if err != nil {
					log.Default().Println("Failed to clean up locations", err)
				}

			}()
		}

		conf := config.GetConfig(db)
		log.Default().Println("Sleeping for ", conf.PollInterval, " seconds")
		time.Sleep(time.Duration(conf.PollInterval) * time.Second)
	}
}
