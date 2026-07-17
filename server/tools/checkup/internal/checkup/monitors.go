package checkup

import "fmt"

// DisplayMonitors displays all monitors
func (c *Client) DisplayMonitors() {
	c.printBanner("Monitors")
	monitorsList, err := c.usageClient.TotalMonitors(c.ctx)
	if c.handleError(err) != nil {
		return
	}

	for i, monitor := range monitorsList {
		fmt.Printf("[%d] %s %s %s\n", i, monitor.APIKey, monitor.CreatedAt.Format("2006-01-02 15:04:05"), monitor.URL)
	}
}

// DisplayMonitorsCheckup displays comprehensive monitor information
func (c *Client) DisplayMonitorsCheckup() {
	c.DisplayMonitors()
}
