package email

import (
	"errors"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
)

// AuthType selects the SMTP authentication mechanism.
type AuthType int

const (
	AuthLogin AuthType = iota
	AuthPlain
	AuthCRAMMD5
)

// smtpSender sends mail through an SMTP server (self-hosted or any provider that
// exposes SMTP). It satisfies Sender.
type smtpSender struct {
	server      string
	port        int
	username    string
	password    string
	fromAddress string
	authType    AuthType

	// sendMail is smtp.SendMail in production; overridable in tests.
	sendMail func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// newSMTPFromEnv builds an SMTP sender from SMTP_SERVER, SMTP_PORT,
// EMAIL_USERNAME, EMAIL_PASSWORD, SMTP_AUTH_TYPE and EMAIL_FROM_ADDRESS.
func newSMTPFromEnv() (*smtpSender, error) {
	username, err := requireEnv("EMAIL_USERNAME")
	if err != nil {
		return nil, err
	}
	password, err := requireEnv("EMAIL_PASSWORD")
	if err != nil {
		return nil, err
	}

	portStr := getEnvWithDefault("SMTP_PORT", "587")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	var authType AuthType
	switch strings.ToLower(getEnvWithDefault("SMTP_AUTH_TYPE", "login")) {
	case "login":
		authType = AuthLogin
	case "plain":
		authType = AuthPlain
	case "crammd5":
		authType = AuthCRAMMD5
	default:
		return nil, fmt.Errorf("unsupported SMTP_AUTH_TYPE: %s", getEnvWithDefault("SMTP_AUTH_TYPE", "login"))
	}

	from := envDefaultFrom()
	if from == "" {
		from = username
	}

	return &smtpSender{
		server:      getEnvWithDefault("SMTP_SERVER", "smtp-mail.outlook.com"),
		port:        port,
		username:    username,
		password:    password,
		fromAddress: from,
		authType:    authType,
		sendMail:    smtp.SendMail,
	}, nil
}

func (s *smtpSender) Send(msg Message) error {
	if err := msg.resolveFrom(s.fromAddress); err != nil {
		return err
	}
	if err := msg.validate(); err != nil {
		return err
	}

	auth, err := s.auth()
	if err != nil {
		return fmt.Errorf("failed to create auth: %w", err)
	}

	endpoint := fmt.Sprintf("%s:%d", s.server, s.port)
	if err := s.sendMail(endpoint, auth, msg.From, msg.recipients(), renderMIME(msg)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (s *smtpSender) auth() (smtp.Auth, error) {
	switch s.authType {
	case AuthLogin:
		return &loginAuth{username: s.username, password: s.password}, nil
	case AuthPlain:
		return smtp.PlainAuth("", s.username, s.password, s.server), nil
	case AuthCRAMMD5:
		return smtp.CRAMMD5Auth(s.username, s.password), nil
	default:
		return nil, fmt.Errorf("unsupported auth type: %v", s.authType)
	}
}

// renderMIME builds the RFC 822 message for SMTP delivery. ContentType defaults
// to text/plain via Message.contentType so the header is never emitted with an
// empty type.
func renderMIME(msg Message) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", msg.From))
	b.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	if len(msg.CC) > 0 {
		b.WriteString(fmt.Sprintf("CC: %s\r\n", strings.Join(msg.CC, ", ")))
	}
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	b.WriteString(fmt.Sprintf("Content-Type: %s; charset=UTF-8\r\n", msg.contentType()))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Body)
	return []byte(b.String())
}

// loginAuth implements the LOGIN SMTP auth mechanism, which net/smtp does not
// provide out of the box but many servers (e.g. Outlook) require.
type loginAuth struct {
	username string
	password string
}

func (a *loginAuth) Start(*smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unknown server response")
		}
	}
	return nil, nil
}
