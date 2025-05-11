package models

import "time"

type User struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}
