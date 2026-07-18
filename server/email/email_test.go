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

func TestSMTPAuthMechanisms(t *testing.T) {
	base := &smtpSender{server: "smtp.example.com", username: "u", password: "p"}

	t.Run("login returns loginAuth", func(t *testing.T) {
		base.authType = AuthLogin
		auth, err := base.auth()
		if err != nil {
			t.Fatalf("auth: %v", err)
		}
		if _, ok := auth.(*loginAuth); !ok {
			t.Errorf("expected *loginAuth, got %T", auth)
		}
	})

	for name, at := range map[string]AuthType{"plain": AuthPlain, "crammd5": AuthCRAMMD5} {
		t.Run(name+" returns a non-nil auth", func(t *testing.T) {
			base.authType = at
			auth, err := base.auth()
			if err != nil || auth == nil {
				t.Errorf("auth() = (%v, %v), want a non-nil auth", auth, err)
			}
		})
	}

	t.Run("unknown auth type errors", func(t *testing.T) {
		base.authType = AuthType(99)
		if _, err := base.auth(); err == nil {
			t.Error("expected an error for an unknown auth type")
		}
	})
}

func TestLoginAuthProtocol(t *testing.T) {
	auth := &loginAuth{username: "user", password: "pass"}

	mech, resp, err := auth.Start(nil)
	if err != nil || mech != "LOGIN" || len(resp) != 0 {
		t.Fatalf("Start = (%q, %v, %v), want (LOGIN, empty, nil)", mech, resp, err)
	}

	if r, _ := auth.Next([]byte("Username:"), true); string(r) != "user" {
		t.Errorf("username prompt returned %q", r)
	}
	if r, _ := auth.Next([]byte("Password:"), true); string(r) != "pass" {
		t.Errorf("password prompt returned %q", r)
	}
	if _, err := auth.Next([]byte("Surprise:"), true); err == nil {
		t.Error("expected an error for an unknown prompt")
	}
	if r, err := auth.Next(nil, false); err != nil || r != nil {
		t.Errorf("end of exchange = (%v, %v), want (nil, nil)", r, err)
	}
}

func TestNewSMTPFromEnv(t *testing.T) {
	t.Run("defaults with From falling back to username", func(t *testing.T) {
		t.Setenv("EMAIL_USERNAME", "me@example.com")
		t.Setenv("EMAIL_PASSWORD", "secret")
		t.Setenv("EMAIL_FROM_ADDRESS", "")
		t.Setenv("SMTP_SERVER", "")
		t.Setenv("SMTP_PORT", "")
		t.Setenv("SMTP_AUTH_TYPE", "")

		s, err := newSMTPFromEnv()
		if err != nil {
			t.Fatalf("newSMTPFromEnv: %v", err)
		}
		if s.server != "smtp-mail.outlook.com" || s.port != 587 || s.authType != AuthLogin {
			t.Errorf("unexpected defaults: %+v", s)
		}
		if s.fromAddress != "me@example.com" {
			t.Errorf("From = %q, want the username fallback", s.fromAddress)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		t.Setenv("EMAIL_USERNAME", "me@example.com")
		t.Setenv("EMAIL_PASSWORD", "secret")
		t.Setenv("EMAIL_FROM_ADDRESS", "noreply@example.com")
		t.Setenv("SMTP_SERVER", "smtp.gmail.com")
		t.Setenv("SMTP_PORT", "465")
		t.Setenv("SMTP_AUTH_TYPE", "plain")

		s, err := newSMTPFromEnv()
		if err != nil {
			t.Fatalf("newSMTPFromEnv: %v", err)
		}
		if s.server != "smtp.gmail.com" || s.port != 465 || s.authType != AuthPlain {
			t.Errorf("custom values not applied: %+v", s)
		}
		if s.fromAddress != "noreply@example.com" {
			t.Errorf("From = %q", s.fromAddress)
		}
	})

	errorCases := []struct {
		name string
		env  map[string]string
	}{
		{"missing username", map[string]string{"EMAIL_USERNAME": "", "EMAIL_PASSWORD": "p"}},
		{"missing password", map[string]string{"EMAIL_USERNAME": "u", "EMAIL_PASSWORD": ""}},
		{"invalid port", map[string]string{"EMAIL_USERNAME": "u", "EMAIL_PASSWORD": "p", "SMTP_PORT": "not-a-port"}},
		{"invalid auth type", map[string]string{"EMAIL_USERNAME": "u", "EMAIL_PASSWORD": "p", "SMTP_AUTH_TYPE": "magic"}},
	}
	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			if _, err := newSMTPFromEnv(); err == nil {
				t.Errorf("expected an error for %s", tc.name)
			}
		})
	}
}

func TestResendHTMLUsesHtmlField(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender := &resendSender{apiKey: "re_test", fromAddress: "from@example.com", endpoint: srv.URL}
	err := sender.Send(Message{
		To: []string{"to@example.com"}, Subject: "S",
		Body: "<h1>Hi</h1>", ContentType: "text/html",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if _, ok := body["html"]; !ok {
		t.Errorf("html message should set the html field, got: %v", body)
	}
	if _, ok := body["text"]; ok {
		t.Errorf("html message must not set the text field, got: %v", body)
	}
}

func TestMessageRecipients(t *testing.T) {
	m := Message{
		To:  []string{"a@x.com", "b@x.com"},
		CC:  []string{"c@x.com"},
		BCC: []string{"d@x.com"},
	}
	got := m.recipients()
	want := []string{"a@x.com", "b@x.com", "c@x.com", "d@x.com"}
	if len(got) != len(want) {
		t.Fatalf("recipients = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("recipients[%d] = %q, want %q", i, got[i], want[i])
		}
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
