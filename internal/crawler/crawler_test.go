package crawler

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/yosofbadr/WebCrawler/internal/storage"
)

// MockFetcher is a mock implementation of the Fetcher interface.
type MockFetcher struct {
	mu       sync.Mutex
	fetchMap map[string]string
	fetchErr error
}

func NewMockFetcher() *MockFetcher {
	return &MockFetcher{
		fetchMap: make(map[string]string),
	}
}

func (m *MockFetcher) Fetch(url string) (io.ReadCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	return io.NopCloser(strings.NewReader(m.fetchMap[url])), nil
}

// MockStorage is a mock implementation of the Storage interface.
type MockStorage struct {
	mu     sync.Mutex
	results []storage.Result
	saveErr  error
}

func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

func (m *MockStorage) Save(result *storage.Result) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	m.results = append(m.results, *result)
	return nil
}

func (m *MockStorage) Close() error {
	return nil
}

func TestNew(t *testing.T) {
	startURL := "https://example.com"
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	fetcher := NewMockFetcher()
	storage := NewMockStorage()

	c, err := New(startURL, WithLogger(logger), WithFetcher(fetcher), WithStorage(storage), WithWorkers(1))
	if err != nil {
		t.Fatalf("Unexpected error creating crawler: %v", err)
	}

	if c.startURL.String() != startURL {
		t.Errorf("Expected startURL to be %s, but got %s", startURL, c.startURL.String())
	}
	if c.logger != logger {
		t.Errorf("Expected logger to be %v, but got %v", logger, c.logger)
	}
	if c.fetcher != fetcher {
		t.Errorf("Expected fetcher to be %v, but got %v", fetcher, c.fetcher)
	}
	if c.storage != storage {
		t.Errorf("Expected storage to be %v, but got %v", storage, c.storage)
	}
	if c.workers != 1 {
		t.Errorf("Expected workers to be %d, but got %d", 1, c.workers)
	}
}

func TestCrawl(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`<a href="/page1">Page 1</a><a href="/page2">Page 2</a>`))
		case "/page1":
			w.Write([]byte(`<a href="/page3">Page 3</a>`))
		case "/page2":
			w.Write([]byte(`No links here`))
		case "/page3":
			w.Write([]byte(`Final page`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Setup mock dependencies
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	
	// Use the mock fetcher to control responses
	mockFetcher := NewMockFetcher()
	mockFetcher.fetchMap[server.URL] = `<a href="/page1">Page 1</a><a href="/page2">Page 2</a>`
	mockFetcher.fetchMap[server.URL + "/page1"] = `<a href="/page3">Page 3</a>`
	mockFetcher.fetchMap[server.URL + "/page2"] = `No links here`
	mockFetcher.fetchMap[server.URL + "/page3"] = `Final page`

	mockStorage := NewMockStorage()

	// Create crawler
	c, err := New(server.URL, WithLogger(logger), WithFetcher(mockFetcher), WithStorage(mockStorage), WithWorkers(1))
	if err != nil {
		t.Fatalf("Unexpected error creating crawler: %v", err)
	}

	// Perform crawl
	c.Crawl()

	// Assert results
	expectedResults := []storage.Result{
		{Page: server.URL, Links: []string{server.URL + "/page1", server.URL + "/page2"}},
		{Page: server.URL + "/page1", Links: []string{server.URL + "/page3"}},
		{Page: server.URL + "/page2", Links: []string{}},
		{Page: server.URL + "/page3", Links: []string{}},
	}

	// Sort results for consistent comparison
	// TODO: Implement sorting or a better comparison method

	if len(mockStorage.results) != len(expectedResults) {
		t.Fatalf("Expected %d results, but got %d", len(expectedResults), len(mockStorage.results))
	}

	// Basic check for now, more robust comparison needed
	// For simplicity, just check if all expected pages are present
	foundPages := make(map[string]bool)
	for _, res := range mockStorage.results {
		foundPages[res.Page] = true
	}

	for _, expectedRes := range expectedResults {
		if !foundPages[expectedRes.Page] {
			t.Errorf("Expected page %s not found in results", expectedRes.Page)
		}
	}
}
