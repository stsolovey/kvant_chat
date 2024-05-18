package models

import "net"

type Room struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Members map[*User]net.Conn
}
