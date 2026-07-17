package email

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// resendSender delivers mail through the Resend HTTP API (https://resend.com).
type resendSender struct {
	apiKey      string
	fromAddress string
	endpoint    string // overridable in tests
}

// newResendFromEnv builds a Resend sender from RESEND_API_KEY and
// EMAIL_FROM_ADDRESS.
func newResendFromEnv() (*resendSender, error) {
	apiKey, err := requireEnv("RESEND_API_KEY")
	if err != nil {
		return nil, err
	}
	return &resendSender{
		apiKey:      apiKey,
		fromAddress: envDefaultFrom(),
		endpoint:    "https://api.resend.com/emails",
	}, nil
}

func (s *resendSender) Send(msg Message) error {
	if err := msg.resolveFrom(s.fromAddress); err != nil {
		return err
	}
	if err := msg.validate(); err != nil {
		return err
	}

	payload := map[string]any{
		"from":    msg.From,
		"to":      msg.To,
		"subject": msg.Subject,
	}
	if len(msg.CC) > 0 {
		payload["cc"] = msg.CC
	}
	if len(msg.BCC) > 0 {
		payload["bcc"] = msg.BCC
	}
	if msg.isHTML() {
		payload["html"] = msg.Body
	} else {
		payload["text"] = msg.Body
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return doRequest(req, "resend")
}
