package email

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// httpClient is shared by the HTTP-API providers. A timeout keeps a hung
// provider from blocking a send indefinitely.
var httpClient = &http.Client{Timeout: 15 * time.Second}

// doRequest sends req and returns an error unless the response status is 2xx,
// including a trimmed response body to aid debugging.
func doRequest(req *http.Request, provider string) error {
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s request failed: %w", provider, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("%s returned %s: %s", provider, resp.Status, string(body))
}
