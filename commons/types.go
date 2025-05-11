package commons

import "errors"

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdrawal"
	TransactionTypeTransfer TransactionType = "transfer"
)

var InsufficientBalanceError = errors.New("insufficient balance error")
