package validator

import (
	"net/http"
	"time"
)

func ValidateLink(urlStr string) string {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Head(urlStr)
	if err != nil {
		return "dead"
	}
	defer resp.Body.Close()
	if resp.StatusCode <= 200 && resp.StatusCode < 400 {
		return "alive"
	}
	return "dead"
}
