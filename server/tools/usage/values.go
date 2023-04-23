package main

import (
	"fmt"

	"github.com/tom-draper/api-analytics/server/database"
)

type UserAgentCount struct {
	UserAgent string
	Count     string
}

func UserAgents() {
	db := database.OpenDBConnection()

	query := "SELECT user_agent, COUNT(*) AS count FROM requests GROUP BY user_agent ORDER BY count DESC;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var userAgents []UserAgentCount
	for rows.Next() {
		var userAgent UserAgentCount
		err := rows.Scan(&userAgent.UserAgent, &userAgent.Count)
		if err == nil {
			userAgents = append(userAgents, userAgent)
		}
	}

	fmt.Println(userAgents)
}

func IPAddresses() {
	db := database.OpenDBConnection()

	query := "SELECT ip_address, COUNT(*) AS count FROM requests GROUP BY ip_address ORDER BY count DESC;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var userAgents []UserAgentCount
	for rows.Next() {
		var userAgent UserAgentCount
		err := rows.Scan(&userAgent.UserAgent, &userAgent.Count)
		if err == nil {
			userAgents = append(userAgents, userAgent)
		}
	}

	fmt.Println(userAgents)
}
