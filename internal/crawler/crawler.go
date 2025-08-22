package crawler

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sync"

	"github.com/yosofbadr/WebCrawler/internal/fetcher"
	"github.com/yosofbadr/WebCrawler/internal/link"
)

type Option func(*Crawler)

type Crawler struct {
	startURL   *url.URL
	targetHost string
	workers    int
	logger     *slog.Logger
	fetcher    fetcher.HTTPFetcher

	waitGroup sync.WaitGroup
}

func New(startURL string, opts ...Option) (*Crawler, error) {
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		return nil, fmt.Errorf("invalid start URL: %w", err)
	}

	crawler := &Crawler{
		startURL:   parsedURL,
		targetHost: parsedURL.Host,
		workers:    5,
		logger:     slog.New(slog.NewTextHandler(os.Stdout, nil)),
		fetcher:    nil,
	}

	for _, opt := range opts {
		opt(crawler)
	}

	if crawler.fetcher == nil {
		return nil, fmt.Errorf("fetcher is not set")
	}

	return crawler, nil
}

func WithLogger(logger *slog.Logger) Option {
	return func(crawler *Crawler) {
		crawler.logger = logger
	}
}

func WithFetcher(fetcher fetcher.HTTPFetcher) Option {
	return func(crawler *Crawler) {
		crawler.fetcher = fetcher
	}
}

func (crawler *Crawler) Crawl() {
	crawler.logger.Info("Starting crawling process...", "url", crawler.startURL.String())

	tasks := make(chan *url.URL)

	crawler.waitGroup.Add(1)
	go func() {
		tasks <- crawler.startURL
	}()

	for i := 0; i < crawler.workers; i++ {
		go crawler.worker(i, tasks)
	}

	crawler.waitGroup.Wait()
	close(tasks)
}

func (crawler *Crawler) worker(id int, tasks chan *url.URL) {
	for currentURL := range tasks {
		crawler.processURL(currentURL, tasks)
		crawler.waitGroup.Done()
	}
}

func (crawler *Crawler) processURL(currentURL *url.URL, tasks chan<- *url.URL) {
	normalisedURL := normaliseURL(currentURL)

	log := crawler.logger.With("url", normalisedURL)
	log.Info("Processing URL")

	body, err := crawler.fetcher.Fetch(normalisedURL)
	if err != nil {
		log.Error("Failed to fetch page", "error", err)
		return
	}
	defer body.Close()

	links, err := link.Extract(body, currentURL)
	if err != nil {
		log.Error("Failed to extract links", "error", err)
		return
	}

	log.Info("Extracted links", "count", len(links))

	var foundLinks []string
	for _, linkURL := range links {
		foundLinks = append(foundLinks, linkURL.String())
		if crawler.isInScope(linkURL) {
			crawler.waitGroup.Add(1)
			go func(u *url.URL) { tasks <- u }(linkURL)
		}
	}
	fmt.Printf("Page: %s\nLinks: %v\n\n", normalisedURL, foundLinks)
}

func (crawler *Crawler) isInScope(url *url.URL) bool {
	return url.Host == crawler.targetHost
}

// normaliseURL removes the fragment (#) from a URL to treat it as canonical.
func normaliseURL(url *url.URL) string {
	url.Fragment = ""
	return url.String()
}
