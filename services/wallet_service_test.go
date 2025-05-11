package services

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	mocks2 "WalletApp/db/mocks"
	"WalletApp/models"
	"WalletApp/models/responses"
)

func TestWalletServiceV1_Deposit(t *testing.T) {
	idempotencyKey := "test-key"
	walletID := "1234"
	amount := 120.0
	userID := "2222"

	t.Run("should return bad request response if the user is not authorized for the wallet", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
		resp := walletService.Deposit(context.Background(), idempotencyKey, walletID, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Status)
		assert.Equal(t, "unauthorized access to the wallet", resp.Message)
	})

	t.Run("should return internal server error upon errors when checking the wallet ownership", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("test db error"))
		resp := walletService.Deposit(context.Background(), idempotencyKey, walletID, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "error while checking wallet ownership", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return conflict response if there is an idempotency token already", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
		resp := walletService.Deposit(context.Background(), idempotencyKey, walletID, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusConflict, resp.Status)
		assert.Equal(t, "duplicate request: idempotency key already exists", resp.Message)
	})

	t.Run("should return internal server error response upon errors when checking the idempotency token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("test db error"))
		resp := walletService.Deposit(context.Background(), idempotencyKey, walletID, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "error while checking idempotency for deposit action", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return success response after successfully recording the transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().InsertTxnAndGetWalletBalance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(amount, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

		resp := walletService.Deposit(context.Background(), idempotencyKey, walletID, amount, userID)

		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.Status)
		assert.Equal(t, "deposit successful", resp.Message)
		if userWalletResp, ok := resp.Data.(responses.WalletBalanceResponse); ok {
			assert.Equal(t, amount, userWalletResp.Balance)
			assert.Equal(t, walletID, userWalletResp.WalletID)
		} else {
			assert.Fail(t, "cannot cast the response")
		}
	})

	t.Run("should return an error if the transaction recording is unsuccessful", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().InsertTxnAndGetWalletBalance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(-1.0, errors.New("test db error"))
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

		resp := walletService.Deposit(context.Background(), idempotencyKey, walletID, amount, userID)

		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "failed to record the deposit transaction", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})
}

func TestWalletServiceV1_Withdraw(t *testing.T) {
	idempotencyKey := "test-key"
	walletID := "1234"
	amount := 120.0
	userID := "2222"

	t.Run("should return bad request response if the user is not authorized for the wallet", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
		resp := walletService.Withdraw(context.Background(), idempotencyKey, walletID, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Status)
		assert.Equal(t, "unauthorized access to the wallet", resp.Message)
	})

	t.Run("should return internal server error upon errors when checking the wallet ownership", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("test db error"))
		resp := walletService.Withdraw(context.Background(), idempotencyKey, walletID, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "error while checking wallet ownership", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return conflict response if there is an idempotency token already", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
		resp := walletService.Withdraw(context.Background(), idempotencyKey, walletID, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusConflict, resp.Status)
		assert.Equal(t, "duplicate request: idempotency key already exists", resp.Message)
	})

	t.Run("should return internal server error response upon errors when checking the idempotency token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("test db error"))
		resp := walletService.Withdraw(context.Background(), idempotencyKey, walletID, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "error while checking idempotency for deposit action", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return internal server error response upon errors when checking the account balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(-1.0, errors.New("test db error"))
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		resp := walletService.Withdraw(context.Background(), idempotencyKey, walletID, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "failed to check balance", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return bad request response upon if there is insufficient funds in the wallet", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(amount-1, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		resp := walletService.Withdraw(context.Background(), idempotencyKey, walletID, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusBadRequest, resp.Status)
		assert.Equal(t, "insufficient funds to withdraw", resp.Message)
	})

	t.Run("should return success response after successfully recording the transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(amount, nil)
		databaseMock.EXPECT().InsertTxnAndGetWalletBalance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(amount, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

		resp := walletService.Withdraw(context.Background(), idempotencyKey, walletID, amount, userID)

		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.Status)
		assert.Equal(t, "withdrawal successful", resp.Message)
		if userWalletResp, ok := resp.Data.(responses.WalletBalanceResponse); ok {
			assert.Equal(t, amount, userWalletResp.Balance)
			assert.Equal(t, walletID, userWalletResp.WalletID)
		} else {
			assert.Fail(t, "cannot cast the response")
		}
	})

	t.Run("should return an error if the transaction recording is unsuccessful", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(amount, nil)
		databaseMock.EXPECT().InsertTxnAndGetWalletBalance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(-1.0, errors.New("test db error"))
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

		resp := walletService.Withdraw(context.Background(), idempotencyKey, walletID, amount, userID)

		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "failed to record the withdrawal transaction", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})
}

