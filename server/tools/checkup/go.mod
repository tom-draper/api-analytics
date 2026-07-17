module github.com/tom-draper/api-analytics/server/tools/checkup

go 1.26.0

require (
	github.com/fatih/color v1.19.0
	github.com/tom-draper/api-analytics/server/database v0.0.0
	github.com/tom-draper/api-analytics/server/email v0.0.0
	github.com/tom-draper/api-analytics/server/tools/config v0.0.0
	github.com/tom-draper/api-analytics/server/tools/monitor v0.0.0
	github.com/tom-draper/api-analytics/server/tools/usage v0.0.0
	golang.org/x/text v0.40.0
)

replace (
	github.com/tom-draper/api-analytics/server/database => ../../database
	github.com/tom-draper/api-analytics/server/email => ../../email
	github.com/tom-draper/api-analytics/server/tools/config => ../config
	github.com/tom-draper/api-analytics/server/tools/monitor => ../monitor
	github.com/tom-draper/api-analytics/server/tools/usage => ../usage
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.23 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)
