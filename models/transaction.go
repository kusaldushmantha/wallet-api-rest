package models

import (
	"database/sql"
	"time"
)

type Transaction struct {
	ID         string         `db:"id"`
	WalletID   string         `db:"from_wallet_id"`
	Type       string         `db:"type"` // deposit, withdrawal, transfer
	Amount     float64        `db:"amount"`
	ToWalletID sql.NullString `db:"to_wallet_id,omitempty"`
	CreatedAt  time.Time      `db:"created_at"`
}
