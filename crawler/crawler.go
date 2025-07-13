package crawler

import (
	"crawl4dead/models"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

func FetchLinks(urlStr string) ([]models.Link, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	var links []models.Link
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					absURL := resolveURL(urlStr, a.Val)
					links = append(links, models.Link{URL: absURL, Source: urlStr})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return links, nil
}

func resolveURL(base, href string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return href
	}
	parsedURL, err := url.Parse(href)
	if err != nil {
		return href
	}
	return baseURL.ResolveReference(parsedURL).String()
}

func IsExternal(urlStr, baseDomain string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return true // Treat invalid URLs as external
	}
	return parsedURL.Host != "" && parsedURL.Host != baseDomain
}
