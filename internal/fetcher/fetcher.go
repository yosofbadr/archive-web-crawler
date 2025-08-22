package fetcher

import (
	"net/http"
)

type HTTPFetcher interface{}

type Fetcher struct {
	client *http.Client
}

func New(client *http.Client) *Fetcher {
	return &Fetcher{client: client}
}
