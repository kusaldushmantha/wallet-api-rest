package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2/log"

	"WalletApp/commons"
	"WalletApp/db"
	"WalletApp/models/responses"
)

type walletServiceV1 struct {
	DB    db.Database
	Cache db.Cache
}

func NewWalletServiceV1(dbClient db.Database, cacheClient db.Cache) WalletService {
	return &walletServiceV1{
		DB:    dbClient,
		Cache: cacheClient,
	}
}

// Deposit performs a deposit transaction into the specified wallet.
//
// It first validates whether the requesting user is authorized to access the wallet.
// Then it enforces idempotency using a cache-based key to prevent duplicate deposits.
// If the request is valid and not a duplicate, it inserts a deposit transaction and returns the updated balance.
//
// Parameters:
//   - ctx: Context used for cancellation and timeout control.
//   - idempotencyKey: Unique key from the client to ensure the operation is idempotent.
//   - account: Wallet ID where the deposit should be made.
//   - amount: Amount to deposit (must be > 0).
//   - userID: ID of the user making the request.
//
// Returns:
//   - A Response struct containing the updated balance, message, HTTP status code, and any error encountered.
func (v *walletServiceV1) Deposit(ctx context.Context, idempotencyKey, account string, amount float64, userID string) responses.Response {
	isAuthorized, err := v.isAuthorized(ctx, account, userID)
	if err != nil {
		return internalError("error while checking wallet ownership", err)
	}
	if !isAuthorized {
		return unauthorizedResponse()
	}

	isDuplicateRequest, err := v.setIdempotency(ctx, idempotencyKey, commons.TransactionTypeDeposit, userID, account, amount)
	if err != nil {
		return internalError("error while checking idempotency for deposit action", err)
	}
	if isDuplicateRequest {
		return duplicateRequestResponse()
	}

	updatedBalance, err := v.DB.InsertTxnAndGetWalletBalance(ctx, account, "", amount, commons.TransactionTypeDeposit)
	if err != nil {
		if errors.Is(err, commons.InsufficientBalanceError) {
			return badRequest(err.Error())
		}
		return internalError("failed to record the deposit transaction", err)
	}

	return successBalanceResponse(account, updatedBalance, "deposit successful")
}

// Withdraw performs a withdrawal transaction from the specified wallet.
//
// It verifies that the user is authorized to access the wallet, ensures the request is not a duplicate
// using an idempotency key, and checks whether the wallet has sufficient balance for the withdrawal.
// If all checks pass, it records the withdrawal transaction and returns the updated wallet balance.
//
// Parameters:
//   - ctx: Context for managing request-scoped values, cancellation, and timeout.
//   - idempotencyKey: Unique client-provided key used to prevent duplicate withdrawals.
//   - account: Wallet ID from which the amount should be withdrawn.
//   - amount: Amount to withdraw (must be > 0).
//   - userID: ID of the user initiating the withdrawal.
//
// Returns:
//   - A Response struct containing the updated balance (if successful), a status message,
//     HTTP status code, and any error that occurred during the operation.
func (v *walletServiceV1) Withdraw(ctx context.Context, idempotencyKey, account string, amount float64, userID string) responses.Response {
	isAuthorized, err := v.isAuthorized(ctx, account, userID)
	if err != nil {
		return internalError("error while checking wallet ownership", err)
	}
	if !isAuthorized {
		return unauthorizedResponse()
	}

	isDuplicateRequest, err := v.setIdempotency(ctx, idempotencyKey, commons.TransactionTypeWithdraw, userID, account, amount)
	if err != nil {
		return internalError("error while checking idempotency for deposit action", err)
	}
	if isDuplicateRequest {
		return duplicateRequestResponse()
	}

	updatedBalance, err := v.DB.InsertTxnAndGetWalletBalance(ctx, account, "", amount, commons.TransactionTypeWithdraw)
	if err != nil {
		if errors.Is(err, commons.InsufficientBalanceError) {
			return badRequest(err.Error())
		}
		return internalError("failed to perform the withdrawal transaction", err)
	}

	return successBalanceResponse(account, updatedBalance, "withdrawal successful")
}

// Transfer performs a fund transfer from one wallet to another.
//
// It ensures the user is authorized to use the source wallet, prevents duplicate requests
// using an idempotency key, verifies sufficient balance in the source wallet, and confirms
// that the recipient wallet exists. If all validations pass, it records the transfer transaction
// and returns the updated balance of the source wallet.
//
// Parameters:
//   - ctx: Context for managing request lifetime and enforcing timeouts or cancellation.
//   - idempotencyKey: Unique key provided by the client to prevent duplicate transfers.
//   - toAccount: Destination wallet ID to which the funds will be transferred.
//   - fromAccount: Source wallet ID from which the funds will be deducted.
//   - amount: Amount of money to transfer (must be > 0).
//   - userID: ID of the user initiating the transfer.
//
// Returns:
//   - A Response struct containing the updated balance (if successful), a status message,
//     HTTP status code, and any error that occurred during the operation.
func (v *walletServiceV1) Transfer(ctx context.Context, idempotencyKey, fromAccount, toAccount string, amount float64, userID string) responses.Response {
	if fromAccount == toAccount {
		return badRequest("cannot transfer between same accounts")
	}
	isAuthorized, err := v.isAuthorized(ctx, fromAccount, userID)
	if err != nil {
		return internalError("error while checking wallet ownership", err)
	}
	if !isAuthorized {
		return unauthorizedResponse()
	}

	isDuplicateRequest, err := v.setIdempotency(ctx, idempotencyKey, commons.TransactionTypeTransfer, userID, fmt.Sprintf("%s-%s", fromAccount, toAccount), amount)
	if err != nil {
		return internalError("error while checking idempotency for transfer action", err)
	}
	if isDuplicateRequest {
		return duplicateRequestResponse()
	}

	if exists, err := v.walletExists(ctx, toAccount); err != nil {
		return internalError("failed to verify recipient wallet", err)
	} else if !exists {
		return badRequest("recipient wallet id does not exist")
	}

	updatedBalance, err := v.DB.InsertTxnAndGetWalletBalance(ctx, fromAccount, toAccount, amount, commons.TransactionTypeTransfer)
	if err != nil {
		if errors.Is(err, commons.InsufficientBalanceError) {
			return badRequest(err.Error())
		}
		return internalError("failed to record the transfer transaction", err)
	}

	return successBalanceResponse(fromAccount, updatedBalance, "transfer successful")
}

