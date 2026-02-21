package models

//map[string]map[string]string

type Room struct {
	RoomId    string          `json:"roomId"`
	CreatedAt string          `json:"createdAt"`
	Users     map[string]User `json:"users"`
}
