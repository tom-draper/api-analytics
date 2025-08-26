package log

import (
	"log"
	"os"
	"strings"
)

func LogToFile(msg string) {
	logFile := os.Getenv("LOG_OUTPUT_FILE")
	if logFile == "" {
		logFile = "./api.log"
	}
	
	if strings.ToLower(logFile) == "stdout" {
		log.Println(msg)
	} else {
		f, err := os.OpenFile("./api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
	
		log.SetOutput(f)
		log.Println(msg)
	}
}