// GetBalance retrieves the current balance of a specified wallet.
//
// It first verifies that the requesting user is authorized to access the wallet.
// If authorized, it fetches the wallet's balance from the database and returns it
// in the response.
//
// Parameters:
//   - ctx: Context used to manage request timeout and cancellation.
//   - walletID: ID of the wallet whose balance is being requested.
//   - userID: ID of the user making the request.
//
// Returns:
//   - A Response struct containing the wallet balance (if successful), a status message,
//     HTTP status code, and any error that occurred during the operation.
func (v *walletServiceV1) GetBalance(ctx context.Context, walletID, userID string) responses.Response {
	isAuthorized, err := v.isAuthorized(ctx, walletID, userID)
	if err != nil {
		return internalError("error while checking wallet ownership", err)
	}
	if !isAuthorized {
		return unauthorizedResponse()
	}

	dbBalance, err := v.DB.GetWalletBalance(ctx, walletID)
	if err != nil {
		return internalError("failed to fetch balance", err)
	}

	return successBalanceResponse(walletID, dbBalance, "success")
}

// GetTransactionHistory retrieves a paginated list of transactions for a given wallet.
//
// The method first ensures the requesting user is authorized to access the specified wallet.
// If authorized, it queries the database for transactions associated with the wallet using
// the provided limit and offset for pagination. The transactions are then mapped to a
// summarized response format.
//
// Parameters:
//   - ctx: Context for managing timeouts and cancellations.
//   - walletID: The ID of the wallet whose transactions are being retrieved.
//   - userID: The ID of the user making the request.
//   - limit: The maximum number of transactions to return.
//   - offset: The number of transactions to skip before starting to return results.
//
// Returns:
//   - A Response struct containing a list of transaction summaries (if successful), a status message,
//     HTTP status code, and any error that occurred during the operation.
func (v *walletServiceV1) GetTransactionHistory(ctx context.Context, walletID, userID string, limit, offset int32) responses.Response {
	isAuthorized, err := v.isAuthorized(ctx, walletID, userID)
	if err != nil {
		return internalError("error while checking wallet ownership", err)
	}
	if !isAuthorized {
		return unauthorizedResponse()
	}

	transactions, err := v.DB.GetTransactions(ctx, walletID, limit, offset)
	if err != nil {
		return internalError("failed to fetch transactions", err)
	}

	summary := make([]*responses.TransactionHistory, 0, len(transactions))
	for _, tx := range transactions {
		summary = append(summary, &responses.TransactionHistory{
			Type:       tx.Type,
			Amount:     tx.Amount,
			ToWalletID: tx.ToWalletID.String,
			CreatedAt:  tx.CreatedAt,
		})
	}

	return getResponse(responses.TransactionHistoryResponse{
		WalletID:     walletID,
		Transactions: summary,
		Limit:        limit,
		Offset:       offset,
	}, "transaction history retrieval successful", http.StatusOK, nil)
}

func (v *walletServiceV1) isAuthorized(ctx context.Context, walletID, userID string) (bool, error) {
	isOwner, err := v.DB.CheckWalletOwner(ctx, walletID, userID)
	if err != nil {
		log.Errorf("ownership check failed for wallet '%s' for user %s : %v", walletID, userID, err)
		return false, err
	}
	return isOwner, err
}

func (v *walletServiceV1) setIdempotency(ctx context.Context, key string, action commons.TransactionType, userID, walletID string, amount float64) (bool, error) {
	cacheKey := fmt.Sprintf("idm-%s-%s-%s-%s-%f", string(action), key, userID, walletID, amount)
	ok, err := v.Cache.SetWithExpirationIfKeyIsNotSet(ctx, cacheKey, cacheKey, commons.IdempotencyCacheTTL)
	if err != nil {
		log.Errorf("idempotency check failed for key '%s': %v", cacheKey, err)
		return false, err
	}
	// SetWithExpirationIfKeyIsNotSet returns true if the key is set, so here we have to return false
	return !ok, err
}

func (v *walletServiceV1) walletExists(ctx context.Context, walletID string) (bool, error) {
	w, err := v.DB.GetWallet(ctx, walletID)
	return w != nil, err
}
