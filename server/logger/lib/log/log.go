package log

import (
	"fmt"
	"log"
	"os"
)

func LogErrorToFile(ipAddress string, apiKey string, msg string) {
	text := fmt.Sprintf("%s %s :: %s", ipAddress, apiKey, msg)
	LogToFile(text)
}

func LogRequestsToFile(apiKey string, inserted int, totalRequests int) {
	text := fmt.Sprintf("key=%s: %d/%d", apiKey, inserted, totalRequests)
	LogToFile(text)
}

func LogToFile(msg string) {
	f, err := os.OpenFile("./requests.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(msg)
}
