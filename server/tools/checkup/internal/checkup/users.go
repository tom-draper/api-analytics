package checkup

import "github.com/tom-draper/api-analytics/server/tools/usage"

// DisplayTopUsers displays the top users by request count
func (c *Client) DisplayTopUsers() {
	c.printBanner("Top Users")
	topUsers, err := c.usageClient.TopUsers(c.ctx, 10)
	if c.handleError(err) != nil {
		return
	}
	usage.DisplayUsers(topUsers)
}

// DisplayUnusedUsers displays users who have never made requests
func (c *Client) DisplayUnusedUsers() {
	c.printBanner("Unused Users")
	unusedUsers, err := c.usageClient.UnusedUsers(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	usage.DisplayUserTimes(unusedUsers)
}

// DisplayUsersSinceLastRequest displays time since users' last requests
func (c *Client) DisplayUsersSinceLastRequest() {
	c.printBanner("Users Since Last Request")
	sinceLastRequestUsers, err := c.usageClient.SinceLastRequestUsers(c.ctx)
	if c.handleError(err) != nil {
		return
	}
	usage.DisplayUserTimes(sinceLastRequestUsers)
}

// DisplayUsersCheckup displays comprehensive user information
func (c *Client) DisplayUsersCheckup() {
	c.DisplayTopUsers()
	c.DisplayUnusedUsers()
	c.DisplayUsersSinceLastRequest()
}
