package crawler

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sync"

	"github.com/yosofbadr/WebCrawler/internal/fetcher"
	"github.com/yosofbadr/WebCrawler/internal/link"
	"github.com/yosofbadr/WebCrawler/internal/storage"
	"golang.org/x/time/rate"
)

type Option func(*Crawler)

type Crawler struct {
	startURL   *url.URL
	targetHost string
	workers    int
	limiter    *rate.Limiter
	logger     *slog.Logger
	fetcher    fetcher.HTTPFetcher
	storage    *storage.Storage

	visited   sync.Map
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
		workers:    10,
		limiter:    rate.NewLimiter(10, 1),
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

func WithStorage(storage *storage.Storage) Option {
	return func(crawler *Crawler) {
		crawler.storage = storage
	}
}

func WithWorkers(workers int) Option {
	return func(c *Crawler) {
		c.workers = workers
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

	if !crawler.markAsVisited(normalisedURL) {
		return
	}

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

	result := &storage.Result{
		Page: normalisedURL,
	}
	for _, linkURL := range links {
		result.Links = append(result.Links, linkURL.String())
		if crawler.isInScope(linkURL) {
			crawler.waitGroup.Add(1)
			go func(url *url.URL) { tasks <- url }(linkURL)
		}
	}

	if crawler.storage != nil {
		if err := crawler.storage.Save(result); err != nil {
			log.Error("Failed to save result", "error", err)
		}
	}
}

func (crawler *Crawler) isInScope(url *url.URL) bool {
	return url.Host == crawler.targetHost
}

func (crawler *Crawler) markAsVisited(url string) bool {
	_, loaded := crawler.visited.LoadOrStore(url, struct{}{})
	return !loaded
}

// normaliseURL removes the fragment (#) from a URL to treat it as canonical.
func normaliseURL(url *url.URL) string {
	url.Fragment = ""
	return url.String()
}
