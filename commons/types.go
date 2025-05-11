package commons

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdrawal"
	TransactionTypeTransfer TransactionType = "transfer"
)
