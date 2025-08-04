package user

import "errors"


func ValidateUser(u *User) error {
	if !u.Role.IsValid() {
		return errors.New("invalid role")
	}
	if u.Email.String() == "" {
		return errors.New("email required")
	}
	return nil
}
