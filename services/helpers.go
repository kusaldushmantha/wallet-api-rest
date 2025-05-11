package services

import (
	"net/http"

	"WalletApp/models/responses"
)

func successBalanceResponse(walletID string, balance float64, msg string) responses.Response {
	return getResponse(responses.WalletBalanceResponse{WalletID: walletID, Balance: balance}, msg, http.StatusOK, nil)
}

func internalError(msg string, err error) responses.Response {
	return getResponse(nil, msg, http.StatusInternalServerError, err)
}

func unauthorizedResponse() responses.Response {
	return getResponse(nil, "unauthorized access to the wallet", http.StatusUnauthorized, nil)
}

func duplicateRequestResponse() responses.Response {
	return getResponse(nil, "duplicate request: idempotency key already exists", http.StatusConflict, nil)
}

func badRequest(msg string) responses.Response {
	return getResponse(nil, msg, http.StatusBadRequest, nil)
}

func getResponse(data any, message string, status int, err error) responses.Response {
	return responses.Response{
		Status:  status,
		Message: message,
		Data:    data,
		Error:   err,
	}
}
