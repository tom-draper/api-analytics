package email

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// sendGridSender delivers mail through the SendGrid v3 API.
type sendGridSender struct {
	apiKey      string
	fromAddress string
	endpoint    string // overridable in tests
}

// newSendGridFromEnv builds a SendGrid sender from SENDGRID_API_KEY and
// EMAIL_FROM_ADDRESS.
func newSendGridFromEnv() (*sendGridSender, error) {
	apiKey, err := requireEnv("SENDGRID_API_KEY")
	if err != nil {
		return nil, err
	}
	return &sendGridSender{
		apiKey:      apiKey,
		fromAddress: envDefaultFrom(),
		endpoint:    "https://api.sendgrid.com/v3/mail/send",
	}, nil
}

func (s *sendGridSender) Send(msg Message) error {
	if err := msg.resolveFrom(s.fromAddress); err != nil {
		return err
	}
	if err := msg.validate(); err != nil {
		return err
	}

	personalization := map[string]any{"to": sendGridAddrs(msg.To)}
	if len(msg.CC) > 0 {
		personalization["cc"] = sendGridAddrs(msg.CC)
	}
	if len(msg.BCC) > 0 {
		personalization["bcc"] = sendGridAddrs(msg.BCC)
	}

	payload := map[string]any{
		"personalizations": []any{personalization},
		"from":             map[string]string{"email": msg.From},
		"subject":          msg.Subject,
		"content": []any{
			map[string]string{"type": msg.contentType(), "value": msg.Body},
		},
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

	return doRequest(req, "sendgrid")
}

func sendGridAddrs(addresses []string) []map[string]string {
	out := make([]map[string]string, 0, len(addresses))
	for _, addr := range addresses {
		out = append(out, map[string]string{"email": addr})
	}
	return out
}