func TestWalletServiceV1_Transfer(t *testing.T) {
	idempotencyKey := "test-key"
	fromAccount := "1234"
	toAccount := "4321"
	amount := 120.0
	userID := "2222"

	t.Run("should return bad request response if the to account and from account is same", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, fromAccount, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusBadRequest, resp.Status)
		assert.Equal(t, "cannot transfer between same accounts", resp.Message)
	})

	t.Run("should return bad request response if the user is not authorized for the wallet", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, toAccount, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Status)
		assert.Equal(t, "unauthorized access to the wallet", resp.Message)
	})

	t.Run("should return internal server error upon errors when checking the wallet ownership", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("test db error"))
		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, toAccount, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "error while checking wallet ownership", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return conflict response if there is an idempotency token already", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, toAccount, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusConflict, resp.Status)
		assert.Equal(t, "duplicate request: idempotency key already exists", resp.Message)
	})

	t.Run("should return internal server error response upon errors when checking the idempotency token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("test db error"))
		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, toAccount, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "error while checking idempotency for transfer action", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return internal server error response upon errors when checking the account balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(-1.0, errors.New("test db error"))
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, toAccount, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "failed to check balance", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return internal server error response upon errors when checking the recipient wallet existence", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Return(nil, errors.New("test db error"))
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(amount+1, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, toAccount, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "failed to verify recipient wallet", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return bad request response if the recipient wallet id does not exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Return(nil, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(amount+1, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, toAccount, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusBadRequest, resp.Status)
		assert.Equal(t, "recipient wallet id does not exist", resp.Message)
	})

	t.Run("should return bad request response upon if there is insufficient funds in the wallet", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(amount-1, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, toAccount, amount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusBadRequest, resp.Status)
		assert.Equal(t, "insufficient funds to transfer", resp.Message)
	})

	t.Run("should return success response after successfully recording the transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(amount, nil)
		databaseMock.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Return(&models.Wallet{}, nil)
		databaseMock.EXPECT().InsertTxnAndGetWalletBalance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(amount, nil)
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, toAccount, amount, userID)

		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.Status)
		assert.Equal(t, "transfer successful", resp.Message)
		if userWalletResp, ok := resp.Data.(responses.WalletBalanceResponse); ok {
			assert.Equal(t, amount, userWalletResp.Balance)
			assert.Equal(t, fromAccount, userWalletResp.WalletID)
		} else {
			assert.Fail(t, "cannot cast the response")
		}
	})

	t.Run("should return an error if the transaction recording is unsuccessful", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		databaseMock := mocks2.NewMockDatabase(ctrl)
		cacheMock := mocks2.NewMockCache(ctrl)
		walletService := walletServiceV1{DB: databaseMock, Cache: cacheMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(amount, nil)
		databaseMock.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Return(&models.Wallet{}, nil)
		databaseMock.EXPECT().InsertTxnAndGetWalletBalance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(-1.0, errors.New("test db error"))
		cacheMock.EXPECT().SetWithExpirationIfKeyIsNotSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

		resp := walletService.Transfer(context.Background(), idempotencyKey, fromAccount, toAccount, amount, userID)

		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "failed to record the transfer transaction", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})
}

