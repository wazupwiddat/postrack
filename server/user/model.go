package user

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           uint   `gorm:"primary_key"`
	Email        string `gorm:"size:100;unique;not null"`
	PasswordHash string `gorm:"size:100"`
}

const (
	UniqueConstraintEmail = "users_email_key"
)

type EmailDuplicateError struct {
	Email string
}

func (e *EmailDuplicateError) Error() string {
	return fmt.Sprintf("Email '%s' already exists", e.Email)
}
