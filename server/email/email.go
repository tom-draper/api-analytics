// Package email provides an optional, provider-agnostic way to send email.
//
// A Sender is selected from the environment with NewFromEnv, driven by
// EMAIL_PROVIDER (smtp, resend, sendgrid, mailgun or ses). When the variable is
// unset or "none", email is considered disabled and NewFromEnv returns a nil
// Sender, so callers can treat email as an optional feature:
//
//	sender, err := email.NewFromEnv()
//	if err != nil {
//		// misconfiguration
//	}
//	if sender == nil {
//		// email disabled; skip
//	} else {
//		err = sender.Send(msg)
//	}
package email

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Message is a provider-agnostic email. Body is interpreted according to
// ContentType, which defaults to "text/plain" when empty.
type Message struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	ContentType string // "text/plain" (default) or "text/html"
}

// Sender delivers a Message through a specific provider.
type Sender interface {
	Send(msg Message) error
}

// ErrDisabled is returned by helpers that require a configured provider when
// none is set.
var ErrDisabled = errors.New("email is disabled: set EMAIL_PROVIDER to enable")

// NewFromEnv builds a Sender from EMAIL_PROVIDER and the provider's own
// environment variables. It returns (nil, nil) when email is disabled
// (EMAIL_PROVIDER unset or "none"), and an error only when a provider is named
// but misconfigured.
func NewFromEnv() (Sender, error) {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_PROVIDER")))
	switch provider {
	case "", "none", "off", "disabled":
		return nil, nil
	case "smtp":
		return newSMTPFromEnv()
	case "resend":
		return newResendFromEnv()
	case "sendgrid":
		return newSendGridFromEnv()
	case "mailgun":
		return newMailgunFromEnv()
	case "ses", "aws-ses":
		return newSESFromEnv()
	default:
		return nil, fmt.Errorf("unsupported EMAIL_PROVIDER %q (want smtp, resend, sendgrid, mailgun or ses)", provider)
	}
}

// contentType returns the message content type, defaulting to text/plain. This
// default lives here rather than only in a provider so that a message rendered
// anywhere carries a valid type instead of an empty "Content-Type: ; ...".
func (m Message) contentType() string {
	if m.ContentType == "" {
		return "text/plain"
	}
	return m.ContentType
}

// isHTML reports whether the body should be sent as HTML.
func (m Message) isHTML() bool {
	return strings.HasPrefix(m.contentType(), "text/html")
}

// recipients returns every address the message is delivered to.
func (m Message) recipients() []string {
	all := make([]string, 0, len(m.To)+len(m.CC)+len(m.BCC))
	all = append(all, m.To...)
	all = append(all, m.CC...)
	all = append(all, m.BCC...)
	return all
}

// resolveFrom fills an empty From with the provider's default address and
// reports an error if neither is set.
func (m *Message) resolveFrom(defaultFrom string) error {
	if m.From == "" {
		m.From = defaultFrom
	}
	if m.From == "" {
		return errors.New("no From address: set the message From or EMAIL_FROM_ADDRESS")
	}
	return nil
}

// validate checks the fields every provider requires.
func (m Message) validate() error {
	if len(m.To) == 0 {
		return errors.New("at least one recipient is required")
	}
	return nil
}

// envDefaultFrom is the shared fallback From address for every provider.
func envDefaultFrom() string {
	return os.Getenv("EMAIL_FROM_ADDRESS")
}

// requireEnv returns the value of key or an error naming it, so provider
// constructors fail with a clear message about what is missing.
func requireEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