func TestWalletServiceV1_GetBalance(t *testing.T) {
	fromAccount := "1234"
	userID := "2222"
	amount := 120.0

	t.Run("should return bad request response if the user is not authorized for the wallet", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
		resp := walletService.GetBalance(context.Background(), fromAccount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Status)
		assert.Equal(t, "unauthorized access to the wallet", resp.Message)
	})

	t.Run("should return internal server error upon errors when checking the wallet ownership", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("test db error"))
		resp := walletService.GetBalance(context.Background(), fromAccount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "error while checking wallet ownership", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return internal server error upon errors when retrieving the balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(-1.0, errors.New("test db error"))
		resp := walletService.GetBalance(context.Background(), fromAccount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "failed to fetch balance", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return success response after retrieving the balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetWalletBalance(gomock.Any(), gomock.Any()).Return(amount, nil)
		resp := walletService.GetBalance(context.Background(), fromAccount, userID)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.Status)
		assert.Equal(t, "success", resp.Message)
	})

}

func TestWalletServiceV1_GetTransactionHistory(t *testing.T) {
	userID := "2222"
	walletID := "1234567"

	t.Run("should return bad request response if the user is not authorized for the wallet", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
		resp := walletService.GetTransactionHistory(context.Background(), walletID, userID, 10, 0)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Status)
		assert.Equal(t, "unauthorized access to the wallet", resp.Message)
	})

	t.Run("should return internal server error upon errors when checking the wallet ownership", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("test db error"))
		resp := walletService.GetTransactionHistory(context.Background(), walletID, userID, 10, 0)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "error while checking wallet ownership", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return internal server error upon errors when retrieving the transaction history", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetTransactions(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("test db error"))
		resp := walletService.GetTransactionHistory(context.Background(), walletID, userID, 10, 0)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusInternalServerError, resp.Status)
		assert.Equal(t, "failed to fetch transactions", resp.Message)
		assert.Equal(t, "test db error", resp.Error.Error())
	})

	t.Run("should return success response after retrieving the balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		transactions := []*models.Transaction{
			{
				WalletID:  "aa4d319f-2b7a-436e-a582-b862fc930e2b",
				Type:      "deposit",
				Amount:    100,
				CreatedAt: time.Now(),
			},
			{
				WalletID:  "d8e02c06-339a-49b2-8ffd-310fd2ec1947",
				Type:      "withdrawal",
				Amount:    20,
				CreatedAt: time.Now(),
			},
			{
				WalletID:   "a5fd3b19-8caa-412a-94e6-4b68a1fcac4f",
				Type:       "transfer",
				ToWalletID: sql.NullString{String: "e0925d9a-03ce-4639-b78b-1a52fec87ac4", Valid: true},
				Amount:     40,
				CreatedAt:  time.Now(),
			},
		}

		databaseMock := mocks2.NewMockDatabase(ctrl)
		walletService := walletServiceV1{DB: databaseMock}
		databaseMock.EXPECT().CheckWalletOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		databaseMock.EXPECT().GetTransactions(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(transactions, nil)

		resp := walletService.GetTransactionHistory(context.Background(), walletID, userID, 10, 0)

		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.Status)
		assert.Equal(t, "transaction history retrieval successful", resp.Message)
		if trsResp, ok := resp.Data.(responses.TransactionHistoryResponse); ok {
			assert.Equal(t, walletID, trsResp.WalletID)
			assert.Equal(t, len(transactions), len(trsResp.Transactions))
			for i, tr := range transactions {
				assert.Equal(t, tr.ToWalletID.String, trsResp.Transactions[i].ToWalletID)
				assert.Equal(t, tr.Type, trsResp.Transactions[i].Type)
				assert.Equal(t, tr.CreatedAt.UnixNano(), trsResp.Transactions[i].CreatedAt.UnixNano())
			}
		} else {
			assert.Fail(t, "cannot cast into response")
		}
	})

}
