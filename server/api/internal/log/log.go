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
		Filename:   "./api.log",
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

func Info(msg string) {
	Logger.Info().Msg(msg)
}

func Fatal(msg string) {
	Logger.Fatal().Msg(msg)
}
