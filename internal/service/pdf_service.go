package service

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/remiehneppo/be-task-management/types"
	"github.com/remiehneppo/be-task-management/utils"
)

type PDFService interface {
	GetTotalPages(filePath string) (int, error)
	ProcessPDF(req *types.ProcessPDFRequest) ([]*types.DocumentChunk, error)
	ExtractPageContent(req *types.ExtractPageContentRequest) ([]string, error)
}

type DocumentServiceConfig struct {
	MaxChunkSize int // Maximum size for text chunks
	OverlapSize  int // Size of overlap between chunks
	BatchSize    int // Number of pages to process in a batch
}

// pdfService handles PDF processing operations
// Implements the PDFService interface
type pdfService struct {
	maxChunkSize int // Maximum size of each text chunk
	overlapSize  int // Size of overlap between chunks
	batchSize    int // Number of pages to process in a batch
}

var DefaultDocumentServiceConfig = DocumentServiceConfig{
	MaxChunkSize: 1024,
	OverlapSize:  128,
	BatchSize:    3,
}

// NewPDFService creates a new PDF service with configurable chunk sizes
func NewPDFService(config DocumentServiceConfig) PDFService {
	return &pdfService{
		maxChunkSize: config.MaxChunkSize,
		overlapSize:  config.OverlapSize,
		batchSize:    config.BatchSize,
	}
}

func (s *pdfService) GetTotalPages(filePath string) (int, error) {
	cmd := exec.Command("pdfinfo", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("error running pdfinfo: %v", err)
	}

	scanner := bufio.NewScanner(&out)
	re := regexp.MustCompile(`Pages:\s+(\d+)`)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); len(matches) == 2 {
			return strconv.Atoi(matches[1])
		}
	}

	return 0, fmt.Errorf("unable to determine page count from pdfinfo")
}

// ProcessPDF reads and chunks a PDF file
// Parameters:
//   - filePath: Path to the PDF file
//   - req: Upload request containing metadata
//
// Returns:
//   - []*types.DocumentChunk: List of document chunks
//   - error: Error if processing fails
func (s *pdfService) ProcessPDF(req *types.ProcessPDFRequest) ([]*types.DocumentChunk, error) {
	chunks := make([]*types.DocumentChunk, 0)

	// Get total pages
	totalPages, err := s.GetTotalPages(req.FilePath)
	if err != nil {
		return nil, err
	}
	// Extract all text from the PDF
	texts, err := s.ExtractPageContent(&types.ExtractPageContentRequest{
		FilePath: req.FilePath,
		ToolUse:  req.ToolUse,
		FromPage: 1,
		ToPage:   totalPages,
	},
	)
	if err != nil {
		return nil, types.ErrFailedExtractTextFromPDF
	}

	lastText := ""

	// Process each page
	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		text := texts[pageNum-1]
		text = lastText + " " + s.cleanText(text)

		// Skip empty text
		if text == "" {
			continue
		}

		// Create chunks for this page
		pageChunks, newLastText := s.createChunks(text, len(chunks), pageNum)
		if len(pageChunks) == 0 {
			lastText = newLastText
			continue
		}
		if len(newLastText) > 0 {
			lastText = newLastText
			chunks = append(chunks, pageChunks[:len(pageChunks)-1]...)
		} else {
			chunks = append(chunks, pageChunks...)
		}
	}
	return chunks, nil
}

// createTempDir creates a temporary directory for processing
func (s *pdfService) createTempDir(pdfPath string) (string, error) {
	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		os.Mkdir("temp", os.ModePerm)
	}
	tempFolder := filepath.Join("temp", utils.GetFileNameWithoutExt(pdfPath))
	if _, err := os.Stat(tempFolder); err == nil {
		os.RemoveAll(tempFolder)
	}
	err := os.Mkdir(tempFolder, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	return tempFolder, nil
}

