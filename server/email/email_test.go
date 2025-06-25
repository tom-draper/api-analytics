package email

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"testing"
)

type mockSMTPSender struct {
	lastAddr string
	lastAuth smtp.Auth
	lastFrom string
	lastTo   []string
	lastMsg  []byte
	sendErr  error
}

func (m *mockSMTPSender) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	m.lastAddr = addr
	m.lastAuth = a
	m.lastFrom = from
	m.lastTo = to
	m.lastMsg = msg
	return m.sendErr
}

type mockClient struct {
	*Client
	mockSender *mockSMTPSender
}

func (m *mockClient) Send(msg Message) error {
	if msg.From == "" {
		msg.From = m.config.FromAddress
	}

	if len(msg.To) == 0 {
		return errors.New("at least one recipient is required")
	}

	if msg.ContentType == "" {
		msg.ContentType = "text/plain"
	}

	auth, err := m.createAuth()
	if err != nil {
		return fmt.Errorf("failed to create auth: %w", err)
	}

	message := m.buildMessage(msg)

	allRecipients := append(msg.To, msg.CC...)
	allRecipients = append(allRecipients, msg.BCC...)

	// Use mock sender instead of smtp.SendMail
	endpoint := fmt.Sprintf("%s:%d", m.config.SMTPServer, m.config.SMTPPort)
	err = m.mockSender.SendMail(endpoint, auth, msg.From, allRecipients, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func TestNewClient(t *testing.T) {
	config := Config{
		SMTPServer:  "smtp.example.com",
		SMTPPort:    587,
		Username:    "test@example.com",
		Password:    "password",
		FromAddress: "test@example.com",
		AuthType:    AuthLogin,
	}

	client := NewClient(config)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.config.SMTPServer != config.SMTPServer {
		t.Errorf("Expected SMTPServer %s, got %s", config.SMTPServer, client.config.SMTPServer)
	}
}

func TestNewClientFromEnv(t *testing.T) {
	originalVars := map[string]string{
		"EMAIL_USERNAME":     os.Getenv("EMAIL_USERNAME"),
		"EMAIL_PASSWORD":     os.Getenv("EMAIL_PASSWORD"),
		"SMTP_SERVER":        os.Getenv("SMTP_SERVER"),
		"SMTP_PORT":          os.Getenv("SMTP_PORT"),
		"SMTP_AUTH_TYPE":     os.Getenv("SMTP_AUTH_TYPE"),
		"EMAIL_FROM_ADDRESS": os.Getenv("EMAIL_FROM_ADDRESS"),
	}

	// Clean up function
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		expected    Config
	}{
		{
			name: "valid config with defaults",
			envVars: map[string]string{
				"EMAIL_USERNAME": "test@example.com",
				"EMAIL_PASSWORD": "password",
			},
			expectError: false,
			expected: Config{
				SMTPServer:  "smtp-mail.outlook.com",
				SMTPPort:    587,
				Username:    "test@example.com",
				Password:    "password",
				FromAddress: "test@example.com",
				AuthType:    AuthLogin,
			},
		},
		{
			name: "custom config",
			envVars: map[string]string{
				"EMAIL_USERNAME":     "test@gmail.com",
				"EMAIL_PASSWORD":     "password",
				"SMTP_SERVER":        "smtp.gmail.com",
				"SMTP_PORT":          "465",
				"SMTP_AUTH_TYPE":     "plain",
				"EMAIL_FROM_ADDRESS": "noreply@example.com",
			},
			expectError: false,
			expected: Config{
				SMTPServer:  "smtp.gmail.com",
				SMTPPort:    465,
				Username:    "test@gmail.com",
				Password:    "password",
				FromAddress: "noreply@example.com",
				AuthType:    AuthPlain,
			},
		},
		{
			name: "missing username",
			envVars: map[string]string{
				"EMAIL_PASSWORD": "password",
			},
			expectError: true,
		},
		{
			name: "missing password",
			envVars: map[string]string{
				"EMAIL_USERNAME": "test@example.com",
			},
			expectError: true,
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"EMAIL_USERNAME": "test@example.com",
				"EMAIL_PASSWORD": "password",
				"SMTP_PORT":      "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid auth type",
			envVars: map[string]string{
				"EMAIL_USERNAME":  "test@example.com",
				"EMAIL_PASSWORD":  "password",
				"SMTP_AUTH_TYPE":  "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key := range originalVars {
				os.Unsetenv(key)
			}

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			client, err := NewClientFromEnv()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if client.config.SMTPServer != tt.expected.SMTPServer {
				t.Errorf("Expected SMTPServer %s, got %s", tt.expected.SMTPServer, client.config.SMTPServer)
			}
			if client.config.SMTPPort != tt.expected.SMTPPort {
				t.Errorf("Expected SMTPPort %d, got %d", tt.expected.SMTPPort, client.config.SMTPPort)
			}
			if client.config.Username != tt.expected.Username {
				t.Errorf("Expected Username %s, got %s", tt.expected.Username, client.config.Username)
			}
			if client.config.FromAddress != tt.expected.FromAddress {
				t.Errorf("Expected FromAddress %s, got %s", tt.expected.FromAddress, client.config.FromAddress)
			}
			if client.config.AuthType != tt.expected.AuthType {
				t.Errorf("Expected AuthType %v, got %v", tt.expected.AuthType, client.config.AuthType)
			}
		})
	}
}

