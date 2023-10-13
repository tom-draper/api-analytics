# Monitor

A program to ping all user-registered URLs for monitoring and record and store status codes and response times. Scheduled to run every 30 minutes as a cron job.

## Development

```bash
go run main.go
```

## Production

```bash
go build -o bin/main main.go
./bin/main
```
