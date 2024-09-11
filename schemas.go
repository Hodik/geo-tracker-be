package main

import (
	"log"

	"github.com/Hodik/geo-tracker-be/auth"
)

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

type UpdateUser struct {
	Name *string `json:"name"`
}

type UpdateSettings struct {
	TrackingDevices *[]uint `json:"tracking_devices"`
}

type UpdateMe struct {
	User     *UpdateUser     `json:"user"`
	Settings *UpdateSettings `json:"settings"`
}

type UserSettingsOut struct {
	TrackingDevices []GPSDevice `json:"tracking_devices"`
}

type UserProfile struct {
	User     *auth.User       `json:"user"`
	Settings *UserSettingsOut `json:"settings"`
}

func (u *UpdateMe) ToUser(existing *auth.User) *auth.User {
	if u.User == nil {
		return existing
	}

	if u.User.Name != nil {
		existing.Name = u.User.Name
	}

	return existing
}

func (u *UpdateMe) ToSettings(existing *UserSettings) (*UserSettings, error) {
	if u.Settings == nil {
		return existing, nil
	}

	if u.Settings.TrackingDevices != nil {
		var devices []GPSDevice
		if result := db.Where("id IN ?", *u.Settings.TrackingDevices).Find(&devices); result.Error != nil {
			return nil, result.Error
		}
		existing.TrackingDevices = devices
	}

	return existing, nil
}

func ToUserProfile(u *auth.User, settings *UserSettings) *UserProfile {
	return &UserProfile{User: u, Settings: settings.ToUserSettingsOut()}
}

func (settings *UserSettings) ToUserSettingsOut() *UserSettingsOut {
	return &UserSettingsOut{TrackingDevices: settings.TrackingDevices}
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
		existing.Imei = u.Imei
	}

	if u.Tracking != nil {
		existing.Tracking = u.Tracking
	}

	return existing
}
