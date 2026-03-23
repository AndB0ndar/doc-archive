package models

import "time"

type UserDB struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

func (UserDB) TableName() string {
	return "users"
}
