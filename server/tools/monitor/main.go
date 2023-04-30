package main

import (
	"log"
	"net/http"
)

func checkNewUser() bool {
	// database.DeleteUser()
	resp, err := http.Get("https://apianalytics-server.com/api/generate-api-key")
	if err != nil {
		panic(err)
	}
}

func main() {
	checkNewUser()
}
