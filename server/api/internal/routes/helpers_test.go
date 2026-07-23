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

func TestStreamGzipJSONRoundTrip(t *testing.T) {
	input := map[string]any{"hello": "world", "n": float64(42)}

	var buf bytes.Buffer
	if err := streamGzipJSON(&buf, input); err != nil {
		t.Fatalf("streamGzipJSON: %v", err)
	}

	gzr, err := gzip.NewReader(&buf)
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

// TestDashboardDataJSONShape locks the field names the dashboard reads from a
// paginated response. Renaming a tag here would silently break pagination.
func TestDashboardDataJSONShape(t *testing.T) {
	raw, err := json.Marshal(DashboardData{
		UserAgents: UserAgentsLookup{1: "Mozilla/5.0"},
		Requests:   [][11]any{{"1.2.3.4", "/path"}},
		HasMore:    true,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out map[string]json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"user_agents", "requests", "has_more"} {
		if _, ok := out[key]; !ok {
			t.Errorf("response is missing the %q field: %s", key, raw)
		}
	}
	if string(out["has_more"]) != "true" {
		t.Errorf("has_more = %s, want true", out["has_more"])
	}
}

// TestRequestDataJSONShape locks the public /data row field names.
func TestRequestDataJSONShape(t *testing.T) {
	raw, err := json.Marshal(RequestData{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{
		"ip_address", "path", "hostname", "user_agent", "method",
		"status", "response_time", "location", "referrer", "user_id", "created_at",
	} {
		if _, ok := out[key]; !ok {
			t.Errorf("RequestData JSON is missing the %q field: %s", key, raw)
		}
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
