package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func getDatabaseURL() string {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	url := os.Getenv("POSTGRES_URL")
	return url
}

func NewConnection() *pgx.Conn {
	url := getDatabaseURL()
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		panic(err)
	}
	return conn
}

func DeleteUser(apiKey string) error {
	conn := NewConnection()
	defer conn.Close(context.Background())

	query := "DELETE FROM users WHERE api_key = $1;"
	_, err := conn.Exec(context.Background(), query, apiKey)
	return err
}

func DeleteRequests(apiKey string) error {
	conn := NewConnection()
	defer conn.Close(context.Background())

	query := "DELETE FROM requests WHERE api_key = $1;"
	_, err := conn.Exec(context.Background(), query, apiKey)
	return err
}

func DeleteMonitors(apiKey string) error {
	conn := NewConnection()
	defer conn.Close(context.Background())

	query := "DELETE FROM monitor WHERE api_key = $1;"
	_, err := conn.Exec(context.Background(), query, apiKey)
	return err
}

func DeletePings(apiKey string) error {
	conn := NewConnection()
	defer conn.Close(context.Background())

	query := "DELETE FROM pings WHERE api_key = $1;"
	_, err := conn.Exec(context.Background(), query, apiKey)
	return err
}
