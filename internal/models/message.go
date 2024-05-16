package models

import "time"

type Message struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"roomId"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
