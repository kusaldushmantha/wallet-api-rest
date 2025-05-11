package requests

type BaseTransactionRequest struct {
	Amount           float64 `json:"amount"`
	IdempotencyToken string  `json:"idempotency_token"`
}
type DepositRequest struct {
	BaseTransactionRequest
}
type WithdrawRequest struct {
	BaseTransactionRequest
}
type TransferRequest struct {
	BaseTransactionRequest
	ToAccount string `json:"recipient_wallet_id"`
}
