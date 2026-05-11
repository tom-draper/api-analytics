package log

import (
	"fmt"
	"os"
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

func now() string {
	return time.Now().UTC().Format(timeFormat)
}

func Info(msg string) {
	fmt.Printf("%s %s\n", now(), msg)
}

func Error(msg string) {
	fmt.Fprintf(os.Stderr, "%s ERR %s\n", now(), msg)
}

func Fatal(msg string) {
	fmt.Fprintf(os.Stderr, "%s ERR %s\n", now(), msg)
	os.Exit(1)
}
