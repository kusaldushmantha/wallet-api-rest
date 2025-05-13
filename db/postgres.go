package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"

	"WalletApp/commons"
	"WalletApp/models"
)

type postgresDB struct {
	db *sqlx.DB
}

func NewPostgreSQLDB(db *sqlx.DB) Database {
	return &postgresDB{db: db}
}

func (p *postgresDB) GetWallet(ctx context.Context, walletID string) (*models.Wallet, error) {
	dbCtx, cancel := withTimeout(ctx)
	defer cancel()

	query := `SELECT * FROM wallets WHERE id = $1 LIMIT 1;`
	var exists []*models.Wallet
	err := p.db.SelectContext(dbCtx, &exists, query, walletID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // wallet does not exist
		}
		return nil, err // actual error
	}
	return exists[0], err
}

func (p *postgresDB) CheckWalletOwner(ctx context.Context, walletID string, userID string) (bool, error) {
	dbCtx, cancel := withTimeout(ctx)
	defer cancel()

	query := `SELECT 1 FROM wallets WHERE id = $1 AND user_id = $2;`
	var exists int
	err := p.db.GetContext(dbCtx, &exists, query, walletID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // wallet does not belong to this user
		}
		return false, err // actual error
	}
	return true, nil
}

func (p *postgresDB) GetWalletBalance(ctx context.Context, walletID string) (float64, error) {
	dbCtx, cancel := withTimeout(ctx)
	defer cancel()

	query := `SELECT balance FROM wallets WHERE id = $1;`
	var balance float64
	err := p.db.GetContext(dbCtx, &balance, query, walletID)
	if err != nil {
		return -1, err
	}
	return balance, nil
}

