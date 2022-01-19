package models

type User struct {
	ID       string `json:"id,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Users []*User

// Users of our api
var users = Users{}
