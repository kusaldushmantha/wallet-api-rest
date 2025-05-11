package responses

import (
	"time"
)

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   error  `json:"error,omitempty"`
}

type WalletBalanceResponse struct {
	WalletID string  `json:"wallet_id"`
	Balance  float64 `json:"balance"`
}

type TransactionHistory struct {
	Type       string    `json:"type"` // deposit, withdrawal, transfer
	Amount     float64   `json:"amount"`
	ToWalletID string    `json:"recipient_wallet_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type TransactionHistoryResponse struct {
	WalletID     string                `json:"wallet_id"`
	Transactions []*TransactionHistory `json:"transactions"`
	Limit        int32                 `json:"limit"`
	Offset       int32                 `json:"offset"`
}

type UserWallet struct {
	UserID    string    `json:"user_id"`
	WalletID  string    `json:"wallet_id"`
	CreatedAt time.Time `json:"created_at"`
}

type UserWalletResponse struct {
	UserWallets []*UserWallet `json:"user_wallets"`
}
