package storage

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	file := "test.json"
	s, err := New(file)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer s.Close()
	defer os.Remove(file)

	if s.file == nil {
		t.Error("Expected file to be non-nil")
	}
}

func TestSave(t *testing.T) {
	file := "test.json"
	s, err := New(file)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer s.Close()
	defer os.Remove(file)

	result := &Result{
		Page:  "https://example.com",
		Links: []string{"https://example.com/about"},
	}

	if err := s.Save(result); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	fileContents, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Unexpected error reading file: %v", err)
	}

	var savedResult Result
	if err := json.Unmarshal(fileContents, &savedResult); err != nil {
		t.Fatalf("Unexpected error unmarshalling data: %v", err)
	}

	if savedResult.Page != result.Page {
		t.Fatalf("Expected page %s, but got %s", result.Page, savedResult.Page)
	}

	if len(savedResult.Links) != len(result.Links) {
		t.Fatalf("Expected %d links, but got %d", len(result.Links), len(savedResult.Links))
	}

	if savedResult.Links[0] != result.Links[0] {
		t.Fatalf("Expected link %s, but got %s", result.Links[0], savedResult.Links[0])
	}
}
