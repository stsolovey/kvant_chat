package models

import "time"

type Message struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"room_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
