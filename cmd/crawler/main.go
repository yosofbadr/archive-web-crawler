package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/yosofbadr/WebCrawler/internal/crawler"
	"github.com/yosofbadr/WebCrawler/internal/fetcher"
	"github.com/yosofbadr/WebCrawler/internal/stats"
	"github.com/yosofbadr/WebCrawler/internal/storage"
)

func main() {
	startURL := flag.String("startURL", "", "The URL to start crawling from")
	outputFile := flag.String("outputFile", "crawler-results.json", "The file to write the results to")
	statsFile := flag.String("statsFile", "crawler-stats.json", "The file to write the stats to")
	workers := flag.Int("workers", 10, "The number of workers to use for crawling")
	flag.Parse()

	if *startURL == "" {
		slog.Error("Please provide a starting URL using the -url flag")
		os.Exit(1)
	}

	logFile, err := os.OpenFile("crawler.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		os.Exit(1)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, nil))

	storage, err := storage.New(*outputFile)
	if err != nil {
		logger.Error("Failed to create storage", "error", err)
		os.Exit(1)
	}
	defer storage.Close()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	fetcher := fetcher.New(httpClient)

	crawler, err := crawler.New(*startURL, crawler.WithLogger(logger), crawler.WithFetcher(fetcher), crawler.WithStorage(storage), crawler.WithWorkers(*workers))
	if err != nil {
		logger.Error("Failed to create crawler", "error", err)
		os.Exit(1)
	}

	crawler.Crawl()
	logger.Info("Crawler finished")

	if err := stats.GenerateFromFile(*outputFile, *statsFile); err != nil {
		logger.Error("Failed to generate stats", "error", err)
	}
}
