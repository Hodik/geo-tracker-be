package schemas

import "github.com/Hodik/geo-tracker-be/models"

type UpdateUser struct {
	Name *string `json:"name"`
}

type UpdateMe struct {
	User *UpdateUser `json:"user"`
}

type UserProfile struct {
	User     *models.User         `json:"user"`
	Settings *models.UserSettings `json:"settings"`
}

func (u *UpdateMe) ToUser(existing *models.User) {
	if u.User == nil {
		return
	}

	if u.User.Name != nil {
		existing.Name = u.User.Name
	}
}

func ToUserProfile(u *models.User, settings *models.UserSettings) *UserProfile {
	return &UserProfile{User: u, Settings: settings}
}
