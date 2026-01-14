package check

import (
	"fmt"

	"github.com/tom-draper/api-analytics/server/tools/usage"
)

func EmailBody(usersList []usage.UserRow, requests []usage.UserCount, monitors []usage.UserCount, size string, connections int) string {
	return fmt.Sprintf("%d new users\n%d requests\n%d monitors\nDatabase size: %s\nActive database connections: %d",
		len(usersList), len(requests), len(monitors), size, connections)
}
