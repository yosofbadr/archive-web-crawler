package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/yosofbadr/WebCrawler/internal/crawler"
	"github.com/yosofbadr/WebCrawler/internal/fetcher"
)

func main() {
	startURL := flag.String("startURL", "", "The URL to start crawling from")
	flag.Parse()

	if *startURL == "" {
		slog.Error("Please provide a starting URL using the -url flag")
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	fetcher := fetcher.New(httpClient)

	crawler, err := crawler.New(*startURL, crawler.WithLogger(logger), crawler.WithFetcher(fetcher))
	if err != nil {
		logger.Error("Failed to create crawler", "error", err)
		os.Exit(1)
	}

	crawler.Crawl()
	logger.Info("Crawler finished")
}
