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
)

type DocumentServiceConfig struct {
	MaxChunkSize int // Maximum size for text chunks
	OverlapSize  int // Size of overlap between chunks// Total number of pages in the document
	BatchSize    int // Number of pages to process in a batch
}

// PDFService handles PDF processing operations
type PDFService struct {
	maxChunkSize int // Maximum size of each text chunk
	overlapSize  int // Size of overlap between chunks
	batchSize    int // Number of pages to process in a batch
}

var DefaultDocumentServiceConfig = DocumentServiceConfig{
	MaxChunkSize: 1024,
	OverlapSize:  128,
	BatchSize:    3,
}

// PDFChunk represents a processed chunk of PDF text with metadata

// NewPDFService creates a new PDF service with configurable chunk sizes
func NewPDFService(config DocumentServiceConfig) *PDFService {

	return &PDFService{
		maxChunkSize: config.MaxChunkSize,
		overlapSize:  config.OverlapSize,
		batchSize:    config.BatchSize,
	}
}

func (s *PDFService) GetTotalPages(filePath string) (int, error) {
	totalPages, err := getNumPages(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get total pages: %w", err)
	}
	return totalPages, nil
}

// ProcessPDF reads and chunks a PDF file
// Parameters:
//   - filePath: Path to the PDF file
//   - c: Channel to send processed chunks
//
// Returns:
//   - error: Error if processing fails
func (s *PDFService) ProcessPDF(filePath string, req types.UploadRequest) ([]*types.DocumentChunk, error) {
	chunks := make([]*types.DocumentChunk, 0)
	// Get total pages
	totalPages, err := getNumPages(filePath)
	if err != nil {
		return nil, err
	}

	// Extract all text from the PDF
	texts, err := s.extractAllText(filePath)
	if err != nil {
		return nil, types.ErrFailedExtractTextFromPDF
	}
	lastText := ""
	// Process each page
	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		text := texts[pageNum-1]
		// Clean text
		text = lastText + " " + s.cleanText(text)

		// Skip empty text
		if text == "" {
			continue
		}

		// Create metadata for this page
		// metadata := types.DocumentMetadata{
		// 	Source:     req.Source,
		// 	Title:      req.Title + ".pdf",
		// 	TotalPages: totalPages,
		// 	Tags:       req.Tags,
		// }
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

// getFileNameWithoutExt extracts filename without extension from a file path
func GetFileNameWithoutExt(filepath string) string {
	// Get base filename from path
	base := filepath[strings.LastIndex(filepath, "/")+1:]

	// Remove extension
	if idx := strings.LastIndex(base, "."); idx != -1 {
		base = base[:idx]
	}

	return base
}

func (s *PDFService) CreateTempDir(pdfPath string) (string, error) {
	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		os.Mkdir("temp", os.ModePerm)
	}
	tempFolder := filepath.Join("temp", GetFileNameWithoutExt(pdfPath))
	if _, err := os.Stat(tempFolder); err == nil {
		os.RemoveAll(tempFolder)
	}
	err := os.Mkdir(tempFolder, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	return tempFolder, nil
}

// ConvertPDFToImages converts all pages of a PDF file to images
// Parameters:
//   - pdfPath: Path to the PDF file
//   - outputDir: Directory to save images (optional, defaults to temp/<filename>)
//
// Returns:
//   - []string: Paths to generated image files
//   - error: Error if conversion fails
func (s *PDFService) ConvertPDFToImages(pdfPath string, outputDir string) ([]string, error) {

	// Create temp directory if outputDir not specified
	if outputDir == "" {
		if _, err := os.Stat("temp"); os.IsNotExist(err) {
			os.Mkdir("temp", os.ModePerm)
		}
		outputDir = filepath.Join("temp", GetFileNameWithoutExt(pdfPath))
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

	// Convert PDF to images using pdftoppm
	// -png: output PNG images
	// -r 300: set resolution to 300 DPI
	// -thread: use multithreading
	convertCmd := exec.Command("pdftoppm",
		"-png",
		"-r", "300",
		// "-thread",
		"-hide-annotations",
		pdfPath,
		filepath.Join(outputDir, "page"))

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

	for i, imagePath := range imagePaths {
		processedImage := filepath.Join(outputDir, "processed_"+strconv.Itoa(i+1)+".png")

		// Tiền xử lý ảnh với ImageMagick để cải thiện chất lượng OCR
		// Các bước: grayscale -> tăng độ tương phản -> khử nhiễu -> binary threshold
		preprocessCmd := exec.Command("convert", imagePath,
			"-density", "300", // Đặt mật độ dpi
			"-colorspace", "gray", // Chuyển sang thang xám
			"-brightness-contrast", "0x30", // Tăng độ tương phản
			"-normalize",          // Cân bằng histogram
			"-despeckle",          // Khử điểm nhiễu
			"-filter", "Gaussian", // Lọc Gaussian
			"-define", "filter:sigma=1.5", // Thông số cho lọc
			"-threshold", "50%", // Phân loại đen trắng
			"-sharpen", "0x1.0", // Làm sắc nét
			processedImage)

		if err := preprocessCmd.Run(); err != nil {
			log.Printf("Warning: image preprocessing failed: %v, using original image", err)
			processedImage = imagePath
		}
		// Thay thế đường dẫn ảnh gốc bằng đường dẫn ảnh đã xử lý
		imagePaths[i] = processedImage
	}

	return imagePaths, nil
}

func (s *PDFService) extractAllText(filePath string) ([]string, error) {
	totalPages, err := getNumPages(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get total pages: %w", err)
	}
	tempDir, err := s.CreateTempDir(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Chuyển đổi PDF thành ảnh - giữ nguyên phần này
	imagePaths, err := s.ConvertPDFToImages(filePath, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to convert PDF to images: %w", err)
	}

	// Khởi tạo mảng kết quả
	results := make([]string, totalPages)

	// Xử lý theo batch thay vì tất cả các trang cùng lúc
	for batchStart := 0; batchStart < len(imagePaths); batchStart += s.batchSize {
		// Xác định kết thúc của batch hiện tại
		batchEnd := batchStart + s.batchSize
		if batchEnd > len(imagePaths) {
			batchEnd = len(imagePaths)
		}

		log.Printf("Processing batch %d-%d of %d pages", batchStart+1, batchEnd, len(imagePaths))

		// Xử lý một batch trang
		err := s.processPageBatch(imagePaths[batchStart:batchEnd], batchStart, results)
		if err != nil {
			return nil, fmt.Errorf("error processing batch %d-%d: %w", batchStart+1, batchEnd, err)
		}
	}

	return results, nil
}

// Hàm mới để xử lý một batch trang
func (s *PDFService) processPageBatch(imagePaths []string, startIndex int, results []string) error {
	// Tạo một channel để lưu dữ liệu văn bản từ mỗi trang
	type PageText struct {
		PageNum int
		Text    string
	}
	textChan := make(chan PageText, len(imagePaths))

	// WaitGroup để đảm bảo tất cả các công việc trong batch hoàn thành
	var wg sync.WaitGroup

	// Khởi chạy các worker để xử lý từng trang trong batch
	for i, imgPath := range imagePaths {
		pageNum := startIndex + i
		wg.Add(1)
		go func(imgPath string, pageNum int) {
			defer wg.Done()

			// Cho phép CPU nghỉ ngơi một chút giữa các trang
			if pageNum > startIndex {
				time.Sleep(100 * time.Millisecond)
			}

			// Trích xuất văn bản từ trang hiện tại
			text, err := s.ExtractText(imgPath)
			if err != nil {
				log.Printf("Warning: failed to extract text from page %d: %v", pageNum+1, err)
				return
			}

			// Gửi kết quả vào channel
			textChan <- PageText{
				PageNum: pageNum,
				Text:    text,
			}
		}(imgPath, pageNum)
	}

	// Goroutine để đóng channel kết quả khi tất cả các worker hoàn thành
	go func() {
		wg.Wait()
		close(textChan)
	}()

	// Thu thập kết quả của batch hiện tại
	for pageText := range textChan {
		if pageText.PageNum >= 0 && pageText.PageNum < len(results) {
			results[pageText.PageNum] = pageText.Text
		}
	}

	return nil
}

// extractText attempts to extract text from a specific page using multiple methods
func (s *PDFService) ExtractText(imgPath string) (string, error) {

	text, err := s.extractTextWithTesseract(imgPath)
	if err != nil {
		return "", fmt.Errorf("failed to extract text: %w", err)
	}
	// }
	return text, nil
}

// createChunks splits text into overlapping chunks with proper sentence boundaries
func (s *PDFService) createChunks(text string, chunkNum, pageNum int) ([]*types.DocumentChunk, string) {
	var chunks []*types.DocumentChunk
	textLen := len(text)
	lastText := ""

	// Return early if text fits in one chunk
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

	// Track consecutive iterations without significant progress
	stuckCount := 0

	for currentPos < textLen {
		// Save previous position to detect if we're making progress
		previousPos = currentPos

		// Calculate end position for current chunk
		chunkEnd := currentPos + s.maxChunkSize
		if chunkEnd >= textLen {
			// Handle last chunk
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

		// Find nearest sentence end
		sentenceEnd := chunkEnd
		for i := chunkEnd; i > currentPos; i-- {
			if i < textLen && (text[i] == '.' || text[i] == '?' || text[i] == '!') {
				sentenceEnd = i + 1
				break
			}
		}

		// If no sentence end found, use word boundary
		if sentenceEnd == chunkEnd {
			for i := chunkEnd; i > currentPos; i-- {
				if i < textLen && text[i] == ' ' {
					sentenceEnd = i
					break
				}
			}
		}

		// Safety check: if no sentence or word boundary found, force advancement
		if sentenceEnd <= currentPos || sentenceEnd == chunkEnd {
			// Force some minimum advancement (50% of max chunk size is reasonable)
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

		// Update position for next chunk
		currentPos = sentenceEnd - s.overlapSize
		if currentPos < 0 {
			currentPos = 0
		}

		// Safety check: Ensure we're making progress
		minProgress := s.maxChunkSize / 10 // At least 10% of chunk size
		if currentPos <= previousPos || (currentPos-previousPos) < minProgress {
			// Force advancement
			currentPos = previousPos + minProgress
			stuckCount++

			// If we're stuck multiple times, break with a warning
			if stuckCount > 5 {
				log.Printf("Warning: Possible infinite loop detected in text chunking at position %d. Breaking.", currentPos)
				if currentPos < textLen {
					// Add the remaining text as final chunk
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
			// Reset stuck counter if we made good progress
			stuckCount = 0
		}

		chunkNum++
	}

	return chunks, lastText
}

// extractTextWithTesseract extracts text using OCR when pdftotext fails
// Parameters:
//   - pdfPath: Path to the PDF file
//   - pageNumber: Page number to extract text from
//
// Returns:
//   - string: Extracted text
//   - error: Error if extraction fails
func (s *PDFService) extractTextWithTesseract(imgPath string) (string, error) {
	log.Println("Try extracting with tesseract, page:", imgPath)

	// Chạy OCR với Tesseract trên ảnh đã xử lý
	ocrCmd := exec.Command("tesseract",
		imgPath,
		"stdout",
		"-l", "vie+rus+eng", // Các ngôn ngữ
		"--oem", "3", // LSTM OCR Engine
		"--psm", "3", // Auto-page segmentation
		"--dpi", "300", // Match DPI with conversion
		"-c", "textord_min_linesize=2.5", // Giúp xử lý các dòng văn bản nhỏ
		"-c", "preserve_interword_spaces=1", // Bảo toàn khoảng trống giữa các từ
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

// getNumPages uses pdfinfo to get the total number of pages in a PDF file
// Parameters:
//   - pdfPath: Path to the PDF file
//
// Returns:
//   - int: Number of pages
//   - error: Error if page count cannot be determined
func getNumPages(pdfPath string) (int, error) {
	cmd := exec.Command("pdfinfo", pdfPath)
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

func (s *PDFService) cleanText(text string) string {

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
	// Apply replacements
	cleaned := text
	for old, new := range replacements {
		cleaned = strings.ReplaceAll(cleaned, old, new)
	}

	// Trim leading/trailing whitespace
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}
