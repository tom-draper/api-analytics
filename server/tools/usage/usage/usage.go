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

func DisplayUserCounts(counts []UserCount) {
	p := message.NewPrinter(language.English)
	for _, c := range counts {
		p.Printf("%s: %d\n", c.APIKey, c.Count)
	}
}

func DatabaseSize() (string, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT pg_size_pretty(pg_total_relation_size('requests'));"
	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}

	var size string
	rows.Next()
	err = rows.Scan(&size)
	return size, err
}

func DatabaseColumnSize(column string) (string, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := fmt.Sprintf("SELECT pg_size_pretty(sum(pg_column_size(%s))) AS total_size, pg_size_pretty(avg(pg_column_size(%s))) AS average_size, sum(pg_column_size(%s)) * 100.0 / pg_total_relation_size('requests') AS percentage FROM requests;", column, column, column)
	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}

	var size struct {
		totalSize   string
		averageSize float64
		percentage  float64
	}
	rows.Next()
	err = rows.Scan(&size.totalSize, &size.averageSize, &size.percentage)
	return size.totalSize, err
}

func DatabaseConnections() (int, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT count(*) from pg_stat_activity;"
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}

	var connections int
	rows.Next()
	err = rows.Scan(&connections)
	return connections, err
}
