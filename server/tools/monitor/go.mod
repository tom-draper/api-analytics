module github.com/tom-draper/api-analytics/server/tools/monitor

go 1.26.0

require (
	github.com/tom-draper/api-analytics/server/database v0.0.0
	github.com/tom-draper/api-analytics/server/email v0.0.0
	github.com/tom-draper/api-analytics/server/tools/config v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace (
	github.com/tom-draper/api-analytics/server/database => ../../database
	github.com/tom-draper/api-analytics/server/email => ../../email
	github.com/tom-draper/api-analytics/server/tools/config => ../config
)
