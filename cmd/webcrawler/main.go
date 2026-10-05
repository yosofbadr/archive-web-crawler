package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/yosofb/webcrawler/internal/crawler"
	"github.com/yosofb/webcrawler/internal/logging"
	"github.com/yosofb/webcrawler/internal/metrics"
)

func main() {
	var (
		startURL   string
		maxWorkers int
		maxDepth   int
		timeout    time.Duration
	)

	flag.StringVar(&startURL, "url", "", "Starting URL to crawl (required)")
	flag.IntVar(&maxWorkers, "workers", 10, "Maximum number of concurrent workers")
	flag.IntVar(&maxDepth, "depth", 5, "Maximum crawl depth from the start URL")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "HTTP client timeout")
	flag.Parse()

	if startURL == "" {
		fmt.Fprintln(os.Stderr, "-url is required")
		flag.Usage()
		os.Exit(2)
	}

	u, err := url.Parse(startURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		fmt.Fprintf(os.Stderr, "invalid -url: %v\n", err)
		os.Exit(2)
	}

	logger := logging.New()
	metricsRegistry := metrics.NewRegistry()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := crawler.Config{
		StartURL:   startURL,
		MaxWorkers: maxWorkers,
		MaxDepth:   maxDepth,
		Timeout:    timeout,
		Logger:     logger,
		Metrics:    metricsRegistry,
	}

	c, err := crawler.New(config)
	if err != nil {
		log.Fatalf("failed to create crawler: %v", err)
	}

	results := c.Run(ctx)
	for page := range results {
		fmt.Printf("Visited: %s\n", page.URL)
		for _, link := range page.Links {
			fmt.Printf("  - %s\n", link)
		}
	}
}
