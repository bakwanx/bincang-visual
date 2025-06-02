package models

import (
	"github.com/gofiber/contrib/websocket"
)

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"createAt"`
}

type UserClient struct {
	ID   string          `json:"id"`
	Conn *websocket.Conn `json:"connection"`
}
