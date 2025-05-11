package services

import (
	"context"

	"WalletApp/models/responses"
)

//go:generate mockgen -source=types.go -destination=mocks/types.go -package=mocks
type WalletService interface {
	Deposit(ctx context.Context, idempotencyKey string, toAccount string, amount float64, userID string) responses.Response
	Withdraw(ctx context.Context, idempotencyKey string, fromAccount string, amount float64, userID string) responses.Response
	Transfer(ctx context.Context, idempotencyKey string, toAccount string, fromAccount string, amount float64, userID string) responses.Response
	GetBalance(ctx context.Context, walletID string, userID string) responses.Response
	GetTransactionHistory(ctx context.Context, walletID string, userID string, limit int32, offset int32) responses.Response
}

type UserService interface {
	CreateUser(ctx context.Context) responses.Response
	GetUsers(ctx context.Context) responses.Response
}
