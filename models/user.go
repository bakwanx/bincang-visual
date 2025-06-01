package models

type User struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Username string `json:"username"`
}
