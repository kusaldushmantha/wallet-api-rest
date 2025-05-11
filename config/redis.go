package config

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		DB:   db,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("error connecting to Redis: %v", err)
	}

	fmt.Println("connected to Redis")
	return redisClient
}
