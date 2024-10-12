package usage

import (
	"context"
	"fmt"

	"github.com/tom-draper/api-analytics/server/database"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type UserCount struct {
	APIKey string
	Count  int
}

const (
	hourly  = "1 hour"
	daily   = "24 hours"
	weekly  = "7 days"
	monthly = "30 days"
)

var p = message.NewPrinter(language.English)

func (u UserCount) Display() {
	p.Printf("%s: %d\n", u.APIKey, u.Count)
}

func DisplayUserCounts(counts []UserCount) {
	for _, c := range counts {
		c.Display()
	}
}

func TableSize(ctx context.Context, table string) (string, error) {
	conn := database.NewConnection()
	defer conn.Close(ctx)

	var size string
	query := fmt.Sprintf("SELECT pg_size_pretty(pg_total_relation_size('%s'));", table)
	err := conn.QueryRow(ctx, query).Scan(&size)
	if err != nil {
		return "", err
	}

	return size, nil
}

func TableColumnSize(ctx context.Context, table, column string) (string, string, float64, error) {
	conn := database.NewConnection()
	defer conn.Close(ctx)

	var columnSizeInfo struct {
		TotalSize   string
		AverageSize string
		Percentage  float64
	}

	query := fmt.Sprintf(`
		SELECT pg_size_pretty(sum(pg_column_size(%s))) AS total_size,
		       pg_size_pretty(avg(pg_column_size(%s))) AS average_size,
		       sum(pg_column_size(%s)) * 100.0 / pg_total_relation_size('%s') AS percentage
		FROM %s;`, column, column, column, table, table)

	err := conn.QueryRow(ctx, query).Scan(&columnSizeInfo.TotalSize, &columnSizeInfo.AverageSize, &columnSizeInfo.Percentage)
	if err != nil {
		return "", "", 0, err
	}

	return columnSizeInfo.TotalSize, columnSizeInfo.AverageSize, columnSizeInfo.Percentage, nil
}

func DatabaseConnections(ctx context.Context) (int, error) {
	conn := database.NewConnection()
	defer conn.Close(ctx)

	var connections int
	query := "SELECT count(*) FROM pg_stat_activity;"
	err := conn.QueryRow(ctx, query).Scan(&connections)
	if err != nil {
		return 0, err
	}

	return connections, nil
}
