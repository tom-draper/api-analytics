module github.com/tom-draper/api-analytics/server/tools/usage

go 1.26.0

require (
	github.com/fatih/color v1.19.0
	github.com/joho/godotenv v1.5.1
	github.com/tom-draper/api-analytics/server/database v0.0.0
	golang.org/x/text v0.37.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.9.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	golang.org/x/sys v0.44.0 // indirect
)

replace github.com/tom-draper/api-analytics/server/database => ../../database
