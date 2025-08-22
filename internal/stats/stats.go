package stats

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yosofbadr/WebCrawler/internal/storage"
)

type Stats struct {
	TotalPagesVisited int      `json:"total_pages_visited"`
	VisitedLinks      []string `json:"visited_links"`
}

func Save(stats *Stats, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create stats file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(stats); err != nil {
		return fmt.Errorf("failed to encode stats: %w", err)
	}

	return nil
}

func GenerateFromFile(resultsFilename string, statsFilename string) error {
	resultsFile, err := os.Open(resultsFilename)
	if err != nil {
		return fmt.Errorf("failed to open results file: %w", err)
	}
	defer resultsFile.Close()

	decoder := json.NewDecoder(resultsFile)
	visitedLinks := make(map[string]struct{})

	for {
		var result storage.Result
		if err := decoder.Decode(&result); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to decode result: %w", err)
		}
		visitedLinks[result.Page] = struct{}{}
	}

	links := make([]string, 0, len(visitedLinks))
	for link := range visitedLinks {
		links = append(links, link)
	}

	stats := &Stats{
		TotalPagesVisited: len(visitedLinks),
		VisitedLinks:      links,
	}

	return Save(stats, statsFilename)
}