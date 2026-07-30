package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTryNewUserParsesQuotedUUID(t *testing.T) {
	uuid := "12345678-1234-1234-1234-123456789012" // 36 chars
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(uuid) // writes a JSON string: "uuid"
	}))
	defer srv.Close()

	c := NewClient(srv.URL+"/", "", "", nil)
	if err := c.TryNewUser(); err != nil {
		t.Fatalf("TryNewUser with a valid quoted UUID: %v", err)
	}
}

func TestTryNewUserRejectsMalformedBodyWithoutPanic(t *testing.T) {
	// A body shorter than two bytes must yield a decode error, not a panic (the
	// old sb[1:len(sb)-1] quote-strip would slice out of range).
	for _, body := range []string{"", "\"", "x"} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(body))
		}))

		c := NewClient(srv.URL+"/", "", "", nil)
		if err := c.TryNewUser(); err == nil {
			t.Errorf("TryNewUser with body %q = nil error, want a decode error", body)
		}
		srv.Close()
	}
}
