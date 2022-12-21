package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

func Ping(client http.Client, domain string, secure bool, method string) (int, time.Duration, error) {
	var url string
	if !secure {
		url = "http://" + domain
	} else {
		url = "https://" + domain
	}

	if method != "GET" && method != "HEAD" {
		return 0, time.Duration(0), errors.New("invalid method")
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, time.Duration(0), err
	}

	// Make request
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, time.Duration(0), err
	}
	elapsed := time.Since(start)

	resp.Body.Close()

	return resp.StatusCode, elapsed, nil
}

func main() {
	dialer := net.Dialer{Timeout: 2 * time.Second}
	var client = http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}

	res, elapsed, err := Ping(client, "example.com", true, "GET")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res, elapsed)
	}
}
