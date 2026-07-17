package email

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/smtp"
	"strings"
	"testing"
	"time"
)

func TestMessageContentTypeDefaults(t *testing.T) {
	if got := (Message{}).contentType(); got != "text/plain" {
		t.Errorf("empty ContentType = %q, want text/plain", got)
	}
	if got := (Message{ContentType: "text/html"}).contentType(); got != "text/html" {
		t.Errorf("ContentType = %q, want text/html", got)
	}
	if !(Message{ContentType: "text/html; charset=UTF-8"}).isHTML() {
		t.Error("expected isHTML for a text/html message")
	}
	if (Message{}).isHTML() {
		t.Error("did not expect isHTML for a default (text/plain) message")
	}
}

// Regression: buildMessage previously emitted "Content-Type: ; charset=UTF-8"
// for a message with no ContentType, because the default was applied only in
// Send. renderMIME must always carry a valid type.
func TestRenderMIMEContentType(t *testing.T) {
	mime := string(renderMIME(Message{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "Subject",
		Body:    "Body",
	}))

	if !strings.Contains(mime, "Content-Type: text/plain; charset=UTF-8") {
		t.Errorf("missing defaulted content type, got:\n%s", mime)
	}
	if strings.Contains(mime, "Content-Type: ;") {
		t.Errorf("emitted an empty content type, got:\n%s", mime)
	}

	htmlMIME := string(renderMIME(Message{
		From: "from@example.com", To: []string{"to@example.com"},
		Subject: "S", Body: "<h1>Hi</h1>", ContentType: "text/html",
	}))
	if !strings.Contains(htmlMIME, "Content-Type: text/html; charset=UTF-8") {
		t.Errorf("missing html content type, got:\n%s", htmlMIME)
	}
}

func TestRenderMIMEHeaders(t *testing.T) {
	mime := string(renderMIME(Message{
		From:    "from@example.com",
		To:      []string{"a@example.com", "b@example.com"},
		CC:      []string{"cc@example.com"},
		BCC:     []string{"bcc@example.com"},
		Subject: "Hello",
		Body:    "Body",
	}))

	for _, want := range []string{
		"From: from@example.com",
		"To: a@example.com, b@example.com",
		"CC: cc@example.com",
		"Subject: Hello",
	} {
		if !strings.Contains(mime, want) {
			t.Errorf("MIME missing %q, got:\n%s", want, mime)
		}
	}
	if strings.Contains(mime, "BCC:") {
		t.Error("BCC must not appear in message headers")
	}
}

func TestMessageResolveFromAndValidate(t *testing.T) {
	m := Message{To: []string{"to@example.com"}}
	if err := m.resolveFrom("default@example.com"); err != nil {
		t.Fatalf("resolveFrom: %v", err)
	}
	if m.From != "default@example.com" {
		t.Errorf("From = %q, want default@example.com", m.From)
	}

	explicit := Message{From: "explicit@example.com"}
	_ = explicit.resolveFrom("default@example.com")
	if explicit.From != "explicit@example.com" {
		t.Errorf("resolveFrom overrode an explicit From: %q", explicit.From)
	}

	if err := (&Message{}).resolveFrom(""); err == nil {
		t.Error("expected error when no From is available")
	}
	if err := (Message{From: "f@x.com"}).validate(); err == nil {
		t.Error("expected error for a message with no recipients")
	}
}

func TestNewFromEnvRouting(t *testing.T) {
	t.Run("disabled when unset", func(t *testing.T) {
		t.Setenv("EMAIL_PROVIDER", "")
		sender, err := NewFromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sender != nil {
			t.Errorf("expected nil (disabled) sender, got %T", sender)
		}
	})

	t.Run("explicit none is disabled", func(t *testing.T) {
		t.Setenv("EMAIL_PROVIDER", "none")
		sender, err := NewFromEnv()
		if err != nil || sender != nil {
			t.Errorf("expected (nil, nil), got (%T, %v)", sender, err)
		}
	})

	t.Run("unknown provider errors", func(t *testing.T) {
		t.Setenv("EMAIL_PROVIDER", "pigeon")
		if _, err := NewFromEnv(); err == nil {
			t.Error("expected error for an unknown provider")
		}
	})

	t.Run("resend needs an api key", func(t *testing.T) {
		t.Setenv("EMAIL_PROVIDER", "resend")
		t.Setenv("RESEND_API_KEY", "")
		if _, err := NewFromEnv(); err == nil {
			t.Error("expected error when RESEND_API_KEY is missing")
		}

		t.Setenv("RESEND_API_KEY", "re_test")
		sender, err := NewFromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := sender.(*resendSender); !ok {
			t.Errorf("expected *resendSender, got %T", sender)
		}
	})

	t.Run("smtp needs credentials", func(t *testing.T) {
		t.Setenv("EMAIL_PROVIDER", "smtp")
		t.Setenv("EMAIL_USERNAME", "")
		t.Setenv("EMAIL_PASSWORD", "")
		if _, err := NewFromEnv(); err == nil {
			t.Error("expected error when SMTP credentials are missing")
		}
	})
}