// convertPDFToImages converts all pages of a PDF file to images
// Parameters:
//   - pdfPath: Path to the PDF file
//   - outputDir: Directory to save images (optional, defaults to temp/<filename>)
//
// Returns:
//   - []string: Paths to generated image files
//   - error: Error if conversion fails
func (s *pdfService) convertPDFToImages(pdfPath string, outputDir string, fromPage int, toPage int) ([]string, error) {
	// Create temp directory if outputDir not specified
	if outputDir == "" {
		if _, err := os.Stat("temp"); os.IsNotExist(err) {
			os.Mkdir("temp", os.ModePerm)
		}
		outputDir = filepath.Join("temp", utils.GetFileNameWithoutExt(pdfPath))
		if _, err := os.Stat(outputDir); err == nil {
			os.RemoveAll(outputDir)
		}
		if err := os.Mkdir(outputDir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	} else {
		// Make sure output directory exists
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
				return nil, fmt.Errorf("failed to create output directory: %w", err)
			}
		}
	}
	var convertCmd *exec.Cmd
	// Convert PDF to images using pdftoppm
	if fromPage > 0 && toPage > 0 {
		convertCmd = exec.Command("pdftoppm",
			"-png",
			"-r", "450",
			"-f", strconv.Itoa(fromPage),
			"-l", strconv.Itoa(toPage),
			"-hide-annotations",
			pdfPath,
			filepath.Join(outputDir, "page"))
	} else {
		// Convert all pages
		convertCmd = exec.Command("pdftoppm",
			"-png",
			"-r", "450",
			"-hide-annotations",
			pdfPath,
			filepath.Join(outputDir, "page"))
	}
	var stderr bytes.Buffer
	convertCmd.Stderr = &stderr

	if err := convertCmd.Run(); err != nil {
		return nil, fmt.Errorf("error converting PDF to images: %w, stderr: %s", err, stderr.String())
	}

	// Find all generated images
	pattern := filepath.Join(outputDir, "page-*.png")
	imagePaths, err := filepath.Glob(pattern)
	if err != nil || len(imagePaths) == 0 {
		return nil, fmt.Errorf("failed to find generated images: %w", err)
	}

	// Sort image paths to ensure correct order
	sort.Strings(imagePaths)

	// for i, imagePath := range imagePaths {
	// 	processedImage := filepath.Join(outputDir, "processed_"+strconv.Itoa(i+1)+".png")

	// 	// Preprocess image with ImageMagick to improve OCR quality
	// 	preprocessCmd := exec.Command("convert", imagePath,
	// 		"-density", "300",
	// 		"-colorspace", "gray",
	// 		"-brightness-contrast", "0x30",
	// 		"-normalize",
	// 		"-despeckle",
	// 		"-filter", "Gaussian",
	// 		"-define", "filter:sigma=1.5",
	// 		"-threshold", "50%",
	// 		"-sharpen", "0x1.0",
	// 		processedImage)

	// 	if err := preprocessCmd.Run(); err != nil {
	// 		log.Printf("Warning: image preprocessing failed: %v, using original image", err)
	// 		processedImage = imagePath
	// 	}
	// 	imagePaths[i] = processedImage
	// }

	return imagePaths, nil
}

