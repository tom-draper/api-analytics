package checkup

import (
	"fmt"
	"log"

	"github.com/tom-draper/api-analytics/server/email"
)

// buildEmailBody creates the email body for daily checkup
func (c *Client) buildEmailBody() (string, error) {
	usersList, err := c.usageClient.DailyUsers(c.ctx)
	if err != nil {
		return "", err
	}

	requestsList, err := c.usageClient.DailyUserRequests(c.ctx)
	if err != nil {
		return "", err
	}

	monitorsList, err := c.usageClient.DailyUserMonitors(c.ctx)
	if err != nil {
		return "", err
	}

	size, err := c.usageClient.TableSize(c.ctx, "requests")
	if err != nil {
		return "", err
	}

	connections, err := c.usageClient.DatabaseConnections(c.ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d new users\n%d requests\n%d monitors\nDatabase size: %s\nActive database connections: %d",
		len(usersList), len(requestsList), len(monitorsList), size, connections), nil
}

// SendEmailCheckup sends a daily checkup email
func (c *Client) SendEmailCheckup() error {
	if !c.cfg.HasEmailConfig() {
		return fmt.Errorf("EMAIL_ADDRESS environment variable not set")
	}

	sender, err := email.NewFromEnv()
	if err != nil {
		log.Printf("Error creating email sender: %v\n", err)
		return err
	}
	if sender == nil {
		return fmt.Errorf("email is disabled: set EMAIL_PROVIDER to send checkup emails")
	}

	body, err := c.buildEmailBody()
	if err != nil {
		log.Printf("Error building email body: %v\n", err)
		return err
	}

	return sender.Send(email.Message{
		To:      []string{c.cfg.EmailAddress},
		From:    "",
		Subject: "API Analytics",
		Body:    body,
	})
}
