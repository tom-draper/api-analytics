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
