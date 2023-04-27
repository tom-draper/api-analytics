package email

import "testing"

func TestSendEmail(t *testing.T){
    err := SendEmail("Test Subject", "Test Body", "")
    if err != nil {
		t.Error(err)
	}
}
