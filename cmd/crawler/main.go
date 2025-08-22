package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/yosofbadr/WebCrawler/cmd/crawler/app"
)

func main() {
	startURL := flag.String("startURL", "", "The URL to start crawling from")
	outputFile := flag.String("outputFile", "crawler-results.json", "The file to write the results to")
	statsFile := flag.String("statsFile", "crawler-stats.json", "The file to write the stats to")
	workers := flag.Int("workers", 10, "The number of workers to use for crawling")
	logFile := flag.String("logFile", "crawler.log", "The file to write logs to")
	flag.Parse()

	if err := app.Run(*startURL, *outputFile, *statsFile, *logFile, *workers); err != nil {
		slog.Error("Crawler failed", "error", err)
		os.Exit(1)
	}
}
