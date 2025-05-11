package services

import (
	"context"
	"net/http"

	"WalletApp/db"
	"WalletApp/models/responses"
)

type userServiceV1 struct {
	DB db.Database
}

func NewUserService(dbClient db.Database) UserService {
	return &userServiceV1{
		DB: dbClient,
	}
}

// CreateUser creates a new user and an associated wallet in the system.
//
// This method delegates the creation of the user and wallet to the database layer,
// and returns the user ID and wallet ID if successful. It wraps the result in a
// standard API response format.
//
// Parameters:
//   - ctx: Context for managing request lifecycle, including timeout and cancellation.
//
// Returns:
//   - A Response struct containing the created user's wallet information (user ID, wallet ID,
//     and creation timestamp), a success message, HTTP status code 201, or an internal
//     error response if the operation fails.
func (u userServiceV1) CreateUser(ctx context.Context) responses.Response {
	wallet, err := u.DB.CreateUserWallet(ctx)
	if err != nil || wallet == nil {
		return internalError("failed to create user and wallet", err)
	}
	return getResponse(responses.UserWalletResponse{
		UserWallets: []*responses.UserWallet{
			{
				UserID:    wallet.UserID,
				WalletID:  wallet.ID,
				CreatedAt: wallet.CreatedAt,
			},
		},
	}, "user and wallet created successfully", http.StatusCreated, nil)
}

// GetUsers retrieves all users and their associated wallets from the system.
//
// This method fetches a list of users and their wallet details from the database.
// It returns the data in a structured API response format, including the user ID,
// wallet ID, and creation timestamp. If the operation fails, an internal error
// response is returned.
//
// Parameters:
//   - ctx: Context for managing the request lifecycle, including timeout and cancellation.
//
// Returns:
//   - A Response struct containing the list of user wallets with user ID, wallet ID,
//     and creation timestamp, a success message, and HTTP status code 200.
//     If the operation fails, an internal error response is returned.
func (u userServiceV1) GetUsers(ctx context.Context) responses.Response {
	wallet, err := u.DB.GetWalletUsers(ctx)
	if err != nil || wallet == nil {
		return internalError("failed to get user and wallet", err)
	}

	var userWallets []*responses.UserWallet
	for _, res := range wallet {
		userWallets = append(userWallets, &responses.UserWallet{
			UserID:    res.UserID,
			WalletID:  res.ID,
			CreatedAt: res.CreatedAt,
		})
	}

	return getResponse(responses.UserWalletResponse{
		UserWallets: userWallets,
	}, "user and wallet retrieved successfully", http.StatusOK, nil)
}
