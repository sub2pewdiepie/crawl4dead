package validator

import (
	"net/http"
	"time"
)

func ValidateLink(urlStr string) (string, int32) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Head(urlStr)
	if err != nil {
		return "dead", 0
	}
	defer resp.Body.Close()
	if resp.StatusCode <= 200 && resp.StatusCode < 400 {
		return "alive", int32(resp.StatusCode)
	}
	return "dead", int32(resp.StatusCode)
}
