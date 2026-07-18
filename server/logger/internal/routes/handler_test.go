package routes

import (
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/tom-draper/api-analytics/server/database"
)

func TestTruncateExtremeUnicode(t *testing.T) {
	emoji := "😀" // one 4-byte rune

	// Truncating below the rune's byte size drops it entirely rather than
	// splitting it, which would leave invalid UTF-8.
	for n := 0; n < 4; n++ {
		if got := truncate(emoji, n); got != "" {
			t.Errorf("truncate(4-byte rune, %d) = %q, want empty (a rune cannot be split)", n, got)
		}
	}
	if truncate(emoji, 4) != emoji {
		t.Error("truncate at the exact rune size should keep the whole rune")
	}

	// Property: truncating a run of multi-byte runes at ANY byte length must stay
	// valid UTF-8 and never exceed the requested length.
	s := strings.Repeat("😀", 8) + strings.Repeat("é", 8) // mixed 4-byte and 2-byte runes
	for n := 0; n <= len(s); n++ {
		got := truncate(s, n)
		if len(got) > n {
			t.Fatalf("truncate(...,%d) length %d exceeds the limit", n, len(got))
		}
		if !utf8.ValidString(got) {
			t.Fatalf("truncate(...,%d) produced invalid UTF-8: %q", n, got)
		}
	}
}

func TestIPUsageForPrivacyExtremes(t *testing.T) {
	// The most-negative level is treated as level 0 (least private): full IP use.
	if u := ipUsageForPrivacy(PrivacyLevel(math.MinInt)); !u.store || !u.inferLocation || !u.hash {
		t.Errorf("MinInt privacy = %+v, want full IP usage", u)
	}
	// The most-positive level is treated as the most private: no IP use at all.
	if u := ipUsageForPrivacy(PrivacyLevel(math.MaxInt)); u.store || u.inferLocation || u.hash {
		t.Errorf("MaxInt privacy = %+v, want no IP usage", u)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		n        int
		expected string
	}{
		{
			name:     "shorter than limit is unchanged",
			value:    "/api/v1/users",
			n:        255,
			expected: "/api/v1/users",
		},
		{
			name:     "exactly at limit is unchanged",
			value:    strings.Repeat("a", 255),
			n:        255,
			expected: strings.Repeat("a", 255),
		},
		{
			name:     "ascii truncated at limit",
			value:    strings.Repeat("a", 300),
			n:        255,
			expected: strings.Repeat("a", 255),
		},
		{
			name:     "multi-byte rune not split at boundary",
			value:    strings.Repeat("a", 254) + "é",
			n:        255,
			expected: strings.Repeat("a", 254),
		},
		{
			name:     "value of only multi-byte runes",
			value:    strings.Repeat("日", 10),
			n:        4,
			expected: "日",
		},
		{
			name:     "empty value",
			value:    "",
			n:        255,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.value, tt.n)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, expected %q", tt.value, tt.n, result, tt.expected)
			}
			if len(result) > tt.n {
				t.Errorf("truncate(%q, %d) returned %d bytes, over the limit", tt.value, tt.n, len(result))
			}
			if !utf8.ValidString(result) {
				t.Errorf("truncate(%q, %d) = %q, which is not valid UTF-8", tt.value, tt.n, result)
			}
		})
	}
}

// Truncation runs before validation at ingest, so a value cut mid-rune would be
// rejected as invalid UTF-8 and the request silently dropped.
func TestTruncatedValuesStayValid(t *testing.T) {
	values := []string{
		strings.Repeat("a", 254) + "é",
		strings.Repeat("日", 200),
		"/api/" + strings.Repeat("café/", 100),
		strings.Repeat("Mozilla/5.0 (Windows NT 10.0; Win64; x64) ", 20),
	}

	for _, value := range values {
		truncated := truncate(value, 255)
		if !database.ValidString(truncated) {
			t.Errorf("ValidString(truncate(%.30q..., 255)) = false, expected a truncated value to stay valid", value)
		}
	}
}

