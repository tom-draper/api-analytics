package email

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	SMTPServer  string
	SMTPPort    int
	Username    string
	Password    string
	FromAddress string
	AuthType    AuthType
}

type AuthType int

const (
	AuthLogin AuthType = iota
	AuthPlain
	AuthCRAMMD5
)

type Message struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	ContentType string // "text/plain" or "text/html"
}

type Client struct {
	config Config
}

func NewClient(config Config) *Client {
	return &Client{config: config}
}

func NewClientFromEnv(envFile ...string) (*Client, error) {
	envPath := ".env"
	if len(envFile) > 0 {
		envPath = envFile[0]
	}

	if err := godotenv.Load(envPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load env file: %w", err)
		}
	}

	config := Config{
		SMTPServer:  getEnvWithDefault("SMTP_SERVER", "smtp-mail.outlook.com"),
		Username:    os.Getenv("EMAIL_USERNAME"),
		Password:    os.Getenv("EMAIL_PASSWORD"),
		FromAddress: os.Getenv("EMAIL_FROM_ADDRESS"),
	}

	portStr := getEnvWithDefault("SMTP_PORT", "587")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}
	config.SMTPPort = port

	authTypeStr := getEnvWithDefault("SMTP_AUTH_TYPE", "login")
	switch strings.ToLower(authTypeStr) {
	case "login":
		config.AuthType = AuthLogin
	case "plain":
		config.AuthType = AuthPlain
	case "crammd5":
		config.AuthType = AuthCRAMMD5
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authTypeStr)
	}

	if config.FromAddress == "" {
		config.FromAddress = config.Username
	}

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("EMAIL_USERNAME and EMAIL_PASSWORD must be set")
	}

	return NewClient(config), nil
}

func (c *Client) Send(msg Message) error {
	if msg.From == "" {
		msg.From = c.config.FromAddress
	}

	if len(msg.To) == 0 {
		return errors.New("at least one recipient is required")
	}

	if msg.ContentType == "" {
		msg.ContentType = "text/plain"
	}

	auth, err := c.createAuth()
	if err != nil {
		return fmt.Errorf("failed to create auth: %w", err)
	}

	message := c.buildMessage(msg)

	allRecipients := append(msg.To, msg.CC...)
	allRecipients = append(allRecipients, msg.BCC...)

	endpoint := fmt.Sprintf("%s:%d", c.config.SMTPServer, c.config.SMTPPort)
	err = smtp.SendMail(endpoint, auth, msg.From, allRecipients, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (c *Client) SendSimple(to, subject, body string) error {
	msg := Message{
		To:      []string{to},
		Subject: subject,
		Body:    body,
	}
	return c.Send(msg)
}

func (c *Client) createAuth() (smtp.Auth, error) {
	switch c.config.AuthType {
	case AuthLogin:
		return &loginAuth{
			username: c.config.Username,
			password: c.config.Password,
		}, nil
	case AuthPlain:
		return smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.SMTPServer), nil
	case AuthCRAMMD5:
		return smtp.CRAMMD5Auth(c.config.Username, c.config.Password), nil
	default:
		return nil, fmt.Errorf("unsupported auth type: %v", c.config.AuthType)
	}
}

func (c *Client) buildMessage(msg Message) []byte {
	var builder strings.Builder

	// Required headers
	builder.WriteString(fmt.Sprintf("From: %s\r\n", msg.From))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))

	if len(msg.CC) > 0 {
		builder.WriteString(fmt.Sprintf("CC: %s\r\n", strings.Join(msg.CC, ", ")))
	}

	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	builder.WriteString(fmt.Sprintf("Content-Type: %s; charset=UTF-8\r\n", msg.ContentType))
	builder.WriteString("MIME-Version: 1.0\r\n")

	builder.WriteString("\r\n")

	builder.WriteString(msg.Body)

	return []byte(builder.String())
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type loginAuth struct {
	username string
	password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
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
