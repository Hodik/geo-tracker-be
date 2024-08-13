package main

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
