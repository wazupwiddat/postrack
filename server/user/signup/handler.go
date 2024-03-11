package signup

import (
	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"

	"github.com/wazupwiddat/postrack/server/user"
)

type Request struct {
	Email    string
	Password string
}

type Response struct {
	Id uint
}

func SignUp(db *gorm.DB, req *Request) (*Response, error) {
	// passwordHash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &user.User{
		Email:        req.Email,
		PasswordHash: string(passwordHash),
	}
	id, err := user.Create(db, newUser)
	if err != nil {
		return nil, err
	}
	return &Response{Id: id}, err
}
