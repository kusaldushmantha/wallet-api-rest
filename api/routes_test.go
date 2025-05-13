package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"WalletApp/models/responses"
	"WalletApp/services/mocks"
)

func TestDepositHandler(t *testing.T) {
	t.Run("should return bad request if the user-id header is not present", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/deposit", strings.NewReader(`{"amount": 100}`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return bad request if the amount is not positive", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/deposit", strings.NewReader(`{"amount": 0}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return bad request if the idempotency token is empty", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/deposit", strings.NewReader(`{"amount": 100}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return success response if the service provides success response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockWalletService := mocks.NewMockWalletService(ctrl)
		message := "success"
		userID := "12345"
		walletID := "23456"
		createdAt := time.Now()

		mockWalletService.EXPECT().Deposit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(responses.Response{
			Status:  http.StatusOK,
			Message: message,
			Data: responses.UserWalletResponse{UserWallets: []*responses.UserWallet{{
				UserID:    userID,
				WalletID:  walletID,
				CreatedAt: createdAt,
			}}},
		})
		app := fiber.New()
		setupAPIGroups(app, mockWalletService, nil)

		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/deposit",
			strings.NewReader(`{"amount": 100, "idempotency_token": "abcd-efgh-ijkl"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("error while initiating test api call", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error while reading body", err)
		}

		var bodyMap map[string]interface{}
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			t.Fatal("error while unmarshalling body into a map", err)
		}
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, message, bodyMap["message"])

		dataMap, ok := bodyMap["data"].(map[string]interface{})
		if !ok {
			t.Fatal("data field is not a map")
		}

		dataBytes, err := json.Marshal(dataMap)
		if err != nil {
			t.Fatal(err)
		}

		var userWalletResponse responses.UserWalletResponse
		err = json.Unmarshal(dataBytes, &userWalletResponse)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, userID, userWalletResponse.UserWallets[0].UserID)
		assert.Equal(t, walletID, userWalletResponse.UserWallets[0].WalletID)
		assert.Equal(t, createdAt.UnixNano(), userWalletResponse.UserWallets[0].CreatedAt.UnixNano())
	})

	t.Run("should return an error response if the service provides an error response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockWalletService := mocks.NewMockWalletService(ctrl)
		message := "invalid request"
		expectedError := errors.New("test error")

		mockWalletService.EXPECT().Deposit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(responses.Response{
			Status:  http.StatusBadRequest,
			Message: message,
			Error:   expectedError,
		})
		app := fiber.New()
		setupAPIGroups(app, mockWalletService, nil)

		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/deposit",
			strings.NewReader(`{"amount": 100, "idempotency_token": "abcd-efgh-ijkl"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("error while initiating test api call", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error while reading body", err)
		}

		var bodyMap map[string]interface{}
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			t.Fatal("error while unmarshalling body into a map", err)
		}
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, message, bodyMap["message"])
		assert.Equal(t, expectedError.Error(), bodyMap["error"])
	})
}

func TestWithdrawHandler(t *testing.T) {
	t.Run("should return bad request if the user-id header is not present", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/withdraw", strings.NewReader(`{"amount": 100}`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return bad request if the amount is not positive", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/withdraw", strings.NewReader(`{"amount": 0}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return bad request if the idempotency token is empty", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/withdraw", strings.NewReader(`{"amount": 100}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return success response if the service provides success response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockWalletService := mocks.NewMockWalletService(ctrl)
		message := "success"

		mockWalletService.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(responses.Response{
			Status:  http.StatusOK,
			Message: message,
		})
		app := fiber.New()
		setupAPIGroups(app, mockWalletService, nil)

		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/withdraw",
			strings.NewReader(`{"amount": 100, "idempotency_token": "abcd-efgh-ijkl"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("error while initiating test api call", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error while reading body", err)
		}

		var bodyMap map[string]interface{}
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			t.Fatal("error while unmarshalling body into a map", err)
		}
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, message, bodyMap["message"])
	})

	t.Run("should return an error response if the service provides an error response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockWalletService := mocks.NewMockWalletService(ctrl)
		message := "invalid request"
		expectedError := errors.New("test error")

		mockWalletService.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(responses.Response{
			Status:  http.StatusBadRequest,
			Message: message,
			Error:   expectedError,
		})
		app := fiber.New()
		setupAPIGroups(app, mockWalletService, nil)

		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/withdraw",
			strings.NewReader(`{"amount": 100, "idempotency_token": "abcd-efgh-ijkl"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("error while initiating test api call", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error while reading body", err)
		}

		var bodyMap map[string]interface{}
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			t.Fatal("error while unmarshalling body into a map", err)
		}
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, message, bodyMap["message"])
	})
}

