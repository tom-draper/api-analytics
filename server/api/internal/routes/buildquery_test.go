package routes

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

const testKey = "b56cbd92-1168-4d7b-8d94-0418da207908"

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("bad test date %q: %v", s, err)
	}
	return d
}

func TestBuildDataFetchQueryBase(t *testing.T) {
	query, args := buildDataFetchQuery(testKey, DataFetchQueries{page: 1})

	if !strings.Contains(query, "WHERE api_key = $1") {
		t.Errorf("query missing api_key clause:\n%s", query)
	}
	if !strings.HasSuffix(query, "ORDER BY created_at LIMIT $2 OFFSET $3;") {
		t.Errorf("unexpected tail:\n%s", query)
	}
	// No filters: just api_key, page size and offset.
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[0] != testKey {
		t.Errorf("args[0] = %v, want the api key", args[0])
	}
	if args[1] != 50_000 {
		t.Errorf("page size arg = %v, want 50000", args[1])
	}
	if args[2] != 0 {
		t.Errorf("offset for page 1 = %v, want 0", args[2])
	}
}

func TestBuildDataFetchQueryOffset(t *testing.T) {
	_, args := buildDataFetchQuery(testKey, DataFetchQueries{page: 3})
	// offset = (page-1) * pageSize
	if args[len(args)-1] != 100_000 {
		t.Errorf("offset for page 3 = %v, want 100000", args[len(args)-1])
	}
}

func TestBuildDataFetchQueryDateToUsesFullDay(t *testing.T) {
	// Regression: dateTo must include the whole end day, so it is compared with
	// "< $n::date + interval '1 days'", not "<= $n".
	query, args := buildDataFetchQuery(testKey, DataFetchQueries{
		page:     1,
		dateFrom: mustDate(t, "2024-06-01"),
		dateTo:   mustDate(t, "2024-06-30"),
	})

	if !strings.Contains(query, "r.created_at >= $2") {
		t.Errorf("missing dateFrom clause:\n%s", query)
	}
	if !strings.Contains(query, "r.created_at < $3::date + interval '1 days'") {
		t.Errorf("dateTo must be exclusive of the next day, got:\n%s", query)
	}
	if strings.Contains(query, "created_at <= $") {
		t.Errorf("dateTo should not use <=, got:\n%s", query)
	}
	if args[1] != "2024-06-01" || args[2] != "2024-06-30" {
		t.Errorf("date args = %v, %v", args[1], args[2])
	}
}

func TestBuildDataFetchQuerySingleDate(t *testing.T) {
	query, args := buildDataFetchQuery(testKey, DataFetchQueries{
		page: 1,
		date: mustDate(t, "2024-06-15"),
	})
	if !strings.Contains(query, "r.created_at >= $2 and r.created_at < $3::date + interval '1 days'") {
		t.Errorf("single date should span the whole day, got:\n%s", query)
	}
	// The same day value is bound twice.
	if args[1] != "2024-06-15" || args[2] != "2024-06-15" {
		t.Errorf("single-date args = %v, %v", args[1], args[2])
	}
}

func TestBuildDataFetchQueryValidFiltersApplied(t *testing.T) {
	query, args := buildDataFetchQuery(testKey, DataFetchQueries{
		page:      1,
		status:    200,
		ipAddress: "1.2.3.4",
		location:  "US",
		hostname:  "example.com",
		userID:    "user-1",
	})

	for _, want := range []string{
		"r.status = $",
		"r.ip_address = $",
		"r.location = $",
		"r.hostname = $",
		"r.user_id = $",
	} {
		if !strings.Contains(query, want) {
			t.Errorf("query missing %q:\n%s", want, query)
		}
	}
	for _, v := range []any{200, "1.2.3.4", "US", "example.com", "user-1"} {
		if !argsContain(args, v) {
			t.Errorf("args missing %v: %v", v, args)
		}
	}
}

func TestBuildDataFetchQueryInvalidFiltersDropped(t *testing.T) {
	// Values that fail validation must not appear as SQL clauses; they would
	// otherwise widen or break the query silently.
	query, _ := buildDataFetchQuery(testKey, DataFetchQueries{
		page:      1,
		status:    700, // out of 100-599 range
		ipAddress: "not-an-ip",
		location:  "usa", // must be two uppercase letters
	})

	for _, unwanted := range []string{"r.status = $", "r.ip_address = $", "r.location = $"} {
		if strings.Contains(query, unwanted) {
			t.Errorf("invalid filter leaked into query %q:\n%s", unwanted, query)
		}
	}
}

func TestBuildDataFetchQueryPlaceholdersAreSequential(t *testing.T) {
	// With two filters plus the trailing LIMIT/OFFSET the bind parameters must be
	// numbered $1..$5 with no gaps or repeats.
	query, args := buildDataFetchQuery(testKey, DataFetchQueries{
		page:     2,
		status:   404,
		hostname: "example.com",
	})

	for i := 1; i <= len(args); i++ {
		if !strings.Contains(query, "$"+strconv.Itoa(i)) {
			t.Errorf("query missing placeholder $%d:\n%s", i, query)
		}
	}
	if strings.Contains(query, "$"+strconv.Itoa(len(args)+1)) {
		t.Errorf("query references a placeholder with no argument:\n%s", query)
	}
	// Page 2 offset.
	if args[len(args)-1] != 50_000 {
		t.Errorf("offset = %v, want 50000", args[len(args)-1])
	}
}

func TestParseQueryDate(t *testing.T) {
	t.Run("valid date parses", func(t *testing.T) {
		got := parseQueryDate("2024-06-15")
		if got.IsZero() {
			t.Fatal("expected a non-zero time for a valid date")
		}
		if got.Year() != 2024 || got.Month() != time.June || got.Day() != 15 {
			t.Errorf("parsed %v, want 2024-06-15", got)
		}
	})

	for _, s := range []string{
		"",           // empty
		"not-a-date", // garbage
		"2024-13-40", // out-of-range month/day
		"2024/06/15", // wrong separators
		"15-06-2024", // wrong order
	} {
		t.Run("rejects "+s, func(t *testing.T) {
			if got := parseQueryDate(s); !got.IsZero() {
				t.Errorf("parseQueryDate(%q) = %v, want zero time", s, got)
			}
		})
	}
}

func argsContain(args []any, v any) bool {
	for _, a := range args {
		if a == v {
			return true
		}
	}
	return false
}
