package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var dbURL string

func LoadConfig() error {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("warning: could not load .env file")
	}

	// Get the POSTGRES_URL environment variable
	dbURL = os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		return fmt.Errorf("POSTGRES_URL is not set in the environment")
	}

	log.Printf("POSTGRES_URL=%s\n", dbURL)

	return nil
}

func NewConnection() (*pgx.Conn, error) {
	if dbURL == "" {
		err := LoadConfig()
		if err != nil {
			return nil, err
		}
	}

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