func TestSMTPSend(t *testing.T) {
	var gotAddr, gotFrom string
	var gotTo []string
	var gotMsg []byte

	sender := &smtpSender{
		server: "smtp.example.com", port: 587,
		username: "user@example.com", password: "pw",
		fromAddress: "from@example.com", authType: AuthLogin,
		sendMail: func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
			gotAddr, gotFrom, gotTo, gotMsg = addr, from, to, msg
			return nil
		},
	}

	err := sender.Send(Message{
		To:      []string{"to@example.com"},
		CC:      []string{"cc@example.com"},
		BCC:     []string{"bcc@example.com"},
		Subject: "Hi",
		Body:    "Body",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if gotAddr != "smtp.example.com:587" {
		t.Errorf("addr = %q", gotAddr)
	}
	if gotFrom != "from@example.com" {
		t.Errorf("from = %q, want the default from address", gotFrom)
	}
	if len(gotTo) != 3 {
		t.Errorf("expected 3 envelope recipients (to+cc+bcc), got %v", gotTo)
	}
	if !strings.Contains(string(gotMsg), "Subject: Hi") {
		t.Errorf("message missing subject:\n%s", gotMsg)
	}

	if err := sender.Send(Message{Subject: "no recipients"}); err == nil {
		t.Error("expected error for a message with no recipients")
	}
}

func TestResendSend(t *testing.T) {
	var body map[string]any
	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender := &resendSender{apiKey: "re_test", fromAddress: "from@example.com", endpoint: srv.URL}
	if err := sender.Send(Message{To: []string{"to@example.com"}, Subject: "S", Body: "B"}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if authHeader != "Bearer re_test" {
		t.Errorf("Authorization = %q", authHeader)
	}
	if body["subject"] != "S" || body["from"] != "from@example.com" {
		t.Errorf("unexpected payload: %v", body)
	}
	if _, ok := body["text"]; !ok {
		t.Errorf("plain message should populate text, got: %v", body)
	}
}

func TestSendGridSend(t *testing.T) {
	var raw []byte
	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		raw, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	sender := &sendGridSender{apiKey: "sg_test", fromAddress: "from@example.com", endpoint: srv.URL}
	if err := sender.Send(Message{To: []string{"to@example.com"}, Subject: "S", Body: "B"}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if authHeader != "Bearer sg_test" {
		t.Errorf("Authorization = %q", authHeader)
	}
	for _, want := range []string{"personalizations", "to@example.com", "\"subject\":\"S\""} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("payload missing %q: %s", want, raw)
		}
	}
}

func TestMailgunSend(t *testing.T) {
	var path, authHeader, form string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		authHeader = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		form = string(raw)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender := &mailgunSender{apiKey: "key-test", domain: "mail.example.com", baseURL: srv.URL, fromAddress: "from@example.com"}
	if err := sender.Send(Message{To: []string{"to@example.com"}, Subject: "S", Body: "B"}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if path != "/v3/mail.example.com/messages" {
		t.Errorf("path = %q", path)
	}
	if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
		t.Errorf("expected basic auth, got %q", authHeader)
	}
	if !strings.Contains(form, "to=to%40example.com") || !strings.Contains(form, "subject=S") {
		t.Errorf("unexpected form body: %s", form)
	}
}

func TestSESSendSignsRequest(t *testing.T) {
	var authHeader, amzDate, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		amzDate = r.Header.Get("X-Amz-Date")
		path = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	fixed := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	sender := &sesSender{
		creds:       awsCredentials{accessKeyID: "AKID", secretAccessKey: "secret", region: "eu-west-1"},
		fromAddress: "from@example.com",
		endpoint:    srv.URL + "/v2/email/outbound-emails",
		now:         func() time.Time { return fixed },
	}
	if err := sender.Send(Message{To: []string{"to@example.com"}, Subject: "S", Body: "B"}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if path != "/v2/email/outbound-emails" {
		t.Errorf("path = %q", path)
	}
	if amzDate != "20260102T030405Z" {
		t.Errorf("X-Amz-Date = %q", amzDate)
	}
	if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256 Credential=AKID/20260102/eu-west-1/ses/aws4_request") {
		t.Errorf("unexpected Authorization: %q", authHeader)
	}
	if !strings.Contains(authHeader, "SignedHeaders=") || !strings.Contains(authHeader, "Signature=") {
		t.Errorf("Authorization missing SignedHeaders/Signature: %q", authHeader)
	}
}

func TestSigV4Deterministic(t *testing.T) {
	build := func(body string) string {
		req, _ := http.NewRequest(http.MethodPost, "https://email.eu-west-1.amazonaws.com/v2/email/outbound-emails", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		signV4(req, []byte(body), awsCredentials{accessKeyID: "AKID", secretAccessKey: "secret", region: "eu-west-1"}, "ses", time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
		return req.Header.Get("Authorization")
	}

	a, b := build("payload-one"), build("payload-one")
	if a != b {
		t.Errorf("signing not deterministic:\n%s\n%s", a, b)
	}
	if a == build("payload-two") {
		t.Error("signature did not change with the payload")
	}
}

func TestDoRequestNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("bad key"))
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL, nil)
	err := doRequest(req, "test")
	if err == nil || !strings.Contains(err.Error(), "bad key") {
		t.Errorf("expected error carrying the response body, got %v", err)
	}
}
