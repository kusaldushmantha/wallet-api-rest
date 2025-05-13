package db_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"WalletApp/commons"
	"WalletApp/db"
)

func TestInsertTxnAndGetWalletBalance(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "walletdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		t.Error(err)
	}

	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Error(err)
	}

	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/walletdb?sslmode=disable", host, mappedPort.Port())

	maxRetries := 10
	var sqlxDB *sqlx.DB
	for i := 0; i < maxRetries; i++ {
		sqlxDB, err = sqlx.Connect("postgres", connStr)
		if err == nil {
			break
		}
		log.Infof("postgresql test container is not running yet. backing off %d/%d", i+1, maxRetries)
		time.Sleep(time.Duration(i+1) * time.Second) // exponential backoff
	}
	if err != nil {
		t.Fatal("error while connecting to the postgres test container")
	}

	schemaSQL, err := os.ReadFile("../migration/init.sql")
	if err != nil {
		t.Fatal("error while reading the seed data")
	}
	_, err = sqlxDB.Exec(string(schemaSQL))
	if err != nil {
		t.Fatal("error while executing the seed data")
	}

	pdb := db.NewPostgreSQLDB(sqlxDB)

	t.Cleanup(func() {
		err := pgContainer.Terminate(ctx)
		if err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	t.Run("successful deposit", func(t *testing.T) {
		// Perform the deposit and get the wallet balance.
		balance, err := pdb.InsertTxnAndGetWalletBalance(ctx, "7dbacf5d-3099-4a66-ad3d-2fee93970017", "", 50.0, commons.TransactionTypeDeposit)
		assert.Nil(t, err)
		assert.Equal(t, 50.0, balance)

		// Verify that the transaction has been recorded in the 'transactions' table.
		var txnCount int
		err = sqlxDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM transactions WHERE from_wallet_id = $1 AND type = $2", "7dbacf5d-3099-4a66-ad3d-2fee93970017", commons.TransactionTypeDeposit).Scan(&txnCount)
		assert.Nil(t, err)
		assert.Equal(t, 1, txnCount, "Expected one deposit transaction to be recorded.")

		// Verify that the wallet balance in the 'wallets' table has been updated.
		var walletBalance float64
		err = sqlxDB.QueryRowContext(ctx, "SELECT balance FROM wallets WHERE id = $1", "7dbacf5d-3099-4a66-ad3d-2fee93970017").Scan(&walletBalance)
		assert.Nil(t, err)
		assert.Equal(t, 50.0, walletBalance, "Wallet balance should reflect the deposit.")
	})

	t.Run("withdraw more than balance should rollback", func(t *testing.T) {
		_, err := sqlxDB.Exec("UPDATE wallets SET balance=150.0 WHERE id=$1", "7dbacf5d-3099-4a66-ad3d-2fee93970017")
		assert.Nil(t, err)
		_, err = pdb.InsertTxnAndGetWalletBalance(ctx, "7dbacf5d-3099-4a66-ad3d-2fee93970017", "", 1000.0, commons.TransactionTypeWithdraw)
		assert.NotNil(t, err)

		var balance float64
		err = sqlxDB.QueryRow(`SELECT balance FROM wallets WHERE id = $1`, "7dbacf5d-3099-4a66-ad3d-2fee93970017").Scan(&balance)
		assert.Nil(t, err)
		assert.Equal(t, 150.0, balance) // unchanged due to rollback
	})

	t.Run("transfer to non-existent wallet should rollback", func(t *testing.T) {
		_, err := sqlxDB.Exec("UPDATE wallets SET balance=150.0 WHERE id=$1", "7dbacf5d-3099-4a66-ad3d-2fee93970017")
		assert.Nil(t, err)
		_, err = pdb.InsertTxnAndGetWalletBalance(ctx, "7dbacf5d-3099-4a66-ad3d-2fee93970017", "710acea9-142e-4f72-8416-53f032134f08", 20.0, commons.TransactionTypeTransfer)
		assert.NotNil(t, err)

		var balance float64
		err = sqlxDB.QueryRow(`SELECT balance FROM wallets WHERE id = '7dbacf5d-3099-4a66-ad3d-2fee93970017'`).Scan(&balance)
		assert.Nil(t, err)
		assert.Equal(t, 150.0, balance)
	})
}
