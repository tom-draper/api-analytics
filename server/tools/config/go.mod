module github.com/tom-draper/api-analytics/server/tools/config

go 1.26.0

require (
	github.com/joho/godotenv v1.5.1
	github.com/tom-draper/api-analytics/server/database v0.0.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace github.com/tom-draper/api-analytics/server/database => ../../database
