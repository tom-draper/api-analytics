// Package alert emails an operator when a monitored URL changes between up and
// down. It is entirely optional: unless email alerts are enabled and an email
// provider and recipient are configured, every method is a no-op.
package alert

import (
	"fmt"
	"sync"
	"time"

	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/monitor/internal/log"
)

// Result is the outcome of pinging one monitored URL in a cycle.
type Result struct {
	APIKey string
	URL    string
	Status int
	Err    error // non-nil when the URL could not be reached
}

// down reports whether a result counts as downtime for alerting: the URL was
// unreachable, or the server returned a 5xx status. A 4xx is left out so that
// auth-protected endpoints answering HEAD with 401/403/405 do not alert.
func (r Result) down() bool {
	return r.Err != nil || r.Status >= 500
}

// Alerter tracks the up/down state of each monitored URL across cycles and
// emails on transitions. The zero value, and any Alerter from New with alerts
// disabled, is a safe no-op.
type Alerter struct {
	sender    email.Sender
	recipient string

	mu        sync.Mutex
	downState map[string]bool // "apiKey|url" -> currently considered down
}

// New builds an Alerter. It returns a disabled Alerter (Enabled reports false)
// when alerts are turned off, no email provider is configured (sender is nil),
// or no recipient is set, so callers can always call Evaluate safely.
func New(enabled bool, sender email.Sender, recipient string) *Alerter {
	if !enabled || sender == nil || recipient == "" {
		return &Alerter{}
	}
	return &Alerter{
		sender:    sender,
		recipient: recipient,
		downState: make(map[string]bool),
	}
}

// Enabled reports whether the Alerter will actually send email.
func (a *Alerter) Enabled() bool {
	return a.sender != nil && a.recipient != ""
}

// Evaluate compares each result against the previous cycle and emails on any
// up->down or down->up transition. Emails are sent after releasing the state
// lock so a slow provider does not stall the next cycle's bookkeeping.
func (a *Alerter) Evaluate(results []Result) {
	if !a.Enabled() {
		return
	}

	type change struct {
		result Result
		down   bool
	}
	var changes []change

	a.mu.Lock()
	for _, r := range results {
		key := r.APIKey + "|" + r.URL
		wasDown := a.downState[key]
		isDown := r.down()
		if isDown != wasDown {
			a.downState[key] = isDown
			changes = append(changes, change{result: r, down: isDown})
		}
	}
	a.mu.Unlock()

	for _, c := range changes {
		a.notify(c.result, c.down)
	}
}

func (a *Alerter) notify(r Result, isDown bool) {
	var subject, status string
	if isDown {
		subject = fmt.Sprintf("[API Analytics] %s is DOWN", r.URL)
		if r.Err != nil {
			status = "unreachable: " + r.Err.Error()
		} else {
			status = fmt.Sprintf("HTTP %d", r.Status)
		}
	} else {
		subject = fmt.Sprintf("[API Analytics] %s has recovered", r.URL)
		status = fmt.Sprintf("HTTP %d", r.Status)
	}

	body := fmt.Sprintf("URL: %s\nStatus: %s\nTime: %s\n", r.URL, status, time.Now().UTC().Format(time.RFC3339))
	if err := a.sender.Send(email.Message{
		To:      []string{a.recipient},
		Subject: subject,
		Body:    body,
	}); err != nil {
		log.Error(fmt.Sprintf("failed to send alert for %s: %v", r.URL, err))
		return
	}
	log.Info(fmt.Sprintf("sent alert: %s", subject))
}
