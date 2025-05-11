package api

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"WalletApp/models/requests"
	"WalletApp/models/responses"
)

func getUserIDHeaderValue(c *fiber.Ctx) (string, error) {
	userID := c.Get("X-User-ID")
	if userID == "" {
		return "", errors.New("mandatory header X-User-ID not provided")
	}
	return userID, nil
}

func payloadValidation[T any](c *fiber.Ctx, req *T) error {
	if err := c.BodyParser(req); err != nil {
		return errors.New("invalid request payload")
	}

	switch r := any(req).(type) {
	case *requests.DepositRequest:
		if r.Amount <= 0 {
			return errors.New("invalid amount")
		}
		if r.IdempotencyToken == "" {
			return errors.New("missing idempotency token")
		}
	case *requests.WithdrawRequest:
		if r.Amount <= 0 {
			return errors.New("invalid amount")
		}
		if r.IdempotencyToken == "" {
			return errors.New("missing idempotency token")
		}
	case *requests.TransferRequest:
		if r.Amount <= 0 {
			return errors.New("invalid amount")
		}
		if r.IdempotencyToken == "" {
			return errors.New("missing idempotency token")
		}
		if r.ToAccount == "" {
			return errors.New("missing recipient account")
		}
	}
	return nil
}

func parseQueryInt(value string, defaultVal int32) int32 {
	if i, err := strconv.Atoi(value); err == nil && i >= 0 {
		return int32(i)
	}
	return defaultVal
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": msg,
	})
}

func response(c *fiber.Ctx, resp responses.Response) error {
	fibreMap := fiber.Map{
		"message": resp.Message,
		"data":    resp.Data,
	}
	if resp.Error != nil {
		fibreMap["error"] = resp.Error.Error()
	}
	return c.Status(resp.Status).JSON(fibreMap)
}
