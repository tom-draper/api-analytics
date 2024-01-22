package usage

import (
	"fmt"

	"github.com/tom-draper/api-analytics/server/database"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type UserCount struct {
	APIKey string
	Count  int
}

var p = message.NewPrinter(language.English)

func (u UserCount) Display() {
	p.Printf("%s: %d\n", u.APIKey, u.Count)
}

func DisplayUserCounts(counts []UserCount) {
	for _, c := range counts {
		c.Display()
	}
}

func TableSize(table string) (string, error) {
	conn := database.NewConnection()
	defer conn.Close()

	var size string
	query := fmt.Sprintf("SELECT pg_size_pretty(pg_total_relation_size('%s'));", table)
	err := conn.QueryRow(query).Scan(&size)
	if err != nil {
		return "", err
	}

	return size, err
}

func TableColumnSize(table string, column string) (string, error) {
	conn := database.NewConnection()
	defer conn.Close()

	var size struct {
		totalSize   string
		averageSize string
		percentage  float64
	}
	query := fmt.Sprintf("SELECT pg_size_pretty(sum(pg_column_size(%s))) AS total_size, pg_size_pretty(avg(pg_column_size(%s))) AS average_size, sum(pg_column_size(%s)) * 100.0 / pg_total_relation_size('%s') AS percentage FROM %s;", column, column, column, table, table)
	err := conn.QueryRow(query).Scan(&size.totalSize, &size.averageSize, &size.percentage)
	if err != nil {
		return "", err
	}

	return size.totalSize, err
}

func DatabaseConnections() (int, error) {
	conn := database.NewConnection()
	defer conn.Close()

	var connections int
	query := "SELECT count(*) from pg_stat_activity;"
	err := conn.Query(query).Scan(&connections)
	if err != nil {
		return 0, err
	}

	return connections, err
}
