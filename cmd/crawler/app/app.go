package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/yosofbadr/WebCrawler/internal/crawler"
	"github.com/yosofbadr/WebCrawler/internal/fetcher"
	"github.com/yosofbadr/WebCrawler/internal/stats"
	"github.com/yosofbadr/WebCrawler/internal/storage"
)

func Run(startURL, outputFile, statsFile, logFile string, workers int) error {
	if startURL == "" {
		return fmt.Errorf("please provide a starting URL using the -url flag")
	}

	logF, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logF.Close()

	logger := slog.New(slog.NewJSONHandler(logF, nil))

	storageSaver, err := storage.New(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}
	defer storageSaver.Close()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	fetcher := fetcher.New(httpClient)

	crawler, err := crawler.New(startURL, crawler.WithLogger(logger), crawler.WithFetcher(fetcher), crawler.WithStorage(storageSaver), crawler.WithWorkers(workers))
	if err != nil {
		return fmt.Errorf("failed to create crawler: %w", err)
	}

	crawler.Crawl()
	logger.Info("Crawler finished")

	if err := stats.GenerateFromFile(outputFile, statsFile); err != nil {
		return fmt.Errorf("failed to generate stats: %w", err)
	}

	return nil
}
