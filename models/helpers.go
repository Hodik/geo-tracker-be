package models

func containsUser(users []*User, user *User) bool {
	for _, u := range users {
		if u.ID == user.ID {
			return true
		}
	}
	return false
}
