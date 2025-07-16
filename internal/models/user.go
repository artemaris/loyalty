package models

import (
	"time"
)

type User struct {
	ID       int64     `json:"id" db:"id"`
	Login    string    `json:"login" db:"login"`
	Password string    `json:"-" db:"password_hash"`
	Created  time.Time `json:"created" db:"created_at"`
}

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
