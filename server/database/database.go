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
		log.Println("warning: Could not load .env file. Make sure it exists.")
	}

	// Get the POSTGRES_URL environment variable
	dbURL = os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		return fmt.Errorf("POSTGRES_URL is not set in the environment")
	}
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

func DeleteUser(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := "DELETE FROM users WHERE api_key = $1;"
	_, err = conn.Exec(context.Background(), query, apiKey)
	return err
}

func DeleteRequests(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := "DELETE FROM requests WHERE api_key = $1;"
	_, err = conn.Exec(context.Background(), query, apiKey)
	return err
}

func DeleteMonitors(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := "DELETE FROM monitor WHERE api_key = $1;"
	_, err = conn.Exec(context.Background(), query, apiKey)
	return err
}

func DeletePings(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := "DELETE FROM pings WHERE api_key = $1;"
	_, err = conn.Exec(context.Background(), query, apiKey)
	return err
}
