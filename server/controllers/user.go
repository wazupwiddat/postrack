package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"

	"github.com/wazupwiddat/postrack/server/user"
	"github.com/wazupwiddat/postrack/server/user/login"
	"github.com/wazupwiddat/postrack/server/user/signup"
)

type LoginReqest struct {
	Email    string
	Password string
}

func (c Controller) HandleSignup(w http.ResponseWriter, r *http.Request) {
	var req LoginReqest
	json.NewDecoder(r.Body).Decode(&req)

	// validate the request
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and Password is required", http.StatusBadRequest)
		return
	}

	res, err := signup.SignUp(c.db, &signup.Request{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil || res.Id == 0 {
		http.Error(w, "Failed to create user", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (c Controller) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginReqest
	json.NewDecoder(r.Body).Decode(&req)

	// validate the request
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and Password is required", http.StatusBadRequest)
		return
	}

	res, err := login.Login(c.db, &login.Request{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		if _, ok := err.(*user.EmailNotExistsError); ok {
			http.Error(w, "Invalid login", http.StatusNotFound)
			return
		}
		if _, ok := err.(*login.PasswordMismatchError); ok {
			http.Error(w, "Invalid login", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if res == nil {
		http.Error(w, "Invalid login", http.StatusNotFound)
		return
	}

	// create a JWT token"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    res.User.ID,
		"email": res.User.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	// sign the JWT token with a secret key
	tokenString, err := token.SignedString([]byte(c.cfg.JWT.Secret))
	if err != nil {
		http.Error(w, "Failed to create JWT token", http.StatusInternalServerError)
		return
	}

	// send the JWT token as a response
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

func userFromRequestContext(r *http.Request, db *gorm.DB) (*user.User, error) {
	userID, ok := r.Context().Value("id").(float64)
	if !ok {
		log.Println(r.Context().Value("id"))
		return nil, fmt.Errorf("No user ID on Request context")
	}
	log.Println("UserID: ", userID)
	u, err := user.FindByID(db, uint(userID))
	if err != nil {
		return nil, fmt.Errorf("User with ID:%f does not exist", userID)
	}
	return u, nil
}
