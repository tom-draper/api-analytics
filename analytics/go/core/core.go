package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Data struct {
	APIKey       string `json:"api_key"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	Path         string `json:"path"`
	UserAgent    string `json:"user_agent"`
	Method       string `json:"method"`
	ResponseTime int64  `json:"response_time"`
	Status       int    `json:"status"`
	Framework    string `json:"framework"`
}

func LogRequest(data Data) {
	reqBody, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	http.Post("https://api-analytics-server.vercel.app/api/log-request", "application/json", bytes.NewBuffer(reqBody))
}
