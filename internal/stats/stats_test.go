package stats

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/yosofbadr/WebCrawler/internal/storage"
)

func TestGenerateFromFile(t *testing.T) {
	// Create a dummy input file
	inputFile := "test_results.json"
	outputFile := "test_stats.json"

	results := []storage.Result{
		{Page: "https://example.com", Links: []string{"https://example.com/about", "https://example.com/contact"}},
		{Page: "https://example.com/about", Links: []string{"https://example.com"}},
		{Page: "https://google.com", Links: []string{"https://google.com/search"}},
	}

	file, err := os.Create(inputFile)
	if err != nil {
		t.Fatalf("Unexpected error creating input file: %v", err)
	}
	defer file.Close()

	for _, result := range results {
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			t.Fatalf("Unexpected error encoding data to input file: %v", err)
		}
	}
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// Generate stats
	if err := GenerateFromFile(inputFile, outputFile); err != nil {
		t.Fatalf("Unexpected error generating stats: %v", err)
	}

	// Verify the output file
	statsFileContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Unexpected error reading output file: %v", err)
	}

	var actualStats Stats
	if err := json.Unmarshal(statsFileContent, &actualStats); err != nil {
		t.Fatalf("Unexpected error unmarshalling stats: %v", err)
	}

	if actualStats.TotalPagesVisited != 3 {
		t.Errorf("Expected total_pages_visited to be %d, but got %d", 3, actualStats.TotalPagesVisited)
	}

	if len(actualStats.VisitedLinks) != 3 {
		t.Errorf("Expected %d visited links, but got %d", 3, len(actualStats.VisitedLinks))
	}

	// We don't care about the order of visited links, just that they are present
	expectedVisitedLinks := map[string]struct{}{
		"https://example.com":       {},
		"https://example.com/about": {},
		"https://google.com":        {},
	}

	for _, link := range actualStats.VisitedLinks {
		if _, ok := expectedVisitedLinks[link]; !ok {
			t.Errorf("Unexpected visited link: %s", link)
		}
	}
}