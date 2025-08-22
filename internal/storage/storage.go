
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Result struct {
	Page  string   `json:"page"`
	Links []string `json:"links"`
}

type Storage struct {
	file   *os.File
	mu     sync.Mutex
	encoder *json.Encoder
}

func New(filename string) (*Storage, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage file: %w", err)
	}

	return &Storage{
		file:   file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (s *Storage) Save(result *Result) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode result: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.file.Close()
}
