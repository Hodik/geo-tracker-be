package main

import "log"

type CreateGPSDevice struct {
	Imei     string `json:"imei" binding:"required"`
	Password string `json:"password" binding:"required"`

	Number   *string `json:"number"`
	Tracking *bool   `json:"tracking"`
}

type UpdateGPSDevice struct {
	Number   *string `json:"number"`
	Imei     *string `json:"imei"`
	Tracking *bool   `json:"tracking"`
}

func (c *CreateGPSDevice) ToGPSDevice() *GPSDevice {
	log.Println(c.Tracking)
	d := &GPSDevice{
		Number:   c.Number,
		Imei:     c.Imei,
		Password: c.Password,
		Tracking: c.Tracking,
	}
	return d
}

func (u *UpdateGPSDevice) ToGPSDevice(existing *GPSDevice) *GPSDevice {
	if u.Number != nil {
		existing.Number = u.Number
	}

	if u.Imei != nil {
		existing.Imei = *u.Imei
	}

	if u.Tracking != nil {
		existing.Tracking = u.Tracking
	}

	return existing
}
