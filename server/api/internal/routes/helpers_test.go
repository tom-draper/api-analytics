package routes

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCompressJSONRoundTrip(t *testing.T) {
	input := map[string]any{"hello": "world", "n": float64(42)}

	compressed, err := compressJSON(input)
	if err != nil {
		t.Fatalf("compressJSON: %v", err)
	}

	gzr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("output is not valid gzip: %v", err)
	}
	raw, err := io.ReadAll(gzr)
	if err != nil {
		t.Fatalf("gunzip: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decompressed body is not valid JSON: %v", err)
	}
	if out["hello"] != "world" || out["n"] != float64(42) {
		t.Errorf("round trip mismatch: %v", out)
	}
}

func headerContext(headers map[string]string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	for k, v := range headers {
		c.Request.Header.Set(k, v)
	}
	return c
}

func TestGetAPIKeyFromHeader(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{"x-auth-token", map[string]string{"X-AUTH-TOKEN": "key-1"}, "key-1"},
		{"legacy api-key fallback", map[string]string{"API-Key": "key-2"}, "key-2"},
		{
			"x-auth-token wins over legacy",
			map[string]string{"X-AUTH-TOKEN": "primary", "API-Key": "legacy"},
			"primary",
		},
		{"quotes stripped", map[string]string{"X-AUTH-TOKEN": `"key-3"`}, "key-3"},
		{"surrounding space trimmed", map[string]string{"X-AUTH-TOKEN": "  key-4  "}, "key-4"},
		{"none set", map[string]string{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAPIKeyFromHeader(headerContext(tt.headers)); got != tt.want {
				t.Errorf("getAPIKeyFromHeader = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetNullableString(t *testing.T) {
	if got := getNullableString(nil); got != "" {
		t.Errorf("nil pointer = %q, want empty", got)
	}
	value := "present"
	if got := getNullableString(&value); got != "present" {
		t.Errorf("pointer = %q, want %q", got, value)
	}
}
