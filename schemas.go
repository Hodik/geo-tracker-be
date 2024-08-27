package main

type CreateGPSDevice struct {
	Imei     string `json:"imei" binding:"required"`
	Password string `json:"password" binding:"required"`

	Number *string `json:"number"`
}

type UpdateGPSDevice struct {
	Number *string `json:"number"`
	Imei   *string `json:"imei"`
}

func (c *CreateGPSDevice) ToGPSDevice() *GPSDevice {
	return &GPSDevice{
		Number:   c.Number,
		Imei:     c.Imei,
		Password: c.Password,
	}
}

func (u *UpdateGPSDevice) ToGPSDevice(existing *GPSDevice) *GPSDevice {
	if u.Number != nil {
		existing.Number = u.Number
	}

	if u.Imei != nil {
		existing.Imei = *u.Imei
	}

	return existing
}
