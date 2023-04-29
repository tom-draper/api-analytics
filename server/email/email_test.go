package email

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func getTestEmailAddress() string {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	address := os.Getenv("TEST_EMAIL_ADDRESS")
	return address
}

func TestSendEmail(t *testing.T) {
	address := getTestEmailAddress()
	err := SendEmail("Test Subject", "Test Body", address)
	if err != nil {
		t.Error(err)
	}
}
