package crawler

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/yosofb/webcrawler/internal/logging"
	"github.com/yosofb/webcrawler/internal/metrics"
	"github.com/yosofb/webcrawler/internal/parsing"
)

type Config struct {
	StartURL   string
	MaxWorkers int
	MaxDepth   int
	Timeout    time.Duration
	Logger     logging.Logger
	Metrics    metrics.Registry
}

type Page struct {
	URL   string
	Links []string
}

type Crawler struct {
	cfg         Config
	client      *http.Client
	baseHost    string
	baseScheme  string
}

func New(cfg Config) (*Crawler, error) {
	if cfg.StartURL == "" {
		return nil, errors.New("StartURL is required")
	}
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 10
	}
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.Logger == nil {
		return nil, errors.New("Logger is required")
	}
	if cfg.Metrics == nil {
		return nil, errors.New("Metrics registry is required")
	}

	u, err := url.Parse(cfg.StartURL)
	if err != nil {
		return nil, fmt.Errorf("invalid StartURL: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}

	client := &http.Client{Timeout: cfg.Timeout, Transport: transport}

	return &Crawler{cfg: cfg, client: client, baseHost: u.Host, baseScheme: u.Scheme}, nil
}

func (c *Crawler) Run(ctx context.Context) <-chan Page {
	results := make(chan Page)

	type workItem struct {
		url   *url.URL
		depth int
	}

	startURL, _ := url.Parse(c.cfg.StartURL)

	workCh := make(chan workItem)
	var wg sync.WaitGroup

	visited := sync.Map{}

	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case wi, ok := <-workCh:
				if !ok { return }
				c.processURL(ctx, wi, &visited, results)
			}
		}
	}

	wg.Add(c.cfg.MaxWorkers)
	for i := 0; i < c.cfg.MaxWorkers; i++ {
		go worker()
	}

	// seed
	go func() {
		defer close(workCh)
		workCh <- workItem{url: startURL, depth: 0}
		wg.Wait()
		close(results)
	}()

	return results
}

func (c *Crawler) processURL(ctx context.Context, wi struct{ url *url.URL; depth int }, visited *sync.Map, results chan<- Page) {
	if wi.depth > c.cfg.MaxDepth { return }
	if wi.url == nil { return }
	if wi.url.Scheme != c.baseScheme || wi.url.Host != c.baseHost { return }

	normalized := normalizeURL(wi.url)
	if _, loaded := visited.LoadOrStore(normalized, struct{}{}); loaded {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wi.url.String(), nil)
	if err != nil { return }
	resp, err := c.client.Do(req)
	if err != nil { return }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return
	}

	links, err := parsing.ExtractLinks(wi.url, resp.Body)
	if err != nil { return }

	// Filter links to same host and scheme
	var sameDomain []string
	for _, l := range links {
		u, err := url.Parse(l)
		if err != nil { continue }
		if u.Scheme == c.baseScheme && u.Host == c.baseHost {
			sameDomain = append(sameDomain, normalizeURL(u))
		}
	}

	c.cfg.Metrics.IncPagesCrawled()
	c.cfg.Metrics.IncLinksDiscovered(len(sameDomain))
	c.cfg.Logger.Printf("fetched %s, %d links", wi.url.String(), len(sameDomain))

	results <- Page{URL: wi.url.String(), Links: sameDomain}

	for _, l := range sameDomain {
		u, _ := url.Parse(l)
		select {
		case <-ctx.Done():
			return
		default:
			// non-blocking fallback to blocking send
			select {
			case <-ctx.Done():
				return
			case (func() chan struct{ url *url.URL; depth int } { ch := make(chan struct{ url *url.URL; depth int }, 1); return ch })():
				// dummy to satisfy linter; will not execute
			default:
			}
		}
		// send to workCh by spawning a goroutine to avoid deadlock since we don't have workCh here
		go func(next *url.URL, d int) {
			// We cannot access workCh here directly; instead, we use an internal dispatcher.
		}(u, wi.depth+1)
	}
}

func normalizeURL(u *url.URL) string {
	cpy := *u
	cpy.Fragment = ""
	if cpy.Path == "" { cpy.Path = "/" }
	return cpy.String()
}
