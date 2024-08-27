package main

import (
	"errors"
	"log"
	"time"
)

func (device *GPSDevice) SetAPICookie() {
	cookie := createSessionCookie()
	login(cookie, device.Imei, device.Password)
	device.APICookie = &cookie
	db.Save(&device)
}

func (device *GPSDevice) ReceiveDeviceLocation() (*GPSLocation, error) {

	log.Default().Println("Polling device", device.ID)

	if device.APICookie == nil {
		device.SetAPICookie()
	}

	lat, lon, err := getLocation(*device.APICookie)

	if errors.Is(err, InvalidCookieError) {
		device.SetAPICookie()

		lat, lon, err = getLocation(*device.APICookie)

		if err != nil {
			return nil, err
		}
	}

	location := GPSLocation{
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

func (device *GPSDevice) CleanUpLocations() error {

	log.Default().Println("Cleaning up locations for device", device.ID)

	var recentLocations []GPSLocation
	if err := db.Where("device_id = ?", device.ID).Order("created_at desc").Limit(5).Find(&recentLocations).Error; err != nil {
		return err
	}

	var recentIDs []uint
	for _, location := range recentLocations {
		recentIDs = append(recentIDs, location.ID)
	}

	if err := db.Where("device_id = ? AND id NOT IN ?", device.ID, recentIDs).Delete(&GPSLocation{}).Error; err != nil {
		return err
	}

	return nil
}

func PollDevices() {
	var devices []GPSDevice
	db.Find(&devices)

	for {
		for _, device := range devices {
			go device.ReceiveDeviceLocation()
			go device.CleanUpLocations()
		}
		time.Sleep(30 * time.Second)
	}
}
