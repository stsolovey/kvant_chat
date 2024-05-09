package models

import "time"

type User struct {
	ID           int       `db:"user_id" json:"id"`
	Name         string    `db:"name" json:"name"`
	HashPassword string    `db:"hashed_password" json:"hashPassword,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt,omitempty"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt,omitempty"`
	Deleted      bool      `db:"deleted" json:"deleted,omitempty"`
}

type UserCreateInput struct {
	Name         string `json:"name"`
	HashPassword string `json:"hashPassword"`
}

type UserUpdateInput struct {
	Name         string `json:"name"`
	HashPassword string `json:"hashPassword"`
}

type FeedUsersRequest struct {
	Offset       int    `json:"offset"`
	ItemsPerPage int    `json:"itemsPerPage"`
	Sorting      string `json:"sorting"`
	Descending   bool   `json:"descending"`
	Text         string `json:"text"`
}