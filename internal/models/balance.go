package models

type Balance struct {
	UserID    int64   `json:"-" db:"user_id"`
	Current   float64 `json:"current" db:"current"`
	Withdrawn float64 `json:"withdrawn" db:"withdrawn"`
}
