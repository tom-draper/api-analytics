package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

var out io.Writer = os.Stdout
var errOut io.Writer = io.MultiWriter(os.Stdout, os.Stderr)

func Init() error {
	f, err := os.OpenFile("./monitor.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
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
