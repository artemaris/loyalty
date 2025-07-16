package models

import "time"

type Withdrawal struct {
	ID          int64     `json:"-" db:"id"`
	UserID      int64     `json:"-" db:"user_id"`
	Order       string    `json:"order" db:"order_number"`
	Sum         float64   `json:"sum" db:"sum"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}

type WithdrawalRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
