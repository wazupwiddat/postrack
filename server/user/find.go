package user

import (
	"errors"

	"gorm.io/gorm"
)

type EmailNotExistsError struct{}

func (*EmailNotExistsError) Error() string {
	return "email not exists"
}

type UserIDDoesNotExistError struct{}

func (*UserIDDoesNotExistError) Error() string {
	return "user by id does not exist"
}

func FindByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	res := db.Find(&user, &User{Email: email})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, &EmailNotExistsError{}
	}
	return &user, nil
}

func FindByID(db *gorm.DB, id uint) (*User, error) {
	var user User
	res := db.Find(&user, &User{ID: id})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, &UserIDDoesNotExistError{}
	}
	return &user, nil
}
