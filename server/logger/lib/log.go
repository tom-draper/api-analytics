package lib

import (
	"fmt"
	"log"
	"os"
)

type RequestErrors struct {
	method    int
	framework int
	userAgent int
	hostname  int
	path      int
}

func LogErrorToFile(ipAddress string, apiKey string, msg string) {
	text := fmt.Sprintf("%s %s :: %s", ipAddress, apiKey, msg)
	LogToFile(text)
}

func LogRequestsToFile(ipAddress string, apiKey string, inserted int, totalRequests int, requestErrors RequestErrors) {
	text := fmt.Sprintf("%s %s :: inserted=%d totalRequest=%d :: %d %d %d %d %d", ipAddress, apiKey, inserted, totalRequests, requestErrors.method, requestErrors.framework, requestErrors.userAgent, requestErrors.hostname, requestErrors.path)
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
