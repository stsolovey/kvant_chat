package models

import "time"

type Message struct {
	ID        int       `json:"id,omitempty"`
	Receiver  string    `json:"receiver,omitempty"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
