package email

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestSendEmail(t *testing.T) {
	address := GetEmailAddress()
	err := SendEmail("Test Subject", "Test Body", address)
	if err != nil {
		t.Error(err)
	}
}
