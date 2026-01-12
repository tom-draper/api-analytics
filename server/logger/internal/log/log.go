package log

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger zerolog.Logger

func Init() error {
	// Setup log rotation with lumberjack
	fileWriter := &lumberjack.Logger{
		Filename:   "./logger.log",
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	}

	// Write to both file and console
	multi := io.MultiWriter(os.Stdout, fileWriter)

	// Configure zerolog with ConsoleWriter for single-line format
	Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        multi,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    true,
		FormatLevel: func(i any) string {
			return "" // Remove level prefix
		},
	}).With().Timestamp().Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	return nil
}

func Close() error {
	return nil
}

// Info logs an informational message
func Info(msg string) {
	Logger.Info().Msg(msg)
}

// Fatal logs a fatal error and exits
func Fatal(msg string) {
	Logger.Fatal().Msg(msg)
}

// InfoWithFields logs with structured fields
func InfoWithFields(msg string, fields map[string]interface{}) {
	event := Logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// LogErrorToFile logs error messages with context (structured)
func LogErrorToFile(ipAddress string, apiKey string, msg string) {
	Logger.Error().
		Str("ip_address", ipAddress).
		Str("api_key", apiKey).
		Msg(msg)
}

// LogRequestsToFile logs request statistics (structured)
func LogRequestsToFile(apiKey string, inserted int, totalRequests int) {
	Logger.Info().
		Str("api_key", apiKey).
		Int("inserted", inserted).
		Int("total_requests", totalRequests).
		Msg("Requests logged")
}

// LogToFile logs a message
func LogToFile(msg string) {
	Logger.Info().Msg(msg)
}