// ExtractPageContent extracts text from all pages of a PDF
// Parameters:
//   - filePath: Path to the PDF file
//
// Returns:
//   - []string: Extracted text for each page
//   - error: Error if extraction fails
func (s *pdfService) ExtractPageContent(req *types.ExtractPageContentRequest) ([]string, error) {
	totalPages, err := s.GetTotalPages(req.FilePath)
	if req.FromPage < 1 || req.ToPage > totalPages || req.FromPage > req.ToPage {
		return nil, fmt.Errorf("invalid page range: %d-%d", req.FromPage, req.ToPage)
	}
	results := make([]string, req.ToPage-req.FromPage+1)
	if err != nil {
		return nil, fmt.Errorf("failed to get total pages: %w", err)
	}
	if req.ToolUse == "ocr" {
		tempDir, err := s.createTempDir(req.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		imagePaths, err := s.convertPDFToImages(req.FilePath, tempDir, req.FromPage, req.ToPage)
		if err != nil {
			return nil, fmt.Errorf("failed to convert PDF to images: %w", err)
		}

		for batchStart := 0; batchStart < len(imagePaths); batchStart += s.batchSize {
			batchEnd := batchStart + s.batchSize
			if batchEnd > len(imagePaths) {
				batchEnd = len(imagePaths)
			}

			log.Printf("Processing batch %d-%d of %d pages", batchStart+1, batchEnd, len(imagePaths))

			err := s.processPageBatch(imagePaths[batchStart:batchEnd], batchStart, results)
			if err != nil {
				return nil, fmt.Errorf("error processing batch %d-%d: %w", batchStart+1, batchEnd, err)
			}
		}
	} else if req.ToolUse == "pdftotext" {

		for i := req.FromPage - 1; i < req.ToPage; i++ {
			text, err := s.extractTextWithPdftotext(req.FilePath, i+1)
			if err != nil {
				return nil, fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
			}
			results[i] = text
		}
	} else {
		return nil, fmt.Errorf("unsupported tool: %s", req.ToolUse)
	}

	return results, nil
}

// processPageBatch processes a batch of pages
// Parameters:
//   - imagePaths: List of image paths for the batch
//   - startIndex: Starting index of the batch
//   - results: Slice to store extracted text
//
// Returns:
//   - error: Error if processing fails
func (s *pdfService) processPageBatch(imagePaths []string, startIndex int, results []string) error {
	type PageText struct {
		PageNum int
		Text    string
	}
	textChan := make(chan PageText, len(imagePaths))

	var wg sync.WaitGroup

	for i, imgPath := range imagePaths {
		pageNum := startIndex + i
		wg.Add(1)
		go func(imgPath string, pageNum int) {
			defer wg.Done()

			if pageNum > startIndex {
				time.Sleep(100 * time.Millisecond)
			}

			text, err := s.extractTextWithTesseract(imgPath)
			if err != nil {
				log.Printf("Warning: failed to extract text from page %d: %v", pageNum+1, err)
				return
			}

			textChan <- PageText{
				PageNum: pageNum,
				Text:    text,
			}
		}(imgPath, pageNum)
	}

	go func() {
		wg.Wait()
		close(textChan)
	}()

	for pageText := range textChan {
		if pageText.PageNum >= 0 && pageText.PageNum < len(results) {
			results[pageText.PageNum] = pageText.Text
		}
	}

	return nil
}

// extractText attempts to extract text from a specific page using multiple methods
// Parameters:
//   - imgPath: Path to the image file
//
// Returns:
//   - string: Extracted text
//   - error: Error if extraction fails
func (s *pdfService) extractText() (string, error) {
	return "", nil
}

// createChunks splits text into overlapping chunks with proper sentence boundaries
// Parameters:
//   - text: Text to be chunked
//   - chunkNum: Starting chunk number
//   - pageNum: Page number of the text
//
// Returns:
//   - []*types.DocumentChunk: List of document chunks
//   - string: Remaining text that couldn't fit into chunks
func (s *pdfService) createChunks(text string, chunkNum, pageNum int) ([]*types.DocumentChunk, string) {
	var chunks []*types.DocumentChunk
	textLen := len(text)
	lastText := ""

	if textLen <= s.maxChunkSize {
		lastText = text
		return []*types.DocumentChunk{
			{
				Content: text,
				Page:    pageNum,
				Chunk:   chunkNum,
			},
		}, lastText
	}

	currentPos := 0
	previousPos := 0
	stuckCount := 0

	for currentPos < textLen {
		previousPos = currentPos

		chunkEnd := currentPos + s.maxChunkSize
		if chunkEnd >= textLen {
			chunk := strings.TrimSpace(text[currentPos:])
			if chunk != "" {
				chunks = append(chunks, &types.DocumentChunk{
					Content: chunk,
					Page:    pageNum,
					Chunk:   chunkNum,
				})
				lastText = chunk
			}
			break
		}

		sentenceEnd := chunkEnd
		for i := chunkEnd; i > currentPos; i-- {
			if i < textLen && (text[i] == '.' || text[i] == '?' || text[i] == '!') {
				sentenceEnd = i + 1
				break
			}
		}

		if sentenceEnd == chunkEnd {
			for i := chunkEnd; i > currentPos; i-- {
				if i < textLen && text[i] == ' ' {
					sentenceEnd = i
					break
				}
			}
		}

		if sentenceEnd <= currentPos || sentenceEnd == chunkEnd {
			sentenceEnd = currentPos + (s.maxChunkSize / 2)
			if sentenceEnd > textLen {
				sentenceEnd = textLen
			}
		}

		chunk := strings.TrimSpace(text[currentPos:sentenceEnd])
		if chunk != "" {
			chunks = append(chunks, &types.DocumentChunk{
				Content: chunk,
				Page:    pageNum,
				Chunk:   chunkNum,
			})
		}

		currentPos = sentenceEnd - s.overlapSize
		if currentPos < 0 {
			currentPos = 0
		}

		minProgress := s.maxChunkSize / 10
		if currentPos <= previousPos || (currentPos-previousPos) < minProgress {
			currentPos = previousPos + minProgress
			stuckCount++

			if stuckCount > 5 {
				log.Printf("Warning: Possible infinite loop detected in text chunking at position %d. Breaking.", currentPos)
				if currentPos < textLen {
					finalChunk := strings.TrimSpace(text[currentPos:])
					if finalChunk != "" {
						chunks = append(chunks, &types.DocumentChunk{
							Content: finalChunk,
							Page:    pageNum,
							Chunk:   chunkNum + 1,
						})
						lastText = finalChunk
					}
				}
				break
			}
		} else {
			stuckCount = 0
		}

		chunkNum++
	}

	return chunks, lastText
}

// extractTextWithTesseract extracts text using OCR when other methods fail
// Parameters:
//   - imgPath: Path to the image file
//
// Returns:
//   - string: Extracted text
//   - error: Error if extraction fails
func (s *pdfService) extractTextWithTesseract(imgPath string) (string, error) {
	log.Println("Try extracting with tesseract, page:", imgPath)

	ocrCmd := exec.Command("tesseract",
		imgPath,
		"stdout",
		"-l", "vie+rus",
		"--oem", "3",
		"--psm", "3",
		"--dpi", "450",
		// "-c", "textord_min_linesize=2.5",
		// "-c", "preserve_interword_spaces=1",
	)

	var ocrOut bytes.Buffer
	ocrCmd.Stdout = &ocrOut
	if err := ocrCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run tesseract: %w", err)
	}
	ocrText := ocrOut.String()
	if trimmed := strings.TrimSpace(ocrText); len(trimmed) > 0 {
		return trimmed, nil
	} else {
		return "", nil
	}
}

