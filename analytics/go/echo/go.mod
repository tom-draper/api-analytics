module github.com/tom-draper/api-analytics/analytics/go/echo

go 1.26.0

require (
	github.com/labstack/echo/v4 v4.16.0
	github.com/tom-draper/api-analytics/analytics/go/core v0.1.0
)

replace github.com/tom-draper/api-analytics/analytics/go/core => ../core

require (
	github.com/labstack/gommon v0.5.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)
