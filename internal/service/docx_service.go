package service

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/remiehneppo/be-task-management/types"
)

// DOCXService defines the interface for reading text from DOCX files
type DOCXService interface {
	ProcessDocx(filePath string) ([]*types.DocumentChunk, error)
	ReadText(filePath string) ([]string, error)
}

// docxService implements the DOCXService interface
type docxService struct {
	maxChunkSize int
}

// NewDOCXService creates a new instance of DOCXService
func NewDOCXService(
	maxChunkSize int,
) DOCXService {
	return &docxService{
		maxChunkSize: maxChunkSize,
	}
}

// ReadText reads and extracts text from a DOCX file using pandoc
// Parameters:
//   - filePath: Path to the DOCX file
//
// Returns:
//   - []string: Extracted text split into pages
//   - error: Error if reading fails
func (s *docxService) ReadText(filePath string) ([]string, error) {
	// Prepare the pandoc command
	cmd := exec.Command("pandoc", "-f", "docx", "-t", "plain", filePath)

	// Capture the output
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to read DOCX file: %v, stderr: %s", err, stderr.String())
	}

	// Get the extracted text
	extractedText := out.String()

	paras := strings.Split(extractedText, "\n\n")

	// Trim whitespace from each page
	for i := range paras {
		paras[i] = strings.TrimSpace(paras[i])
	}

	return paras, nil
}

// ProcessDocx processes a DOCX file and returns its content as chunks
// Parameters:
//   - filePath: Path to the DOCX file
//
// Returns:
//   - []*types.DocumentChunk: List of document chunks
//   - error: Error if processing fails
func (s *docxService) ProcessDocx(filePath string) ([]*types.DocumentChunk, error) {
	// Read the text from the DOCX file
	paragraphs, err := s.ReadText(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read text from DOCX file: %w", err)
	}

	var chunks []*types.DocumentChunk
	chunkID := 0

	// Process each paragraph
	for _, paragraph := range paragraphs {
		// Skip empty paragraphs
		if len(paragraph) == 0 {
			continue
		}

		// Split the paragraph into chunks if it exceeds maxChunkSize
		for len(paragraph) > s.maxChunkSize {
			// Create a chunk with maxChunkSize
			chunk := &types.DocumentChunk{
				Chunk:   len(chunks),
				Content: paragraph[:s.maxChunkSize],
			}
			chunks = append(chunks, chunk)
			chunkID++

			// Update the paragraph to the remaining content
			paragraph = paragraph[s.maxChunkSize:]
		}

		// Add the remaining part of the paragraph as a chunk
		if len(paragraph) > 0 {
			chunk := &types.DocumentChunk{
				Chunk:   len(chunks),
				Content: paragraph,
			}
			chunks = append(chunks, chunk)
			chunkID++
		}
	}

	return chunks, nil
}
