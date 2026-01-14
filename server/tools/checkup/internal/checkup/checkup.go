package checkup

// DisplayCheckup performs and displays a full system checkup
func (c *Client) DisplayCheckup() {
	c.DisplayServicesTest()
	c.DisplayAPITest()
	c.DisplayLoggerTest()
	c.DisplayDatabaseStats()
	c.DisplayLastHour()
	c.DisplayLast24Hours()
	c.DisplayLastWeek()
	c.DisplayTotal()
}
