package db

import (
	"context"
	"time"

	"WalletApp/commons"
	"WalletApp/models"
)

//go:generate mockgen -source=types.go -destination=mocks/types.go -package=mocks
type Database interface {
	// Wallet Operations
	GetWallet(ctx context.Context, walletID string) (*models.Wallet, error)
	GetWalletBalance(ctx context.Context, walletID string) (float64, error)
	GetTransactions(ctx context.Context, walletID string, limit, offset int32) ([]*models.Transaction, error)
	InsertTxnAndGetWalletBalance(ctx context.Context, fromAccount string, toAccount string, amount float64, trsType commons.TransactionType) (balance float64, error error)
	CheckWalletOwner(ctx context.Context, walletID string, userID string) (bool, error)

	// User management operations
	CreateUserWallet(ctx context.Context) (*models.Wallet, error)
	GetWalletUsers(ctx context.Context) ([]*models.Wallet, error)
}

type Cache interface {
	SetWithExpirationIfKeyIsNotSet(ctx context.Context, key string, value string, duration time.Duration) (bool, error)
}
