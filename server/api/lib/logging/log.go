package logging

import (
	"log"
	"os"
)

var logger *log.Logger

func Init() {
	f, err := os.OpenFile("./api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	logger = log.New(f, "", log.LstdFlags)
}

func Info(msg string) {
	logger.Println(msg)
}

func Fatal(msg string) {
	logger.Fatalln(msg)
}
