package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func getDBLogin() (string, string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	database := os.Getenv("POSTGRES_DATABASE")
	username := os.Getenv("POSTGRES_USERNAME")
	password := os.Getenv("POSTGRES_PASSWORD")
	return database, username, password
}

func OpenDBConnection() *sql.DB {
	database, username, password := getDBLogin()
	args := fmt.Sprintf("host=%s port=%d dbname=%s user='%s' password=%s sslmode=%s", "localhost", 5432, database, username, password, "disable")

	db, err := sql.Open("postgres", args)
	if err != nil {
		panic(err)
	}

	return db
}
