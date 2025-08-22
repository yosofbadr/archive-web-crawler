package main_test

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/yosofbadr/WebCrawler/internal/stats"
	"github.com/yosofbadr/WebCrawler/internal/storage"
	"net/http"
	"net/http/httptest"
)

func TestE2E(t *testing.T) {
	// Create a temporary directory for output files and binary
	tempDir, err := os.MkdirTemp("", "crawler_e2e_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "crawler-results.json")
	statsFile := filepath.Join(tempDir, "crawler-stats.json")
	logFile := filepath.Join(tempDir, "crawler.log")
	binaryPath := filepath.Join(tempDir, "crawler_app")

	// Compile the crawler application
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/crawler")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to compile crawler: %v", err)
	}

	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`<html><body><a href="/page1">Page 1</a><a href="/page2">Page 2</a></body></html>`))
		case "/page1":
			w.Write([]byte(`<html><body><a href="/page3">Page 3</a></body></html>`))
		case "/page2":
			w.Write([]byte(`<html><body>No links here</body></html>`))
		case "/page3":
			w.Write([]byte(`<html><body>Final page</body></html>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Run the compiled crawler binary
	cmd = exec.Command(binaryPath,
		"-startURL", server.URL,
		"-outputFile", outputFile,
		"-statsFile", statsFile,
		"-logFile", logFile,
		"-workers", "1",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Crawler binary failed: %v", err)
	}

	// Verify output files exist
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file %s was not created", outputFile)
	}
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		t.Errorf("Stats file %s was not created", statsFile)
	}

	// Verify content of results file
	file, err := os.Open(outputFile)
	if err != nil {
		t.Fatalf("Failed to open results file: %v", err)
	}
	defer file.Close()

	var results []storage.Result
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var res storage.Result
		if err := json.Unmarshal(scanner.Bytes(), &res); err != nil {
			t.Fatalf("Failed to unmarshal result line: %v", err)
		}
		results = append(results, res)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading results file: %v", err)
	}

	expectedResults := []storage.Result{
		{Page: server.URL, Links: []string{server.URL + "/page1", server.URL + "/page2"}},
		{Page: server.URL + "/page1", Links: []string{server.URL + "/page3"}},
		{Page: server.URL + "/page2", Links: []string{}},
		{Page: server.URL + "/page3", Links: []string{}},
	}

	// Basic check for now, more robust comparison needed
	// For simplicity, just check if all expected pages are present
	foundPages := make(map[string]bool)
	for _, res := range results {
		foundPages[res.Page] = true
	}

	for _, expectedRes := range expectedResults {
		if !foundPages[expectedRes.Page] {
			t.Errorf("Expected page %s not found in results", expectedRes.Page)
		}
	}

	// Verify content of stats file
	statsContent, err := os.ReadFile(statsFile)
	if err != nil {
		t.Fatalf("Failed to read stats file: %v", err)
	}

	var actualStats stats.Stats
	if err := json.Unmarshal(statsContent, &actualStats); err != nil {
		t.Fatalf("Failed to unmarshal stats: %v", err)
	}

	if actualStats.TotalPagesVisited != 4 {
		t.Errorf("Expected total_pages_visited to be %d, but got %d", 4, actualStats.TotalPagesVisited)
	}

	// We don't care about the order of visited links, just that they are present
	expectedVisitedLinks := map[string]struct{}{
		server.URL:               {},
		server.URL + "/page1": {},
		server.URL + "/page2": {},
		server.URL + "/page3": {},
	}

	if len(actualStats.VisitedLinks) != len(expectedVisitedLinks) {
		t.Errorf("Expected %d visited links, but got %d", len(expectedVisitedLinks), len(actualStats.VisitedLinks))
	}

	for _, link := range actualStats.VisitedLinks {
		if _, ok := expectedVisitedLinks[link]; !ok {
			t.Errorf("Unexpected visited link: %s", link)
		}
	}
}
