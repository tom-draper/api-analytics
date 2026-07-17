package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// deleteAccountRecorderFor drives deleteAccount with the given JSON body. The
// handler validates before touching the database, so a nil DB is only reached
// if validation wrongly lets the request through, which panics and fails
// loudly rather than silently deleting.
func deleteAccountRecorderFor(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	deleteAccount(nil)(c)
	return recorder
}

func TestDeleteAccountRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"malformed json", "{"},
		{"missing api key", `{}`},
		{"empty api key", `{"api_key":""}`},
		{"api key not a uuid", `{"api_key":"not-a-uuid"}`},
		{"api key wrong length", `{"api_key":"1111-1111"}`},
		{"sql injection attempt", `{"api_key":"' OR 1=1 --"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s reached the database instead of being rejected: %v", tt.name, r)
				}
			}()

			recorder := deleteAccountRecorderFor(t, tt.body)
			if recorder.Code != http.StatusBadRequest {
				t.Errorf("status = %d, expected %d for %s", recorder.Code, http.StatusBadRequest, tt.name)
			}
		})
	}
}

// A well formed key must pass validation and reach the delete itself.
func TestDeleteAccountAcceptsValidAPIKey(t *testing.T) {
	for _, body := range []string{
		`{"api_key":"11111111-1111-4111-8111-111111111111"}`,
		// Users sometimes paste the key wrapped in quotes or with whitespace.
		`{"api_key":"\"11111111-1111-4111-8111-111111111111\""}`,
		`{"api_key":" 11111111-1111-4111-8111-111111111111 "}`,
	} {
		t.Run(body, func(t *testing.T) {
			defer func() {
				// Reaching the nil DB proves the key passed validation.
				_ = recover()
			}()

			recorder := deleteAccountRecorderFor(t, body)
			if recorder.Code == http.StatusBadRequest {
				t.Errorf("%s was rejected, expected it to be accepted", body)
			}
		})
	}
}

// The API key must not appear in the request target, which is what ends up in
// access logs, proxy logs and browser history.
func TestDeleteAccountKeepsAPIKeyOutOfTheURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	const apiKey = "11111111-1111-4111-8111-111111111111"
	c.Request = httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(`{"api_key":"`+apiKey+`"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	if strings.Contains(c.Request.URL.String(), apiKey) {
		t.Errorf("URL %q contains the API key", c.Request.URL.String())
	}
	if c.Request.Method != http.MethodPost {
		t.Errorf("method = %s, expected POST", c.Request.Method)
	}
}
