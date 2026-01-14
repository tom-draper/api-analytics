package checkup

// DisplayLastHour displays statistics for the last hour
func (c *Client) DisplayLastHour() {
	c.printBanner("Last Hour")

	users, err := c.usageClient.HourlyUsersCount(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Users:", users)

	requests, err := c.usageClient.HourlyRequestsCount(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Requests:", requests)

	monitors, err := c.usageClient.HourlyMonitorsCount(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Monitors:", monitors)
}

// DisplayLast24Hours displays statistics for the last 24 hours
func (c *Client) DisplayLast24Hours() {
	c.printBanner("Last 24 Hours")

	users, err := c.usageClient.DailyUsersCount(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Users:", users)

	requests, err := c.usageClient.DailyRequestsCount(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Requests:", requests)

	monitors, err := c.usageClient.DailyMonitorsCount(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Monitors:", monitors)
}

// DisplayLastWeek displays statistics for the last week
func (c *Client) DisplayLastWeek() {
	c.printBanner("Last Week")

	users, err := c.usageClient.WeeklyUsersCount(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Users:", users)

	requests, err := c.usageClient.WeeklyRequestsCount(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Requests:", requests)

	monitors, err := c.usageClient.WeeklyMonitorsCount(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Monitors:", monitors)
}

// DisplayTotal displays total statistics
func (c *Client) DisplayTotal() {
	c.printBanner("Total")

	users, err := c.usageClient.UsersCount(c.ctx, "")
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Users:", users)

	requests, err := c.usageClient.RequestsCount(c.ctx, "")
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Requests:", requests)

	monitors, err := c.usageClient.MonitorsCount(c.ctx, "")
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Monitors:", monitors)
}
