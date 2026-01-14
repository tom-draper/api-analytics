package monitor

import (
	"fmt"
	"io"
	"net/http"
)

// makeGetRequest performs a GET request with optional API key authentication
func (c *Client) makeGetRequest(url string, apiKey string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if apiKey != "" {
		request.Header = http.Header{
			"X-AUTH-TOKEN": {apiKey},
		}
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