func TestTransferHandler(t *testing.T) {
	t.Run("should return bad request if the user-id header is not present", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/transfer",
			strings.NewReader(`{
				  "amount":100,
				  "recipient_wallet_id": "2cbcd158-56d2-4d45-8113-d51adf9ef57a",
				  "idempotency_token": "gggg"
				}`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return bad request if the amount is not positive", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/transfer",
			strings.NewReader(`{
				  "amount":0,
				  "recipient_wallet_id": "2cbcd158-56d2-4d45-8113-d51adf9ef57a",
				  "idempotency_token": "gggg"
				}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return bad request if the idempotency token is empty", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/transfer",
			strings.NewReader(`{
				  "amount":100,
				  "recipient_wallet_id": "2cbcd158-56d2-4d45-8113-d51adf9ef57a",
				}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return bad request if the recipient wallet id is empty", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/transfer",
			strings.NewReader(`{
				  "amount":100,
				  "idempotency_token": "gggg"
				}`))
		req.Header.Set("Content-Type", "application/json")

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return success response if the service provides success response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockWalletService := mocks.NewMockWalletService(ctrl)
		message := "success"

		mockWalletService.EXPECT().Transfer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(responses.Response{
			Status:  http.StatusOK,
			Message: message,
		})
		app := fiber.New()
		setupAPIGroups(app, mockWalletService, nil)

		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/transfer",
			strings.NewReader(`{
				  "amount":100,
				  "recipient_wallet_id": "2cbcd158-56d2-4d45-8113-d51adf9ef57a",
				  "idempotency_token": "abcd-efgh-ijkl"
				}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("error while initiating test api call", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error while reading body", err)
		}

		var bodyMap map[string]interface{}
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			t.Fatal("error while unmarshalling body into a map", err)
		}
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, message, bodyMap["message"])
	})

	t.Run("should return an error response if the service provides an error response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockWalletService := mocks.NewMockWalletService(ctrl)
		message := "invalid request"
		expectedError := errors.New("test error")

		mockWalletService.EXPECT().Transfer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(responses.Response{
			Status:  http.StatusBadRequest,
			Message: message,
			Error:   expectedError,
		})
		app := fiber.New()
		setupAPIGroups(app, mockWalletService, nil)

		req := httptest.NewRequest("POST", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/transfer",
			strings.NewReader(`{
				  "amount":100,
				  "recipient_wallet_id": "2cbcd158-56d2-4d45-8113-d51adf9ef57a",
				  "idempotency_token": "abcd-efgh-ijkl"
				}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("error while initiating test api call", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error while reading body", err)
		}

		var bodyMap map[string]interface{}
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			t.Fatal("error while unmarshalling body into a map", err)
		}
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, message, bodyMap["message"])
	})
}

func TestGetTransactionsHandler(t *testing.T) {
	t.Run("should return bad request if the user-id header is not present", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("GET", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/transactions", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return success response if the service provides success response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockWalletService := mocks.NewMockWalletService(ctrl)
		message := "success"

		mockWalletService.EXPECT().GetTransactionHistory(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(responses.Response{
			Status:  http.StatusOK,
			Message: message,
		})
		app := fiber.New()
		setupAPIGroups(app, mockWalletService, nil)

		req := httptest.NewRequest("GET", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/transactions", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("error while initiating test api call", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error while reading body", err)
		}

		var bodyMap map[string]interface{}
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			t.Fatal("error while unmarshalling body into a map", err)
		}
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, message, bodyMap["message"])
	})

	t.Run("should return an error response if the service provides an error response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockWalletService := mocks.NewMockWalletService(ctrl)
		message := "invalid request"
		expectedError := errors.New("test error")

		mockWalletService.EXPECT().GetTransactionHistory(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(responses.Response{
			Status:  http.StatusBadRequest,
			Message: message,
			Error:   expectedError,
		})
		app := fiber.New()
		setupAPIGroups(app, mockWalletService, nil)

		req := httptest.NewRequest("GET", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/transactions", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("error while initiating test api call", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error while reading body", err)
		}

		var bodyMap map[string]interface{}
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			t.Fatal("error while unmarshalling body into a map", err)
		}
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, message, bodyMap["message"])
	})
}

func TestWalletBalance(t *testing.T) {
	t.Run("should return bad request if the user-id header is not present", func(t *testing.T) {
		app := fiber.New()
		setupAPIGroups(app, nil, nil)
		req := httptest.NewRequest("GET", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/balance", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return success response if the service provides success response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockWalletService := mocks.NewMockWalletService(ctrl)
		message := "success"

		mockWalletService.EXPECT().GetBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(responses.Response{
			Status:  http.StatusOK,
			Message: message,
		})
		app := fiber.New()
		setupAPIGroups(app, mockWalletService, nil)

		req := httptest.NewRequest("GET", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/balance", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("error while initiating test api call", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error while reading body", err)
		}

		var bodyMap map[string]interface{}
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			t.Fatal("error while unmarshalling body into a map", err)
		}
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, message, bodyMap["message"])
	})

	t.Run("should return an error response if the service provides an error response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockWalletService := mocks.NewMockWalletService(ctrl)
		message := "invalid request"
		expectedError := errors.New("test error")

		mockWalletService.EXPECT().GetBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(responses.Response{
			Status:  http.StatusBadRequest,
			Message: message,
			Error:   expectedError,
		})
		app := fiber.New()
		setupAPIGroups(app, mockWalletService, nil)

		req := httptest.NewRequest("GET", "/wallet/v1/2ad7eec6-51f3-409f-9e82-582a68417f6f/balance", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-123")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("error while initiating test api call", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("error while reading body", err)
		}

		var bodyMap map[string]interface{}
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			t.Fatal("error while unmarshalling body into a map", err)
		}
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, message, bodyMap["message"])
	})
}
