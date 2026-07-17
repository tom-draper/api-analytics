package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func queriesFor(t *testing.T, rawQuery string) DataFetchQueries {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/data?"+rawQuery, nil)
	return getQueriesFromRequest(c)
}

func TestGetQueriesFromRequestUserIDSpellings(t *testing.T) {
	// The documented name is userId; userID was the only spelling the handler
	// used to accept. An unmatched name silently drops the filter and returns
	// unfiltered data, so every reasonable spelling must resolve.
	for _, rawQuery := range []string{
		"userId=abc",
		"userID=abc",
		"user_id=abc",
		"userid=abc",
		"USERID=abc",
		"User_Id=abc",
	} {
		t.Run(rawQuery, func(t *testing.T) {
			if got := queriesFor(t, rawQuery).userID; got != "abc" {
				t.Errorf("userID = %q, expected %q", got, "abc")
			}
		})
	}
}

func TestGetQueriesFromRequestIPAddressSpellings(t *testing.T) {
	for _, rawQuery := range []string{
		"ipAddress=1.2.3.4",
		"ip_address=1.2.3.4",
		"ipaddress=1.2.3.4",
		"IPAddress=1.2.3.4",
		"ip=1.2.3.4",
	} {
		t.Run(rawQuery, func(t *testing.T) {
			if got := queriesFor(t, rawQuery).ipAddress; got != "1.2.3.4" {
				t.Errorf("ipAddress = %q, expected %q", got, "1.2.3.4")
			}
		})
	}
}

func TestGetQueriesFromRequestDateRangeSpellings(t *testing.T) {
	for _, rawQuery := range []string{
		"dateFrom=2022-01-01",
		"date_from=2022-01-01",
		"datefrom=2022-01-01",
	} {
		t.Run(rawQuery, func(t *testing.T) {
			got := queriesFor(t, rawQuery).dateFrom
			if got.IsZero() || got.Format("2006-01-02") != "2022-01-01" {
				t.Errorf("dateFrom = %v, expected 2022-01-01", got)
			}
		})
	}
}

func TestGetQueriesFromRequestDocumentedExample(t *testing.T) {
	// The example from the README's data API section.
	queries := queriesFor(t, "page=3&dateFrom=2022-01-01&hostname=apianalytics.dev&status=200&userId=b56cbd92-1168-4d7b-8d94-0418da207908")

	if queries.page != 3 {
		t.Errorf("page = %d, expected 3", queries.page)
	}
	if queries.hostname != "apianalytics.dev" {
		t.Errorf("hostname = %q, expected %q", queries.hostname, "apianalytics.dev")
	}
	if queries.status != 200 {
		t.Errorf("status = %d, expected 200", queries.status)
	}
	if queries.userID != "b56cbd92-1168-4d7b-8d94-0418da207908" {
		t.Errorf("userID = %q, expected the documented user ID", queries.userID)
	}
	if queries.dateFrom.Format("2006-01-02") != "2022-01-01" {
		t.Errorf("dateFrom = %v, expected 2022-01-01", queries.dateFrom)
	}
}

func TestGetQueriesFromRequestUnsetFiltersAreEmpty(t *testing.T) {
	queries := queriesFor(t, "")

	if queries.page != 1 {
		t.Errorf("page = %d, expected the default of 1", queries.page)
	}
	if queries.userID != "" || queries.ipAddress != "" || queries.hostname != "" || queries.location != "" {
		t.Errorf("expected no filters to be set, got %+v", queries)
	}
	if queries.status != 0 {
		t.Errorf("status = %d, expected 0", queries.status)
	}
	if queries.compact {
		t.Error("compact = true, expected false")
	}
}

func TestGetQueriesFromRequestConflictingSpellingsAreDeterministic(t *testing.T) {
	// Two spellings of one filter must always resolve to the same value rather
	// than depending on map iteration order.
	first := queriesFor(t, "user_id=aaa&userId=zzz").userID
	for range 50 {
		if got := queriesFor(t, "user_id=aaa&userId=zzz").userID; got != first {
			t.Fatalf("userID resolved to %q then %q; expected a stable result", first, got)
		}
	}
}

func TestNormalizeQueryKey(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"userId", "userid"},
		{"userID", "userid"},
		{"user_id", "userid"},
		{"USER_ID", "userid"},
		{"user-id", "userid"},
		{"ipAddress", "ipaddress"},
		{"ip_address", "ipaddress"},
		{"dateFrom", "datefrom"},
		{"page", "page"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeQueryKey(tt.name); got != tt.expected {
				t.Errorf("normalizeQueryKey(%q) = %q, expected %q", tt.name, got, tt.expected)
			}
		})
	}
}
