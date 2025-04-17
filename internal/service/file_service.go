package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/remiehneppo/be-task-management/internal/repository"
	"github.com/remiehneppo/be-task-management/types"
)

var _ FileService = (*fileService)(nil)

type FileService interface {
	UploadFile(ctx context.Context, req types.UploadFileRequest) (*types.UploadFileResponse, error)
	// DownloadFile(fileID string) (string, error)
	// DeleteFile(fileID string) error
	// GetFileMetadata(fileID string) (*types.FileMetadata, error)
	// GetFileList(userID string, page, limit int) ([]*types.FileMetadata, error)
}

type fileService struct {
	uploadDir        string
	maxSize          int64
	allowedTypes     []string
	fileMetadataRepo repository.FileMetadataRepository
}

func NewFileService(uploadDir string) *fileService {
	return &fileService{
		uploadDir: uploadDir,
	}
}

func (f *fileService) UploadFile(ctx context.Context, req types.UploadFileRequest) (*types.UploadFileResponse, error) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(req.FileHeader.Filename))
	if !f.isAllowedType(ext) {
		return nil, types.ErrUnsupportedFileType
	}
	// Validate file size
	if req.FileHeader.Size > f.maxSize {
		return nil, types.ErrFileTooLarge
	}
	// Set default file name if not provided
	if req.FileName == "" {
		req.FileName = req.FileHeader.Filename
	}
	// Ensure the file name has the correct extension
	if filepath.Ext(req.FileName) != ext {
		req.FileName += ext
	}
	// Construct the file path
	filePath := filepath.Join(f.uploadDir, req.FileName)

	src, err := req.FileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}
	fileMetadata := types.FileMetadata{
		FileName:  req.FileHeader.Filename,
		FileSize:  req.FileHeader.Size,
		FileType:  ext,
		FilePath:  filePath,
		CreatedAt: time.Now().Unix(),
	}
	if err := f.fileMetadataRepo.CreateFileMetadata(ctx, &fileMetadata); err != nil {
		return nil, err
	}
	return &types.UploadFileResponse{
		FileID:   fileMetadata.ID,
		FileName: req.FileName,
		FilePath: filePath,
	}, nil
}

func (f *fileService) isAllowedType(ext string) bool {
	for _, allowedType := range f.allowedTypes {
		if strings.EqualFold(ext, allowedType) {
			return true
		}
	}
	return false
}
