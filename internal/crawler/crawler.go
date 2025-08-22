package crawler

import (
	"fmt"
	"log/slog"
	"net/url"
)

type Crawler struct {
	startURL *url.URL
	logger   *slog.Logger
}

func New(startURL string) (*Crawler, error) {
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		return nil, fmt.Errorf("invalid start URL: %w", err)
	}

	c := &Crawler{
		startURL: parsedURL,
	}

	return c, nil
}
