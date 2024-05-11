package models

import "time"

type User struct {
	ID           int       `db:"user_id" json:"id"`
	UserName     string    `db:"username" json:"username"`
	HashPassword string    `db:"hashed_password" json:"hashPassword,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt,omitempty"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt,omitempty"`
	Deleted      bool      `db:"deleted" json:"deleted,omitempty"`
}

type UserRegisterInput struct {
	UserName     string `json:"username"`
	HashPassword string `json:"hashPassword"`
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
