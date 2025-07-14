package main

import (
	"crawl4dead/crawler"
	"crawl4dead/models"
	"crawl4dead/validator"
	"fmt"
	"net/url"
	"sync"
)

type ProducerConsumer struct {
	results     []models.Result
	visitedMu   sync.RWMutex
	baseDomain  string
	wgConsumers sync.WaitGroup
	wgProducers sync.WaitGroup
	workers     int
	maxDepth    int
	mu          sync.Mutex
	taskChan    chan models.Link
	resultChan  chan models.Result
	visited     map[string]bool
	sem         chan struct{}
}

func NewProducerConsumer(baseDomain string, workers, maxDepth int) *ProducerConsumer {
	return &ProducerConsumer{
		results:    []models.Result{},
		workers:    workers,
		maxDepth:   maxDepth,
		baseDomain: baseDomain,
		taskChan:   make(chan models.Link, 100),
		resultChan: make(chan models.Result, 100),
		sem:        make(chan struct{}, workers),
		visited:    make(map[string]bool),
	}
}

func (pc *ProducerConsumer) Crawl(startURL string) {
	normalizedURL, err := normalizeURL(startURL)
	if err != nil {
		pc.resultChan <- models.Result{
			Link:       models.Link{URL: startURL, Source: "initial", Depth: 0},
			Status:     "dead",
			StatusCode: 0,
		}
	}

	go pc.collectResults()
	for i := 0; i < pc.workers; i++ {
		pc.wgConsumers.Add(1)
		go pc.consumer()
	}
	pc.wgProducers.Add(1)
	go pc.producer(normalizedURL, 0)
	pc.wgProducers.Wait()
	close(pc.taskChan)
	fmt.Println("producers finished")
	pc.wgConsumers.Wait()
	close(pc.resultChan)

}

func (pc *ProducerConsumer) producer(url string, depth int) {
	defer pc.wgProducers.Done()

	//visited URLs
	pc.visitedMu.Lock()
	if pc.visited[url] {
		pc.visitedMu.Unlock()
		return
	}
	pc.visited[url] = true
	pc.visitedMu.Unlock()
	//crawl depth check
	if depth > pc.maxDepth {
		fmt.Println("reached max depth")
		return
	}
	//fetch links
	links, err := crawler.FetchLinks(url, depth)
	fmt.Printf("Crawling %s at depth %d\n", url, depth)
	if err != nil {
		fmt.Printf("Error fetching links for %s: %v\n", url, err)
		pc.resultChan <- models.Result{
			Link:       models.Link{URL: url, Source: "initial", Depth: depth + 1},
			Status:     "dead",
			StatusCode: 0,
		}
		// return
	}
	//form chanel for consumers
	for _, link := range links {
		link.Depth = depth + 1
		select {
		case pc.taskChan <- link:
			//success
		default:
			//stall?
		}
	}
	//crawl for new URLs
	for _, link := range links {
		if link.Depth <= pc.maxDepth && !crawler.IsExternal(link.URL, pc.baseDomain) {
			normalizedLinkURL, err := normalizeURL(link.URL)
			if err != nil {
				continue
			}
			pc.visitedMu.RLock()
			if !pc.visited[normalizedLinkURL] {
				pc.visitedMu.RUnlock()
				pc.wgProducers.Add(1)
				go pc.producer(normalizedLinkURL, link.Depth)
			} else {
				pc.visitedMu.RUnlock()
			}
		}
	}

}
func (pc *ProducerConsumer) consumer() {
	defer pc.wgConsumers.Done()
	for link := range pc.taskChan {
		pc.sem <- struct{}{}
		fmt.Println("consumer doing smth")
		if external := crawler.IsExternal(link.URL, pc.baseDomain); external {
			pc.resultChan <- models.Result{Link: link, Status: "outbound", StatusCode: 0}
		} else {
			status, statusCode := validator.ValidateLink(link.URL)
			pc.resultChan <- models.Result{Link: link, Status: status, StatusCode: statusCode}
		}
		fmt.Println("consumer did smth")
		<-pc.sem
	}
}

func (pc *ProducerConsumer) collectResults() {
	for res := range pc.resultChan {
		pc.mu.Lock()
		pc.results = append(pc.results, res)
		pc.mu.Unlock()
	}
}
func (pc *ProducerConsumer) GetResults() []models.Result {
	return pc.results
}

// normalizeURL normalizes URLs for consistent visited map checks
func normalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// Remove fragments and normalize to scheme://host/path
	return u.Scheme + "://" + u.Host + u.Path, nil
}
