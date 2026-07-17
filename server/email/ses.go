package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// sesSender delivers mail through the Amazon SES v2 API, signing requests with
// Signature Version 4. It uses the standard AWS credential environment
// variables, so it needs no AWS SDK.
type sesSender struct {
	creds       awsCredentials
	fromAddress string
	endpoint    string // overridable in tests
	now         func() time.Time
}

// newSESFromEnv builds an SES sender from AWS_ACCESS_KEY_ID,
// AWS_SECRET_ACCESS_KEY, AWS_REGION (or AWS_DEFAULT_REGION), the optional
// AWS_SESSION_TOKEN, and EMAIL_FROM_ADDRESS.
func newSESFromEnv() (*sesSender, error) {
	accessKey, err := requireEnv("AWS_ACCESS_KEY_ID")
	if err != nil {
		return nil, err
	}
	secretKey, err := requireEnv("AWS_SECRET_ACCESS_KEY")
	if err != nil {
		return nil, err
	}
	region := getEnvWithDefault("AWS_REGION", getEnvWithDefault("AWS_DEFAULT_REGION", ""))
	if region == "" {
		return nil, fmt.Errorf("AWS_REGION is required")
	}

	return &sesSender{
		creds: awsCredentials{
			accessKeyID:     accessKey,
			secretAccessKey: secretKey,
			sessionToken:    getEnvWithDefault("AWS_SESSION_TOKEN", ""),
			region:          region,
		},
		fromAddress: envDefaultFrom(),
		endpoint:    fmt.Sprintf("https://email.%s.amazonaws.com/v2/email/outbound-emails", region),
		now:         time.Now,
	}, nil
}

func (s *sesSender) Send(msg Message) error {
	if err := msg.resolveFrom(s.fromAddress); err != nil {
		return err
	}
	if err := msg.validate(); err != nil {
		return err
	}

	content := map[string]any{"Data": msg.Body, "Charset": "UTF-8"}
	bodyKey := "Text"
	if msg.isHTML() {
		bodyKey = "Html"
	}

	destination := map[string]any{"ToAddresses": msg.To}
	if len(msg.CC) > 0 {
		destination["CcAddresses"] = msg.CC
	}
	if len(msg.BCC) > 0 {
		destination["BccAddresses"] = msg.BCC
	}

	payload := map[string]any{
		"FromEmailAddress": msg.From,
		"Destination":      destination,
		"Content": map[string]any{
			"Simple": map[string]any{
				"Subject": map[string]any{"Data": msg.Subject, "Charset": "UTF-8"},
				"Body":    map[string]any{bodyKey: content},
			},
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
	req.Header.Set("Content-Type", "application/json")
	signV4(req, body, s.creds, "ses", s.now())

	return doRequest(req, "ses")
}