func TestApplyUserAgentIDs(t *testing.T) {
	requests := []ProcessedRequest{
		{Path: "/known/one", UserAgent: "curl/8.4.0"},
		{Path: "/new/one", UserAgent: "brand-new-agent/1.0"},
		{Path: "/known/two", UserAgent: "known-agent/2.0"},
		{Path: "/new/two", UserAgent: "another-new-agent/1.0"},
	}
	ids := map[string]int{"curl/8.4.0": 1, "known-agent/2.0": 2}

	resolved, unresolved := applyUserAgentIDs(requests, ids)

	if unresolved != 2 {
		t.Errorf("unresolved = %d, expected 2", unresolved)
	}
	if len(resolved) != 2 {
		t.Fatalf("len(resolved) = %d, expected 2", len(resolved))
	}
	// Filtering happens in place, so a mis-ordered write would corrupt the
	// surviving entries rather than merely drop the wrong ones.
	if resolved[0].Path != "/known/one" || resolved[0].UserAgentID != 1 {
		t.Errorf("resolved[0] = %+v, expected /known/one with id 1", resolved[0])
	}
	if resolved[1].Path != "/known/two" || resolved[1].UserAgentID != 2 {
		t.Errorf("resolved[1] = %+v, expected /known/two with id 2", resolved[1])
	}
}

func TestApplyUserAgentIDsAllResolve(t *testing.T) {
	requests := []ProcessedRequest{
		{Path: "/a", UserAgent: "curl/8.4.0"},
		{Path: "/b", UserAgent: "known-agent/2.0"},
	}
	ids := map[string]int{"curl/8.4.0": 1, "known-agent/2.0": 2}

	resolved, unresolved := applyUserAgentIDs(requests, ids)

	if unresolved != 0 {
		t.Errorf("unresolved = %d, expected 0", unresolved)
	}
	if len(resolved) != 2 {
		t.Errorf("len(resolved) = %d, expected every request to survive", len(resolved))
	}
	for _, request := range resolved {
		if request.UserAgentID == 0 {
			t.Errorf("%s kept user_agent_id 0, which would fail the foreign key", request.Path)
		}
	}
}

func TestApplyUserAgentIDsNoneResolve(t *testing.T) {
	requests := []ProcessedRequest{
		{Path: "/a", UserAgent: "brand-new-agent/1.0"},
		{Path: "/b", UserAgent: "another-new-agent/1.0"},
	}

	resolved, unresolved := applyUserAgentIDs(requests, map[string]int{})

	if unresolved != 2 {
		t.Errorf("unresolved = %d, expected 2", unresolved)
	}
	if len(resolved) != 0 {
		t.Errorf("len(resolved) = %d, expected 0", len(resolved))
	}
}

func TestRealUserAgentsAccepted(t *testing.T) {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/121.0",
		"curl/8.4.0",
	}

	for _, userAgent := range userAgents {
		if !database.ValidUserAgent(truncate(userAgent, 255)) {
			t.Errorf("ValidUserAgent(%q) = false, expected real browser traffic to be logged", userAgent)
		}
	}
}

func TestIPUsageForPrivacy(t *testing.T) {
	tests := []struct {
		name  string
		level PrivacyLevel
		want  ipUsage
	}{
		{"level 0 stores IP and infers location", P1, ipUsage{store: true, inferLocation: true, hash: true}},
		{"level 1 discards IP but still infers location", P2, ipUsage{store: false, inferLocation: true, hash: true}},
		{"level 2 never accesses IP or infers location", P3, ipUsage{store: false, inferLocation: false, hash: false}},
		{"out-of-range level gets the most private treatment", P3 + 5, ipUsage{}},
		{"negative level is treated as level 0", -1, ipUsage{store: true, inferLocation: true, hash: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ipUsageForPrivacy(tt.level); got != tt.want {
				t.Errorf("ipUsageForPrivacy(%d) = %+v, want %+v", tt.level, got, tt.want)
			}
		})
	}
}
