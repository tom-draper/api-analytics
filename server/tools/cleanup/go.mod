module github.com/tom-draper/api-analytics/server/tools/cleanup

go 1.25

require (
	github.com/joho/godotenv v1.5.1
	github.com/tom-draper/api-analytics/server/database v0.0.0-20251125173416-63025bb33f9f
	github.com/tom-draper/api-analytics/server/tools/usage v0.0.0-20251125173416-63025bb33f9f
)

require (
	github.com/fatih/color v1.18.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.8.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
)

replace github.com/tom-draper/api-analytics/server/database => ../../database

replace github.com/tom-draper/api-analytics/server/tools/usage => ../usage
