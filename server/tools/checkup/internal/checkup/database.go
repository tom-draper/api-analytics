package checkup

// DisplayDatabaseStats displays database statistics
func (c *Client) DisplayDatabaseStats() {
	c.printBanner("Database")

	connections, err := c.usageClient.DatabaseConnections(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Active database connections:", connections)

	size, err := c.usageClient.TableSize(c.ctx, "requests")
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Database size:", size)
}

// DisplayDatabaseTableStats displays detailed database table statistics
func (c *Client) DisplayDatabaseTableStats() {
	c.printBanner("Requests Fields")
	columnSize, err := c.usageClient.RequestsColumnSize(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	columnSize.Display()
}

// DisplayDatabaseCheckup displays comprehensive database information
func (c *Client) DisplayDatabaseCheckup() {
	c.DisplayDatabaseStats()

	totalRequests, err := c.usageClient.RequestsCount(c.ctx, "")
	if c.handleError(err) != nil {
		return
	}
	c.printer.Println("Requests:", totalRequests)

	c.DisplayDatabaseTableStats()
}
