package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"

	"WalletApp/api"
)

func main() {
	app := fiber.New()
	env := os.Getenv("GO_ENV")
	envFile := ".env.local" // default
	if env == "docker" {
		envFile = ".env.docker"
	}
	err := godotenv.Load(envFile)
	if err != nil {
		log.Warnf("no env file loaded from %s err: %v", envFile, err)
	}
	api.Setup(app)
	err = app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
