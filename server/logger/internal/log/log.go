package log

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger    zerolog.Logger
	logWriter *lumberjack.Logger
)

func Init() error {
	logWriter = &lumberjack.Logger{
		Filename:   "./logger.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	multi := io.MultiWriter(os.Stdout, logWriter)

	Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        multi,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    true,
		FormatLevel: func(i any) string {
			switch v := i.(type) {
			case string:
				if v == "error" || v == "fatal" {
					return "ERROR"
				}
			case zerolog.Level:
				if v >= zerolog.ErrorLevel {
					return "ERROR"
				}
			}
			return ""
		},
	}).With().Timestamp().Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	return nil
}

func Close() error {
	if logWriter != nil {
		return logWriter.Close()
	}
	return nil
}

func Info(msg string) {
	Logger.Info().Msg(msg)
}

func Error(msg string) {
	Logger.Error().Msg(msg)
}

func Fatal(msg string) {
	Logger.Fatal().Msg(msg)
}

// LogClientError logs a client-side rejection with IP and API key context at Info level.
func LogClientError(ipAddress string, apiKey string, msg string) {
	if apiKey != "" {
		Logger.Info().Msg(fmt.Sprintf("ip=%s key=%s: %s", ipAddress, apiKey, msg))
	} else {
		Logger.Info().Msg(fmt.Sprintf("ip=%s: %s", ipAddress, msg))
	}
}

// LogRequestsInserted logs the outcome of a batch insert.
func LogRequestsInserted(apiKey string, inserted int, total int) {
	Logger.Info().Msg(fmt.Sprintf("key=%s: requests logged [%d/%d]", apiKey, inserted, total))
}
