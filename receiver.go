package main

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseLocationMessage(message string) (float64, float64, error) {
	parts := strings.Split(message, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid format for Body: %s", message)
	}

	latitude, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid latitude: %s", parts[0])
	}

	longitude, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid longitude: %s", parts[1])
	}

	return latitude, longitude, nil
}

func ReceiveDeviceLocation(number string, latitude float64, longitude float64) (*GPSLocation, error) {
	var device GPSDevice
	result := db.Where("number = ?", number).First(&device)

	if result.Error != nil {
		return nil, result.Error
	}

	location := GPSLocation{
		Latitude:  latitude,
		Longitude: longitude,
		DeviceID:  device.ID,
	}

	result = db.Save(&location)
	if result.Error != nil {
		return nil, result.Error
	}

	return &location, nil
}
