package fetcher

import (
	"fmt"
	"io"
	"net/http"
)

type HTTPFetcher interface {
	Fetch(url string) (io.ReadCloser, error)
}

type Fetcher struct {
	client *http.Client
}

func New(client *http.Client) *Fetcher {
	return &Fetcher{client: client}
}

func (fetcher *Fetcher) Fetch(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}
	req.Header.Set("User-Agent", "SimpleGoCrawler/1.0")

	resp, err := fetcher.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("bad status for %s: %s", url, resp.Status)
	}

	return resp.Body, nil
}
