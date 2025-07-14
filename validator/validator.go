package validator

import (
	"net/http"
	"time"
)

func OldValidateLink(urlStr string) (string, int32) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Head(urlStr)
	if err != nil {
		return "dead", 0
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return "alive", int32(resp.StatusCode)
	}
	return "dead", int32(resp.StatusCode)
}

func ValidateLink(urlStr string) (string, int32) {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// limit to prevent infinite loops
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "dead", 0
	}
	req.Header.Set("User-Agent", "crawl4dead-bot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		// network failure, timeout, DNS issues
		return "dead", 0
	}
	defer resp.Body.Close()

	code := int32(resp.StatusCode)

	switch {
	case code >= 200 && code < 300:
		return "alive", code
	case code >= 300 && code < 400:
		return "redirect", code
	case code >= 400 && code < 500:
		return "forbidden", code
	case code >= 500:
		return "server-error", code
	default:
		return "dead", code
	}
}
