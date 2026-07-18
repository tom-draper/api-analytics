package alert

import (
	"errors"
	"strings"
	"testing"

	"github.com/tom-draper/api-analytics/server/email"
)

type fakeSender struct {
	sent []email.Message
}

func (f *fakeSender) Send(m email.Message) error {
	f.sent = append(f.sent, m)
	return nil
}

func TestResultDown(t *testing.T) {
	cases := []struct {
		result Result
		want   bool
	}{
		{Result{Status: 200}, false},
		{Result{Status: 301}, false},
		{Result{Status: 404}, false},
		{Result{Status: 499}, false},
		{Result{Status: 500}, true},
		{Result{Status: 503}, true},
		{Result{Err: errors.New("timeout")}, true},
	}
	for _, c := range cases {
		if got := c.result.down(); got != c.want {
			t.Errorf("down(status=%d, err=%v) = %v, want %v", c.result.Status, c.result.Err, got, c.want)
		}
	}
}

func TestAlerterDisabled(t *testing.T) {
	// A nil sender disables alerts even when enabled is true.
	a := New(true, nil, "ops@example.com")
	if a.Enabled() {
		t.Fatal("expected disabled Alerter with a nil sender")
	}
	// An empty recipient also disables.
	if New(true, &fakeSender{}, "").Enabled() {
		t.Fatal("expected disabled Alerter with an empty recipient")
	}
	// Not switched on.
	if New(false, &fakeSender{}, "ops@example.com").Enabled() {
		t.Fatal("expected disabled Alerter when alerts are off")
	}
	// Evaluate on a disabled Alerter must be a safe no-op.
	a.Evaluate([]Result{{URL: "u", Status: 500}})
}

func TestAlerterMessageContent(t *testing.T) {
	fs := &fakeSender{}
	a := New(true, fs, "ops@example.com")

	// A 5xx downtime reports the status code.
	a.Evaluate([]Result{{APIKey: "k1", URL: "https://a.example", Status: 503}})
	if len(fs.sent) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(fs.sent))
	}
	if !strings.Contains(fs.sent[0].Body, "HTTP 503") || !strings.Contains(fs.sent[0].Body, "https://a.example") {
		t.Errorf("status alert body missing detail: %q", fs.sent[0].Body)
	}

	// An unreachable URL reports the underlying error.
	a.Evaluate([]Result{{APIKey: "k2", URL: "https://b.example", Err: errors.New("dial timeout")}})
	if len(fs.sent) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(fs.sent))
	}
	if !strings.Contains(fs.sent[1].Body, "unreachable") || !strings.Contains(fs.sent[1].Body, "dial timeout") {
		t.Errorf("unreachable alert body missing detail: %q", fs.sent[1].Body)
	}
}

func TestAlerterTransitions(t *testing.T) {
	fs := &fakeSender{}
	a := New(true, fs, "ops@example.com")
	if !a.Enabled() {
		t.Fatal("expected enabled Alerter")
	}

	up := []Result{{APIKey: "k", URL: "u", Status: 200}}
	down := []Result{{APIKey: "k", URL: "u", Status: 500}}
	stillDown := []Result{{APIKey: "k", URL: "u", Status: 503}}

	// First observation is healthy: no alert.
	a.Evaluate(up)
	if len(fs.sent) != 0 {
		t.Fatalf("expected no alert for a healthy URL, got %d", len(fs.sent))
	}

	// up -> down: one DOWN alert.
	a.Evaluate(down)
	if len(fs.sent) != 1 {
		t.Fatalf("expected 1 alert on going down, got %d", len(fs.sent))
	}
	if !strings.Contains(fs.sent[0].Subject, "DOWN") {
		t.Errorf("expected a DOWN subject, got %q", fs.sent[0].Subject)
	}
	if fs.sent[0].To[0] != "ops@example.com" {
		t.Errorf("alert sent to %v", fs.sent[0].To)
	}

	// Staying down does not re-alert.
	a.Evaluate(stillDown)
	if len(fs.sent) != 1 {
		t.Fatalf("expected no repeat alert while still down, got %d", len(fs.sent))
	}

	// down -> up: one recovery alert.
	a.Evaluate(up)
	if len(fs.sent) != 2 {
		t.Fatalf("expected a recovery alert, got %d total", len(fs.sent))
	}
	if !strings.Contains(fs.sent[1].Subject, "recovered") {
		t.Errorf("expected a recovery subject, got %q", fs.sent[1].Subject)
	}
}
