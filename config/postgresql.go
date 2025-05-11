package config

import (
	"fmt"
	"log"
	"os"

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

	fmt.Println("connected to postgreSQL")
	return DB
}