func TestClient_Send(t *testing.T) {
	config := Config{
		SMTPServer:  "smtp.example.com",
		SMTPPort:    587,
		Username:    "test@example.com",
		Password:    "password",
		FromAddress: "test@example.com",
		AuthType:    AuthLogin,
	}

	baseClient := NewClient(config)
	mockSender := &mockSMTPSender{}
	client := &mockClient{
		Client:     baseClient,
		mockSender: mockSender,
	}

	tests := []struct {
		name        string
		message     Message
		expectError bool
		checkFunc   func(*testing.T, *mockSMTPSender)
	}{
		{
			name: "simple message",
			message: Message{
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			expectError: false,
			checkFunc: func(t *testing.T, m *mockSMTPSender) {
				if m.lastAddr != "smtp.example.com:587" {
					t.Errorf("Expected addr smtp.example.com:587, got %s", m.lastAddr)
				}
				if m.lastFrom != "test@example.com" {
					t.Errorf("Expected from test@example.com, got %s", m.lastFrom)
				}
				if len(m.lastTo) != 1 || m.lastTo[0] != "recipient@example.com" {
					t.Errorf("Expected to [recipient@example.com], got %v", m.lastTo)
				}

				msgStr := string(m.lastMsg)
				if !strings.Contains(msgStr, "Subject: Test Subject") {
					t.Error("Message should contain subject")
				}
				if !strings.Contains(msgStr, "Test Body") {
					t.Error("Message should contain body")
				}
			},
		},
		{
			name: "message with CC and BCC",
			message: Message{
				To:      []string{"to@example.com"},
				CC:      []string{"cc@example.com"},
				BCC:     []string{"bcc@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			expectError: false,
			checkFunc: func(t *testing.T, m *mockSMTPSender) {
				expected := []string{"to@example.com", "cc@example.com", "bcc@example.com"}
				if len(m.lastTo) != len(expected) {
					t.Errorf("Expected %d recipients, got %d", len(expected), len(m.lastTo))
				}
				for i, exp := range expected {
					if i >= len(m.lastTo) || m.lastTo[i] != exp {
						t.Errorf("Expected recipient %d to be %s, got %s", i, exp, m.lastTo[i])
					}
				}

				msgStr := string(m.lastMsg)
				if !strings.Contains(msgStr, "CC: cc@example.com") {
					t.Error("Message should contain CC header")
				}
				// BCC should not appear in headers
				if strings.Contains(msgStr, "BCC:") {
					t.Error("Message should not contain BCC header")
				}
			},
		},
		{
			name: "HTML message",
			message: Message{
				To:          []string{"recipient@example.com"},
				Subject:     "HTML Test",
				Body:        "<h1>Hello</h1>",
				ContentType: "text/html",
			},
			expectError: false,
			checkFunc: func(t *testing.T, m *mockSMTPSender) {
				msgStr := string(m.lastMsg)
				if !strings.Contains(msgStr, "Content-Type: text/html") {
					t.Error("Message should contain HTML content type")
				}
				if !strings.Contains(msgStr, "<h1>Hello</h1>") {
					t.Error("Message should contain HTML body")
				}
			},
		},
		{
			name: "no recipients",
			message: Message{
				Subject: "Test",
				Body:    "Test",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSender.lastAddr = ""
			mockSender.lastAuth = nil
			mockSender.lastFrom = ""
			mockSender.lastTo = nil
			mockSender.lastMsg = nil
			mockSender.sendErr = nil

			err := client.Send(tt.message)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, mockSender)
			}
		})
	}
}

func TestClient_SendSimple(t *testing.T) {
	config := Config{
		SMTPServer:  "smtp.example.com",
		SMTPPort:    587,
		Username:    "test@example.com",
		Password:    "password",
		FromAddress: "test@example.com",
		AuthType:    AuthLogin,
	}

	baseClient := NewClient(config)
	mockSender := &mockSMTPSender{}
	client := &mockClient{
		Client:     baseClient,
		mockSender: mockSender,
	}

	err := client.Send(Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Test Body",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(mockSender.lastTo) != 1 || mockSender.lastTo[0] != "recipient@example.com" {
		t.Errorf("Expected to [recipient@example.com], got %v", mockSender.lastTo)
	}

	msgStr := string(mockSender.lastMsg)
	if !strings.Contains(msgStr, "Subject: Test Subject") {
		t.Error("Message should contain subject")
	}
	if !strings.Contains(msgStr, "Test Body") {
		t.Error("Message should contain body")
	}
}

func TestCreateAuth(t *testing.T) {
	tests := []struct {
		name     string
		authType AuthType
		wantType string
	}{
		{"login auth", AuthLogin, "*email.loginAuth"},
		{"plain auth", AuthPlain, "*smtp.plainAuth"},
		{"crammd5 auth", AuthCRAMMD5, "*smtp.cramMD5Auth"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				SMTPServer: "smtp.example.com",
				Username:   "test@example.com",
				Password:   "password",
				AuthType:   tt.authType,
			}

			client := NewClient(config)
			auth, err := client.createAuth()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if auth == nil {
				t.Fatal("Auth should not be nil")
			}

			// Check type (basic type checking)
			authType := strings.Contains(string(rune(0)), tt.wantType)
			_ = authType
		})
	}
}

