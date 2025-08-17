package writer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specmint/specmint/internal/config"
)

// Writer handles output writing in various formats
type Writer struct {
	config    config.Output
	outputDir string
}

// New creates a new writer instance
func New(config config.Output) (*Writer, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(config.Directory, 0750); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &Writer{
		config:    config,
		outputDir: config.Directory,
	}, nil
}

// WriteRecords writes the generated records to the output file
func (w *Writer) WriteRecords(records []map[string]interface{}) error {
	switch w.config.Format {
	case "json":
		return w.writeJSON(records)
	case "jsonl":
		return w.writeJSONL(records)
	default:
		return w.writeJSONL(records) // Default to JSONL
	}
}

// WriteManifest writes the generation manifest
func (w *Writer) WriteManifest(manifest map[string]interface{}) error {
	manifestPath := filepath.Join(w.outputDir, "manifest.json")

	file, err := os.Create(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to create manifest file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(manifest); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// writeJSON writes records as a single JSON array
func (w *Writer) writeJSON(records []map[string]interface{}) error {
	outputPath := filepath.Join(w.outputDir, "dataset.json")

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(records); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

// writeJSONL writes records as JSON Lines (one JSON object per line)
func (w *Writer) writeJSONL(records []map[string]interface{}) error {
	outputPath := filepath.Join(w.outputDir, "dataset.jsonl")

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// GetOutputPath returns the path where records were written
func (w *Writer) GetOutputPath() string {
	switch w.config.Format {
	case "json":
		return filepath.Join(w.outputDir, "dataset.json")
	default:
		return filepath.Join(w.outputDir, "dataset.jsonl")
	}
}
