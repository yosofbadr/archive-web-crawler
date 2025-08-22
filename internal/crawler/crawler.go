package crawler

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
)

type Option func(*Crawler)

type Crawler struct {
	startURL *url.URL
	logger   *slog.Logger
}

func New(startURL string, opts ...Option) (*Crawler, error) {
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		return nil, fmt.Errorf("invalid start URL: %w", err)
	}

	crawler := &Crawler{
		startURL: parsedURL,
		logger:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	for _, opt := range opts {
		opt(crawler)
	}

	return crawler, nil
}

func WithLogger(logger *slog.Logger) Option {
	return func(crawler *Crawler) {
		crawler.logger = logger
	}
}

func (crawler *Crawler) Crawl() {
	crawler.logger.Info("Starting crawling process...", "url", crawler.startURL.String())
	crawler.ProcessURL(crawler.startURL)
}

func (crawler *Crawler) ProcessURL(currentURL *url.URL) {
	normalisedURL := normalizeURL(currentURL)

	log := crawler.logger.With("url", normalisedURL)
	log.Info("Processing URL")

	// TODO: Fetch and parse the page, extract links, etc.
}

// normalizeURL removes the fragment (#) from a URL to treat it as canonical.
func normalizeURL(url *url.URL) string {
	url.Fragment = ""
	return url.String()
}