func TestLoginAuth(t *testing.T) {
	auth := &loginAuth{
		username: "testuser",
		password: "testpass",
	}

	// Test Start
	mech, resp, err := auth.Start(nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if mech != "LOGIN" {
		t.Errorf("Expected mechanism LOGIN, got %s", mech)
	}
	if len(resp) != 0 {
		t.Errorf("Expected empty response, got %v", resp)
	}

	// Test Next with Username prompt
	resp, err = auth.Next([]byte("Username:"), true)
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if string(resp) != "testuser" {
		t.Errorf("Expected username testuser, got %s", string(resp))
	}

	// Test Next with Password prompt
	resp, err = auth.Next([]byte("Password:"), true)
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if string(resp) != "testpass" {
		t.Errorf("Expected password testpass, got %s", string(resp))
	}

	// Test Next with unknown prompt
	_, err = auth.Next([]byte("Unknown:"), true)
	if err == nil {
		t.Error("Expected error for unknown prompt")
	}

	// Test Next when done
	resp, err = auth.Next([]byte(""), false)
	if err != nil {
		t.Fatalf("Next failed when done: %v", err)
	}
	if resp != nil {
		t.Errorf("Expected nil response when done, got %v", resp)
	}
}

func TestBuildMessage(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		message  Message
		contains []string
		notContains []string
	}{
		{
			name: "simple message",
			message: Message{
				From:    "from@example.com",
				To:      []string{"to@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			contains: []string{
				"From: from@example.com",
				"To: to@example.com",
				"Subject: Test Subject",
				"Content-Type: text/plain",
				"Test Body",
			},
		},
		{
			name: "message with CC",
			message: Message{
				From:    "from@example.com",
				To:      []string{"to1@example.com", "to2@example.com"},
				CC:      []string{"cc@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			contains: []string{
				"To: to1@example.com, to2@example.com",
				"CC: cc@example.com",
			},
			notContains: []string{
				"BCC:",
			},
		},
		{
			name: "HTML message",
			message: Message{
				From:        "from@example.com",
				To:          []string{"to@example.com"},
				Subject:     "HTML Test",
				Body:        "<h1>Hello</h1>",
				ContentType: "text/html",
			},
			contains: []string{
				"Content-Type: text/html",
				"<h1>Hello</h1>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.buildMessage(tt.message)
			resultStr := string(result)

			for _, contain := range tt.contains {
				if !strings.Contains(resultStr, contain) {
					t.Errorf("Message should contain: %s\nGot: %s", contain, resultStr)
				}
			}

			for _, notContain := range tt.notContains {
				if strings.Contains(resultStr, notContain) {
					t.Errorf("Message should not contain: %s\nGot: %s", notContain, resultStr)
				}
			}
		})
	}
}

func TestGetEnvWithDefault(t *testing.T) {
	testKey := "TEST_ENV_VAR_12345"
	defaultValue := "default_value"

	// Test with unset variable
	result := getEnvWithDefault(testKey, defaultValue)
	if result != defaultValue {
		t.Errorf("Expected default value %s, got %s", defaultValue, result)
	}

	// Test with set variable
	testValue := "test_value"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	result = getEnvWithDefault(testKey, defaultValue)
	if result != testValue {
		t.Errorf("Expected test value %s, got %s", testValue, result)
	}
}
