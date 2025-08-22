package main

import (
	"flag"
	"log/slog"
	"os"
)

func main() {
	startURL := flag.String("startURL", "", "The URL to start crawling from")
	flag.Parse()

	if *startURL == "" {
		slog.Error("Please provide a starting URL using the -url flag")
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// TODO: Initialize and start the crawler with the provided URL
	logger.Info("Starting crawler", "url", *startURL)
	logger.Info("Crawler finished")
}
