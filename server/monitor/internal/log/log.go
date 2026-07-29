package log

import (
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const timeFormat = "2006-01-02 15:04:05"

var out io.Writer = os.Stdout
var errOut io.Writer = io.MultiWriter(os.Stdout, os.Stderr)

func Init() error {
	// Rotate the log file so a long-running daemon cannot fill the disk, matching
	// the api and logger services rather than appending to one unbounded file.
	f := &lumberjack.Logger{
		Filename:   "./monitor.log",
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	}
	out = io.MultiWriter(os.Stdout, f)
	errOut = io.MultiWriter(os.Stdout, os.Stderr, f)
	return nil
}

func now() string {
	return time.Now().UTC().Format(timeFormat)
}

func Info(msg string) {
	fmt.Fprintf(out, "%s %s\n", now(), msg)
}

func Error(msg string) {
	fmt.Fprintf(errOut, "%s ERROR %s\n", now(), msg)
}

func Fatal(msg string) {
	fmt.Fprintf(errOut, "%s ERROR %s\n", now(), msg)
	os.Exit(1)
}
