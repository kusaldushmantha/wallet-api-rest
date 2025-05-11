package models

import "time"

type Wallet struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Balance   float64   `db:"balance"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
