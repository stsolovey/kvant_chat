package models

import (
	"net"
	"time"
)

type User struct {
	ID           int       `db:"user_id" json:"id"`
	UserName     string    `db:"username" json:"username"`
	HashPassword string    `db:"hashed_password" json:"hashPassword,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt,omitempty"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt,omitempty"`
	Deleted      bool      `db:"deleted" json:"deleted,omitempty"`
	Conn         net.Conn  `json:"-"`
}

type UserResponse struct {
	ID        int       `json:"id"`
	UserName  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type UserRegisterInput struct {
	UserName     string `json:"username"`
	HashPassword string `json:"password"`
}

type UserLoginInput struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type UserUpdateInput struct {
	UserName     string `json:"username"`
	HashPassword string `json:"hashPassword"`
}

type FeedUsersRequest struct {
	Offset       int    `json:"offset"`
	ItemsPerPage int    `json:"itemsPerPage"`
	Sorting      string `json:"sorting"`
	Descending   bool   `json:"descending"`
	Text         string `json:"text"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseData struct {
	Token string `json:"token"`
	User  struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"user"`
}
