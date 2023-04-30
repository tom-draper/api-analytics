package values

import (
	"fmt"
	"github.com/tom-draper/api-analytics/server/database"
)

type StringCount struct {
	Value string
	Count string
}

func columnStringValuesCount(column string) ([]StringCount, error) {
	db := database.OpenDBConnection()

	query := fmt.Sprintf("SELECT %s, COUNT(*) AS count FROM requests GROUP BY user_agent ORDER BY count DESC;", column)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var count []StringCount
	for rows.Next() {
		var userAgent StringCount
		err := rows.Scan(&userAgent.Value, &userAgent.Count)
		if err == nil {
			count = append(count, userAgent)
		}
	}

	return count, nil
}

func UserAgents() ([]StringCount, error) {
	return columnStringValuesCount("user_agent")
}

func IPAddresses() ([]StringCount, error) {
	return columnStringValuesCount("ip_address")
}

func Locations() ([]StringCount, error) {
	return columnStringValuesCount("location")
}

func AvgResponseTime() (float64, error) {
	db := database.OpenDBConnection()

	query := "SELECT AVG(response_time) FROM requests;"
	rows, err := db.Query(query)
	if err != nil {
		return 0.0, err
	}

	var avg float64
	rows.Next()
	err = rows.Scan(&avg)
	if err != nil {
		return 0.0, err
	}

	return avg, nil
}
