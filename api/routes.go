package api

import (
	"WalletApp/services"
	"github.com/gofiber/fiber/v2"

	"WalletApp/models/requests"
)

// API grouping
func setupAPIGroups(app *fiber.App, walletService services.WalletService, userService services.UserService) {
	walletEndpoints := app.Group("/wallet/v1")
	walletEndpoints.Post("/:id/deposit", depositHandler(walletService))
	walletEndpoints.Post("/:id/withdraw", withdrawHandler(walletService))
	walletEndpoints.Post("/:id/transfer", transferHandler(walletService))
	walletEndpoints.Get("/:id/balance", getBalanceHandler(walletService))
	walletEndpoints.Get("/:id/transactions", getTransactionsHandler(walletService))

	// This is a helper API to create users and wallets if needed.
	userManagementEndpoints := app.Group("/user-management/v1")
	userManagementEndpoints.Post("/", createUserAndWalletHandler(userService))
	userManagementEndpoints.Get("/", getWalletUsersHandler(userService))
}

func depositHandler(walletService services.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		userID, err := getUserIDHeaderValue(c)
		if err != nil {
			return badRequest(c, err.Error())
		}
		walletID := c.Params("id")
		var req requests.DepositRequest
		if err := payloadValidation(c, &req); err != nil {
			return badRequest(c, err.Error())
		}

		resp := walletService.Deposit(ctx, req.IdempotencyToken, walletID, req.Amount, userID)
		return response(c, resp)
	}
}

func withdrawHandler(walletService services.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		userID, err := getUserIDHeaderValue(c)
		if err != nil {
			return badRequest(c, err.Error())
		}

		walletID := c.Params("id")
		var req requests.WithdrawRequest
		if err := payloadValidation(c, &req); err != nil {
			return badRequest(c, err.Error())
		}

		resp := walletService.Withdraw(ctx, req.IdempotencyToken, walletID, req.Amount, userID)
		return response(c, resp)
	}
}

func transferHandler(walletService services.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		userID, err := getUserIDHeaderValue(c)
		if err != nil {
			return badRequest(c, err.Error())
		}

		walletID := c.Params("id")
		var req requests.TransferRequest
		if err := payloadValidation(c, &req); err != nil {
			return badRequest(c, err.Error())
		}

		resp := walletService.Transfer(ctx, req.IdempotencyToken, walletID, req.ToAccount, req.Amount, userID)
		return response(c, resp)
	}
}

func getBalanceHandler(walletService services.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		userID, err := getUserIDHeaderValue(c)
		if err != nil {
			return badRequest(c, err.Error())
		}

		walletID := c.Params("id")
		resp := walletService.GetBalance(ctx, walletID, userID)
		return response(c, resp)
	}
}

func getTransactionsHandler(walletService services.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		userID, err := getUserIDHeaderValue(c)
		if err != nil {
			return badRequest(c, err.Error())
		}

		walletID := c.Params("id")
		// Use query params for pagination
		limit := parseQueryInt(c.Query("limit"), 10)
		offset := parseQueryInt(c.Query("offset"), 0)

		resp := walletService.GetTransactionHistory(ctx, walletID, userID, limit, offset)
		return response(c, resp)
	}
}

func getWalletUsersHandler(userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		resp := userService.GetUsers(ctx)
		return response(c, resp)
	}
}

func createUserAndWalletHandler(userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		resp := userService.CreateUser(ctx)
		return response(c, resp)
	}
}
