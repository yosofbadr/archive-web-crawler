package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/yosofbadr/WebCrawler/internal/crawler"
)

func main() {
	startURL := flag.String("startURL", "", "The URL to start crawling from")
	flag.Parse()

	if *startURL == "" {
		slog.Error("Please provide a starting URL using the -url flag")
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	crawler, err := crawler.New(*startURL, crawler.WithLogger(logger))
	if err != nil {
		logger.Error("Failed to create crawler", "error", err)
		os.Exit(1)
	}

	crawler.Crawl()
	logger.Info("Crawler finished")
}
