package producer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/MishkaRogachev/command-queue-executor/pkg/models"
)

// FileRequestFeed is a request feed that reads requests from a file
type FileRequestFeed struct {
	file    *os.File
	scanner *bufio.Scanner
	isEmpty bool
}

// NewFileRequestFeed creates a new FileRequestFeed instance
func NewFileRequestFeed(filePath string) (*FileRequestFeed, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &FileRequestFeed{
		file:    file,
		scanner: bufio.NewScanner(file),
		isEmpty: false,
	}, nil
}

// Next returns the next request from the file
func (f *FileRequestFeed) Next() (models.RequestWrapper, error) {
	if !f.scanner.Scan() {
		f.isEmpty = true
		return models.RequestWrapper{}, fmt.Errorf("no more requests")
	}

	var cmd models.RequestWrapper
	if err := json.Unmarshal([]byte(f.scanner.Text()), &cmd); err != nil {
		return models.RequestWrapper{}, fmt.Errorf("failed to parse request: %w", err)
	}

	return cmd, nil
}

// IsEmpty returns true if there are no more requests in the file
func (f *FileRequestFeed) IsEmpty() bool {
	return f.isEmpty
}

// Close closes the file
func (f *FileRequestFeed) Close() {
	f.file.Close()
}
