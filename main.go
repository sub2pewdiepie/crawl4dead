package main

import (
	"crawl4dead/crawler"
	"crawl4dead/models"
	"crawl4dead/tui"
	"crawl4dead/validator"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type Crawler struct {
	visited    map[string]bool
	results    []models.Result
	mu         sync.Mutex
	baseDomain string
	maxDepth   int
	workers    int
}

func main() {
	var startURL string
	var maxDepth int
	var workers int

	rootCmd := &cobra.Command{
		Use:   "crawl4dead [flags]",
		Short: "Finds dead links, outbound links and some filters",
		Run: func(cmd *cobra.Command, args []string) {
			u, err := url.Parse(startURL)
			if err != nil {
				fmt.Printf("Invalid URL: %v\n", err)
				os.Exit(1)
			}

			crawler := &Crawler{
				visited:    make(map[string]bool),
				results:    []models.Result{},
				baseDomain: u.Host,
				maxDepth:   maxDepth,
				workers:    workers,
			}
			start := time.Now()
			crawler.Crawl(startURL, 0)
			finished_in := time.Since(start)
			tui.RunTUI(crawler.results, finished_in)
		},
	}
	rootCmd.Flags().StringVarP(&startURL, "url", "u", "", "Starting URL(*)")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 5, "Maximum crawl depth")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 10, "Number of concurrent workers")
	rootCmd.MarkFlagRequired("url")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (c *Crawler) Crawl(urlStr string, depth int) {
	if depth >= c.maxDepth {
		return
	}
	c.mu.Lock()
	if c.visited[urlStr] {
		c.mu.Unlock()
	}
	c.visited[urlStr] = true
	c.mu.Unlock()
	links, err := crawler.FetchLinks(urlStr)
	if err != nil {
		c.mu.Lock()
		c.results = append(c.results, models.Result{Link: models.Link{URL: urlStr, Source: "initial"}, Status: "dead"})
		c.mu.Unlock()
		return
	}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, c.workers)
	resultChan := make(chan models.Result)
	go func() {
		for res := range resultChan {
			c.mu.Lock()
			c.results = append(c.results, res)
			c.mu.Unlock()
		}
	}()
	for _, link := range links {
		if crawler.IsExternal(link.URL, c.baseDomain) {
			c.mu.Lock()
			c.results = append(c.results, models.Result{Link: link, Status: "outbound"})
			c.mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(link models.Link) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			status := validator.ValidateLink(link.URL)
			resultChan <- models.Result{Link: link, Status: status}
			if status == "alive" && !c.visited[link.URL] {
				c.Crawl(link.URL, depth+1)
			}
		}(link)
	}
	wg.Wait()
	close(resultChan)
}
