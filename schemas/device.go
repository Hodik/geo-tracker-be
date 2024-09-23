package schemas

import "github.com/Hodik/geo-tracker-be/models"

type CreateGPSDevice struct {
	Imei     *string `json:"imei"`
	Password *string `json:"password"`

	Number   *string `json:"number"`
	Tracking *bool   `json:"tracking"`
}

type UpdateGPSDevice struct {
	Number   *string `json:"number"`
	Imei     *string `json:"imei"`
	Tracking *bool   `json:"tracking"`
}

func (c *CreateGPSDevice) ToGPSDevice(creator *models.User) *models.GPSDevice {
	if c.Imei == nil {
		defaultTracking := false
		c.Tracking = &defaultTracking
	}

	d := &models.GPSDevice{
		Number:      c.Number,
		Imei:        c.Imei,
		Password:    c.Password,
		Tracking:    c.Tracking,
		CreatedByID: &creator.ID,
	}
	return d
}

func (u *UpdateGPSDevice) ToGPSDevice(existing *models.GPSDevice) {
	if u.Number != nil {
		existing.Number = u.Number
	}

	if u.Imei != nil {
		existing.Imei = u.Imei
	}

	if u.Tracking != nil {
		existing.Tracking = u.Tracking
	}
}
