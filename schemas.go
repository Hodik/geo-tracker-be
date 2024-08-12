package main

type CreateGPSDevice struct {
	Number string  `json:"number" binding:"required"`
	Imei   *uint64 `json:"imei"`
}

type UpdateGPSDevice struct {
	Number *string `json:"number"`
	Imei   *uint64 `json:"imei"`
}

func (c *CreateGPSDevice) ToGPSDevice() *GPSDevice {
	return &GPSDevice{
		Number: c.Number,
		Imei:   c.Imei,
	}
}

func (u *UpdateGPSDevice) ToGPSDevice(existing *GPSDevice) *GPSDevice {
	if u.Number != nil {
		existing.Number = *u.Number
	}

	if u.Imei != nil {
		existing.Imei = u.Imei
	}

	return existing
}
