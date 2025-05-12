package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitDB() *sqlx.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	DB, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("error opening db: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}
	maxConnLifeTime, err := strconv.ParseInt(os.Getenv("DB_MAX_CONN_LIFETIME_SEC"), 10, 32)
	if err != nil {
		log.Warnf("cannot parse %s of DB_MAX_CONN_LIFETIME_SEC. using default 30s", os.Getenv("DB_MAX_CONN_LIFETIME_SEC"))
		maxConnLifeTime = 30
	}
	DB.SetConnMaxLifetime(time.Duration(maxConnLifeTime) * time.Second)
	maxOpenConnections, err := strconv.ParseInt(os.Getenv("DB_MAX_OPEN_CONNECTIONS"), 10, 32)
	if err != nil {
		log.Warnf("cannot parse %s of DB_MAX_CONN_LIFETIME_SEC. using default 10", os.Getenv("DB_MAX_OPEN_CONNECTIONS"))
		maxConnLifeTime = 10
	}
	DB.SetMaxOpenConns(int(maxOpenConnections))

	fmt.Println("connected to postgreSQL")
	return DB
}