func (p *postgresDB) GetTransactions(ctx context.Context, walletID string, limit, offset int32) ([]*models.Transaction, error) {
	dbCtx, cancel := withTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, from_wallet_id, to_wallet_id, type, amount, created_at
		FROM transactions
		WHERE from_wallet_id = $1 OR to_wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;
	`
	var transactions []*models.Transaction
	err := p.db.SelectContext(dbCtx, &transactions, query, walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	return transactions, err
}

func (p *postgresDB) InsertTxnAndGetWalletBalance(ctx context.Context, fromAccount string, toAccount string, amount float64, trsType commons.TransactionType) (float64, error) {
	// Transactional operation to perform atomic update in transaction table and wallet table
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		log.Errorf("error creating a transactional context: %v", err)
		return -1, err
	}

	// Transaction roll back and logging
	defer func() {
		now := time.Now()
		if p := recover(); p != nil {
			if err := tx.Rollback(); err != nil {
				log.Errorf("transaction rollback failed after panic. from: %s | to: %s | amount: %f | type: %s | time: %s", fromAccount, toAccount, amount, trsType, now)
			} else {
				log.Infof("transaction rollback succeeded after panic. from: %s | to: %s | amount: %f | type: %s | time: %s", fromAccount, toAccount, amount, trsType, now)
			}
			return // silently absorb the panic
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Errorf("transaction rollback failed. from: %s | to: %s | amount: %f | type: %s | time: %s", fromAccount, toAccount, amount, trsType, now)
			} else {
				log.Infof("transaction rollback succeeded. from: %s | to: %s | amount: %f | type: %s | time: %s", fromAccount, toAccount, amount, trsType, now)
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			log.Errorf("transaction commit failed. from: %s | to: %s | amount: %f | type: %s | time: %s", fromAccount, toAccount, amount, trsType, now)
			err = commitErr
		} else {
			log.Infof("transaction committed successfully. from: %s | to: %s | amount: %f | type: %s | time: %s", fromAccount, toAccount, amount, trsType, now)
		}
	}()

	var balance float64
	// Check the balance of the source wallet before a withdrawal or transfer
	if trsType == commons.TransactionTypeWithdraw || trsType == commons.TransactionTypeTransfer {
		query := "SELECT balance FROM wallets WHERE id = $1 FOR UPDATE"
		readCtx, cancel := withTimeout(ctx)
		defer cancel()
		row := tx.QueryRowContext(readCtx, query, fromAccount)
		err = row.Scan(&balance)
		if err != nil {
			log.Errorf("transaction select error. from: %s | to: %s, amount: %f | type: %s | time: %s | error: %+v", fromAccount, toAccount, amount, trsType, time.Now(), err)
			return -1, err
		}
		if balance < amount {
			return -1, commons.InsufficientBalanceError
		}
	}

	// Insert transaction record
	err = p.addTransactionRecord(tx, ctx, fromAccount, toAccount, amount, trsType)
	if err != nil {
		log.Errorf("transaction insert error. from: %s | to: %s, amount: %f | type: %s | time: %s | error: %+v", fromAccount, toAccount, amount, trsType, time.Now(), err)
		return -1, err
	}

	senderAmount := amount
	receiverAmount := amount
	// If it is not a deposit, then the amount should be negative for sender
	if trsType == commons.TransactionTypeWithdraw || trsType == commons.TransactionTypeTransfer {
		senderAmount = -amount
	}
	// Update the receiver wallet if the transaction is a transfer
	if trsType == commons.TransactionTypeTransfer {
		err = p.updateWalletBalance(tx, ctx, toAccount, receiverAmount)
		if err != nil {
			log.Errorf("update receiver wallet error. from: %s | to: %s, amount: %f | type: %s | time: %s | error: %+v", fromAccount, toAccount, amount, trsType, time.Now(), err)
			return 0, err
		}
	}
	// Update sender balance and return it
	balance, err = p.updateAndGetWalletBalance(tx, ctx, fromAccount, senderAmount)
	if err != nil {
		log.Errorf("update sender wallet error. from: %s | to: %s, amount: %f | type: %s | time: %s | error: %+v", fromAccount, toAccount, amount, trsType, time.Now(), err)
		return balance, err
	}
	return balance, err
}

func (p *postgresDB) addTransactionRecord(tx *sql.Tx, ctx context.Context, fromAccount string, toAccount string, amount float64, trsType commons.TransactionType) error {
	var query string
	var args []any
	if trsType == commons.TransactionTypeDeposit || trsType == commons.TransactionTypeWithdraw {
		// If it's a deposit or withdrawal, the fromAccount is sufficient
		query = `INSERT INTO transactions (from_wallet_id, type, amount) VALUES ($1, $2, $3);`
		args = []any{fromAccount, trsType, amount}
	} else {
		query = `INSERT INTO transactions (from_wallet_id, to_wallet_id, type, amount) VALUES ($1, $2, $3, $4);`
		args = []any{fromAccount, toAccount, trsType, amount}
	}

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	return nil
}

func (p *postgresDB) updateAndGetWalletBalance(tx *sql.Tx, ctx context.Context, walletID string, amount float64) (float64, error) {
	query := `
		UPDATE wallets
		SET balance = balance + $1,
			updated_at = NOW()
		WHERE id = $2
		RETURNING balance;
    `
	var newBalance float64
	err := tx.QueryRowContext(ctx, query, amount, walletID).Scan(&newBalance)
	if err != nil {
		return 0, fmt.Errorf("failed to update and fetch wallet balance: %w", err)
	}
	return newBalance, nil
}

func (p *postgresDB) updateWalletBalance(tx *sql.Tx, ctx context.Context, walletID string, amount float64) error {
	query := `
		UPDATE wallets
		SET balance = balance + $1,
			updated_at = NOW()
		WHERE id = $2;
    `
	_, err := tx.ExecContext(ctx, query, amount, walletID)
	if err != nil {
		return fmt.Errorf("failed to update and fetch wallet balance: %w", err)
	}
	return nil
}

func (p *postgresDB) CreateUserWallet(ctx context.Context) (*models.Wallet, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var userID string
	userQuery := `INSERT INTO users DEFAULT VALUES RETURNING id`
	err = tx.QueryRowContext(ctx, userQuery).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	var walletID string
	var createdAt time.Time

	walletQuery := `INSERT INTO wallets (user_id) VALUES ($1) RETURNING id, created_at`
	err = tx.QueryRowContext(ctx, walletQuery, userID).Scan(&walletID, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	wallet := &models.Wallet{
		ID:        walletID,
		UserID:    userID,
		CreatedAt: createdAt,
	}

	return wallet, nil
}

func (p *postgresDB) GetWalletUsers(ctx context.Context) ([]*models.Wallet, error) {
	dbCtx, cancel := withTimeout(ctx)
	defer cancel()

	query := `SELECT id, user_id, created_at FROM wallets;`
	var walletUsers []*models.Wallet
	err := p.db.SelectContext(dbCtx, &walletUsers, query)
	if err != nil {
		return nil, err
	}
	return walletUsers, err
}

func withTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, commons.DBOperationTimeout)
}
