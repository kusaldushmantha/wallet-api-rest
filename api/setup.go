package api

import (
	"github.com/gofiber/fiber/v2"

	"WalletApp/config"
	"WalletApp/db"
	"WalletApp/services"
)

func Setup(app *fiber.App) {
	pgClient := config.InitDB()
	redisClient := config.InitRedis()

	database := db.NewPostgreSQLDB(pgClient)
	cache := db.NewRedisCache(redisClient)
	walletService := services.NewWalletServiceV1(database, cache)
	userService := services.NewUserService(database)

	setupAPIGroups(app, walletService, userService)
}
