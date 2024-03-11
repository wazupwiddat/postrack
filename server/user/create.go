package user

import (
	"gorm.io/gorm"
)

func Create(db *gorm.DB, u *User) (uint, error) {
	err := db.Create(u).Error
	if err != nil {
		return 0, &EmailDuplicateError{Email: u.Email}
	}
	return u.ID, nil
}
