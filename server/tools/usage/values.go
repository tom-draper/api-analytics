package main

import (
	"github.com/tom-draper/api-analytics/server/database"
)

type StringCount struct {
	Value string
	Count string
}

func UserAgents() ([]StringCount, error) {
	db := database.OpenDBConnection()

	query := "SELECT user_agent, COUNT(*) AS count FROM requests GROUP BY user_agent ORDER BY count DESC;"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
		panic(err)
	}

	var userAgents []StringCount
	for rows.Next() {
		var userAgent StringCount
		err := rows.Scan(&userAgent.Value, &userAgent.Count)
		if err == nil {
			userAgents = append(userAgents, userAgent)
		}
	}

	return userAgents, nil
}

func IPAddresses() ([]StringCount, error) {
	db := database.OpenDBConnection()

	query := "SELECT ip_address, COUNT(*) AS count FROM requests GROUP BY ip_address ORDER BY count DESC;"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var ipAddresses []StringCount
	for rows.Next() {
		var ipAddress StringCount
		err := rows.Scan(&ipAddress.Value, &ipAddress.Count)
		if err == nil {
			ipAddresses = append(ipAddresses, ipAddress)
		}
	}

	return ipAddresses, nil
}
