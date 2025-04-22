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
	"github.com/remiehneppo/be-task-management/utils"
)

var _ FileService = (*fileService)(nil)

type FileService interface {
	UploadFile(ctx context.Context, req types.UploadFileRequest) (*types.UploadFileResponse, error)
	GetFile(ctx context.Context, filePath string) (*os.File, error)
	// DownloadFile(fileID string) (string, error)
	// DeleteFile(fileID string) error
	// GetFileMetadata(fileID string) (*types.FileMetadata, error)
	// GetFileList(userID string, page, limit int) ([]*types.FileMetadata, error)
}

type fileService struct {
	uploadDir        string
	maxSize          int64
	fileMetadataRepo repository.FileMetadataRepository
}

func NewFileService(uploadDir string, maxSize int64, fileMetadataRepo repository.FileMetadataRepository) *fileService {
	// Ensure the upload directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			panic("failed to create upload directory: " + err.Error())
		}
	}
	return &fileService{
		uploadDir:        uploadDir,
		maxSize:          maxSize,
		fileMetadataRepo: fileMetadataRepo,
	}
}

func (f *fileService) UploadFile(ctx context.Context, req types.UploadFileRequest) (*types.UploadFileResponse, error) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(req.FileHeader.Filename))
	// Validate file size
	if req.FileHeader.Size > f.maxSize {
		return nil, types.ErrFileTooLarge
	}
	// Set default file name if not provided
	if req.FileName == "" {
		req.FileName = utils.GetFileNameWithoutExt(req.FileHeader.Filename)

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
		FileName: req.FileName,
		FilePath: filePath,
	}, nil
}

func (f *fileService) GetFile(ctx context.Context, filePath string) (*os.File, error) {
	filePath = filepath.Join(f.uploadDir, filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
