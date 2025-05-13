package api

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

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
		if err := validateAmount(r.Amount); err != nil {
			return err
		}
		if err := validateIdempotencyToken(r.IdempotencyToken); err != nil {
			return err
		}

	case *requests.WithdrawRequest:
		if err := validateAmount(r.Amount); err != nil {
			return err
		}
		if err := validateIdempotencyToken(r.IdempotencyToken); err != nil {
			return err
		}

	case *requests.TransferRequest:
		if err := validateAmount(r.Amount); err != nil {
			return err
		}
		if err := validateIdempotencyToken(r.IdempotencyToken); err != nil {
			return err
		}
		if err := validateReceiverAccountInfo(r.ToAccount); err != nil {
			return err
		}
	}
	return nil
}

func validateAmount(amount float64) error {
	if amount <= 0 {
		return errors.New("invalid amount")
	}
	if amount > 50000 {
		return errors.New("max amount to transfer is 50,000 in a single request")
	}
	return nil
}

func validateReceiverAccountInfo(walletID string) error {
	if walletID == "" {
		return errors.New("receiver account info is empty")
	}
	if err := validateUUID(walletID); err != nil {
		return errors.New("receiver account info is invalid. wallet id must be a valid UUID")
	}
	return nil
}

func validateIdempotencyToken(token string) error {
	if token == "" {
		return errors.New("missing idempotency token")
	}
	if len(token) < 8 || len(token) > 20 {
		return errors.New("idempotency token must be between 8 and 20 characters")
	}
	return nil
}

func validateUUID(walletID string) error {
	if _, err := uuid.Parse(walletID); err != nil {
		return errors.New("invalid wallet ID: must be a valid UUID")
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