// cleanText cleans up extracted text by removing unwanted characters
// Parameters:
//   - text: Text to be cleaned
//
// Returns:
//   - string: Cleaned text
func (s *pdfService) cleanText(text string) string {
	replacements := map[string]string{
		"\u0000": "",   // Null character
		"\ufffd": "",   // Unicode replacement character
		"\u001b": "",   // Escape character
		"\r":     "",   // Carriage return
		"\f":     "\n", // Form feed to newline
		"  ":     " ",  // Multiple spaces to single space
		"":      "",   // Apple logo
		"‡":      "",   // Double dagger
		"†":      "",   // Dagger
	}

	cleaned := text
	for old, new := range replacements {
		cleaned = strings.ReplaceAll(cleaned, old, new)
	}

	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// extractTextWithPdftotext extracts text from a specific page of a PDF using pdftotext
// Parameters:
//   - filePath: Path to the PDF file
//   - pageNum: Page number to extract text from
//
// Returns:
//   - string: Extracted text
//   - error: Error if extraction fails
func (s *pdfService) extractTextWithPdftotext(filePath string, pageNum int) (string, error) {
	// Validate page number
	if pageNum < 1 {
		return "", fmt.Errorf("invalid page number: %d", pageNum)
	}

	// Construct the command to extract text from a specific page
	cmd := exec.Command("pdftotext", "-f", strconv.Itoa(pageNum), "-l", strconv.Itoa(pageNum), filePath, "-")
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run pdftotext: %w", err)
	}

	// Get the extracted text
	text := out.String()
	if trimmed := strings.TrimSpace(text); len(trimmed) > 0 {
		return trimmed, nil
	}

	return "", nil
}
