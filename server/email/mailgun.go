package email

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// mailgunSender delivers mail through the Mailgun HTTP API.
type mailgunSender struct {
	apiKey      string
	domain      string
	baseURL     string // e.g. https://api.mailgun.net (US) or https://api.eu.mailgun.net (EU)
	fromAddress string
}

// newMailgunFromEnv builds a Mailgun sender from MAILGUN_API_KEY,
// MAILGUN_DOMAIN, EMAIL_FROM_ADDRESS and the optional MAILGUN_BASE_URL (set to
// the EU host for EU accounts).
func newMailgunFromEnv() (*mailgunSender, error) {
	apiKey, err := requireEnv("MAILGUN_API_KEY")
	if err != nil {
		return nil, err
	}
	domain, err := requireEnv("MAILGUN_DOMAIN")
	if err != nil {
		return nil, err
	}
	return &mailgunSender{
		apiKey:      apiKey,
		domain:      domain,
		baseURL:     strings.TrimRight(getEnvWithDefault("MAILGUN_BASE_URL", "https://api.mailgun.net"), "/"),
		fromAddress: envDefaultFrom(),
	}, nil
}

func (s *mailgunSender) Send(msg Message) error {
	if err := msg.resolveFrom(s.fromAddress); err != nil {
		return err
	}
	if err := msg.validate(); err != nil {
		return err
	}

	form := url.Values{}
	form.Set("from", msg.From)
	for _, to := range msg.To {
		form.Add("to", to)
	}
	for _, cc := range msg.CC {
		form.Add("cc", cc)
	}
	for _, bcc := range msg.BCC {
		form.Add("bcc", bcc)
	}
	form.Set("subject", msg.Subject)
	if msg.isHTML() {
		form.Set("html", msg.Body)
	} else {
		form.Set("text", msg.Body)
	}

	endpoint := fmt.Sprintf("%s/v3/%s/messages", s.baseURL, s.domain)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", s.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return doRequest(req, "mailgun")
}
