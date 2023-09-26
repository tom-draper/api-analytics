package usage

import (
	"github.com/tom-draper/api-analytics/server/database"
)

type FrameworkCount struct {
	Framework int
	Count     int
}

func (f FrameworkCount) Display() {
	p.Printf("Framework %d: %d\n", f.Framework, f.Count)
}

func TopFrameworks() ([]FrameworkCount, error) {
	db := database.OpenDBConnection()

	query := "SELECT framework, COUNT(*) AS total_requests FROM requests GROUP BY framework ORDER BY total_requests DESC;"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var frameworks []FrameworkCount
	for rows.Next() {
		framework := new(FrameworkCount)
		err := rows.Scan(&framework.Framework, &framework.Count)
		if err == nil {
			frameworks = append(frameworks, *framework)
		}
	}
	db.Close()

	return frameworks, nil
}
