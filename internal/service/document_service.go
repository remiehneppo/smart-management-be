package service

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/remiehneppo/be-task-management/internal/repository"
	"github.com/remiehneppo/be-task-management/internal/worker"
	"github.com/remiehneppo/be-task-management/types"
	"github.com/remiehneppo/be-task-management/utils"
	"github.com/sirupsen/logrus"
)

var _ DocumentService = (*documentService)(nil)

type DocumentService interface {
	UploadDocument(ctx context.Context, req *types.UploadDocumentRequest, fileHeader *multipart.FileHeader) (*types.UploadDocumentResponse, error)
	BatchUploadDocumentAsync(ctx context.Context, req *types.BatchUploadDocumentRequest) (*types.BatchUploadDocumentResponse, error)
	SearchDocument(ctx context.Context, req *types.SearchDocumentRequest) (*types.SearchDocumentResponse, error)
	AskAI(ctx context.Context, req *types.AskAIRequest) (*types.AskAIResponse, error)
	ViewDocument(ctx context.Context, req *types.ViewDocumentRequest) (*types.ViewDocumentResponse, error)
	DemoGetText(ctx context.Context, req *types.DemoGetTextRequest, fileHeader *multipart.FileHeader) (*types.DemoGetTextResponse, error)
	ProcessDocumentJob() worker.Do
}

type documentService struct {
	allowedTypes        []string
	aiService           AIService
	ragService          RAGService
	fileService         FileService
	pdfService          PDFService
	documentVectorRepo  repository.DocumentVectorRepository
	pendingDocumentRepo repository.PendingDocumentRepository
	lockService         LockService
}

func NewDocumentService(
	aiService AIService,
	ragService RAGService,
	fileService FileService,
	pdfService PDFService,
	documentVectorRepo repository.DocumentVectorRepository,
	pendingDocumentRepo repository.PendingDocumentRepository,
	allowedTypes []string,
	lockService LockService,
) DocumentService {
	return &documentService{
		aiService:           aiService,
		ragService:          ragService,
		fileService:         fileService,
		pdfService:          pdfService,
		documentVectorRepo:  documentVectorRepo,
		pendingDocumentRepo: pendingDocumentRepo,
		allowedTypes:        allowedTypes,
		lockService:         lockService,
	}
}

func (s *documentService) UploadDocument(ctx context.Context, req *types.UploadDocumentRequest, fileHeader *multipart.FileHeader) (*types.UploadDocumentResponse, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !s.isAllowedType(ext) {
		return nil, types.ErrUnsupportedFileType
	}
	if req.Title == "" {
		req.Title = utils.GetFileNameWithoutExt(fileHeader.Filename)
	}
	uploadFileRes, err := s.fileService.UploadFile(ctx, types.UploadFileRequest{
		FileName:   req.Title + ext,
		FileHeader: fileHeader,
	})
	if err != nil {
		return nil, err
	}
	chunks, err := s.pdfService.ProcessPDF(&types.ProcessPDFRequest{
		ToolUse:  req.ToolUse,
		FilePath: uploadFileRes.FilePath,
	})
	if err != nil {
		return nil, err
	}

	// Process the document and get the metadata
	if err := s.documentVectorRepo.SaveBatchDocumentVector(
		ctx,
		&types.DocumentMetadata{
			Title:    req.Title,
			Tags:     req.Tags,
			FilePath: uploadFileRes.FilePath,
		},
		chunks,
	); err != nil {
		return nil, err
	}
	return &types.UploadDocumentResponse{
		FilePath: uploadFileRes.FilePath,
	}, nil
}

func (s *documentService) BatchUploadDocumentAsync(ctx context.Context, req *types.BatchUploadDocumentRequest) (*types.BatchUploadDocumentResponse, error) {

	uploadStates := make([]*types.UploadStatus, 0)

	for _, fileHeader := range req.Files {
		uploadReq := types.UploadFileRequest{
			FileName:   fileHeader.Filename,
			FileHeader: fileHeader,
		}
		uploadRes, err := s.fileService.UploadFile(
			ctx, uploadReq,
		)
		if err != nil {
			uploadStates = append(uploadStates, &types.UploadStatus{
				FileName: uploadReq.FileName,
				Status:   false,
				Message:  err.Error(),
			})
			continue
		}
		uploadStates = append(uploadStates, &types.UploadStatus{
			FileName: uploadReq.FileName,
			Status:   true,
		})
		pendingDocument := &types.PendingDocument{
			DocumentPath: uploadRes.FilePath,
			DocumentName: uploadReq.FileName,
			Tags:         req.Tags,
			ToolUse:      req.ToolUse,
			CreatedAt:    time.Now().Unix(),
		}
		if err := s.pendingDocumentRepo.Save(ctx, pendingDocument); err != nil {
			uploadStates = append(uploadStates, &types.UploadStatus{
				FileName: uploadReq.FileName,
				Status:   false,
				Message:  err.Error(),
			})
			continue
		}
	}

	return &types.BatchUploadDocumentResponse{
		UploadStates: uploadStates,
	}, nil
}

func (s *documentService) ProcessDocumentJob() worker.Do {
	return func() error {
		logrus.Info("Processing pending documents...")
		ctx := context.Background()
		pendingDocuments, _, err := s.pendingDocumentRepo.FindAll(ctx, 0, 100)
		if err != nil {
			return err
		}

		for _, pendingDocument := range pendingDocuments {
			ok, _ := s.lockService.Lock(ctx, pendingDocument.DocumentName, 20*time.Minute)
			if !ok {
				continue
			}
			chunks, err := s.pdfService.ProcessPDF(&types.ProcessPDFRequest{
				ToolUse:  pendingDocument.ToolUse,
				FilePath: pendingDocument.DocumentPath,
			})
			if err != nil {
				continue
			}
			// remove documents with same name in the vector db
			s.documentVectorRepo.RemoveDocuments(ctx, &types.DocumentMetadata{})
			if err := s.documentVectorRepo.SaveBatchDocumentVector(
				context.Background(),
				&types.DocumentMetadata{
					Title:    pendingDocument.DocumentName,
					Tags:     pendingDocument.Tags,
					FilePath: pendingDocument.DocumentPath,
				},
				chunks,
			); err != nil {
				continue
			}
			// remove the pending document
			if err := s.pendingDocumentRepo.Remove(ctx, pendingDocument.ID); err != nil {
				continue
			}
			// unlock the document
			if err := s.lockService.ReleaseLock(ctx, pendingDocument.DocumentName); err != nil {
				continue
			}

		}

		return nil
	}
}

func (s *documentService) SearchDocument(ctx context.Context, req *types.SearchDocumentRequest) (*types.SearchDocumentResponse, error) {
	queries := s.getQueries(req.Query)
	chunks, err := s.documentVectorRepo.SearchDocumentVector(
		ctx,
		&types.DocumentMetadata{
			Title: req.Title,
			Tags:  req.Tags,
		},
		queries,
		req.Limit,
	)
	if err != nil {
		return nil, err
	}
	return &types.SearchDocumentResponse{
		Chunks: chunks,
	}, nil
}

func (s *documentService) AskAI(ctx context.Context, req *types.AskAIRequest) (*types.AskAIResponse, error) {
	queries := s.getQueries(req.Query)
	chunks, err := s.documentVectorRepo.SearchDocumentVector(
		ctx,
		&types.DocumentMetadata{
			Title: req.Title,
			Tags:  req.Tags,
		},
		queries,
		req.Limit,
	)
	if err != nil {
		return nil, err
	}
	answer, err := s.ragService.AskAI(ctx, req.Question, chunks)
	if err != nil {
		return nil, err
	}
	return &types.AskAIResponse{
		Answer: answer,
		Chunks: chunks,
	}, nil
}

func (s *documentService) getQueries(query string) []string {
	return []string{query}
}

func (s *documentService) isAllowedType(ext string) bool {
	for _, allowedType := range s.allowedTypes {
		if strings.EqualFold(ext, allowedType) {
			return true
		}
	}
	return false
}

func (s *documentService) ViewDocument(ctx context.Context, req *types.ViewDocumentRequest) (*types.ViewDocumentResponse, error) {
	file, err := s.fileService.GetFile(ctx, req.FilePath)
	if err != nil {
		return nil, err
	}

	return &types.ViewDocumentResponse{
		Document: file,
	}, nil
}

func (s *documentService) DemoGetText(ctx context.Context, req *types.DemoGetTextRequest, fileHeader *multipart.FileHeader) (*types.DemoGetTextResponse, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !s.isAllowedType(ext) {
		return nil, types.ErrUnsupportedFileType
	}
	tempDir := filepath.Join("temp", "documents")
	// create temp dir if not exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
			return nil, err
		}
	}
	tempFilePath := filepath.Join(tempDir, fileHeader.Filename)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return nil, err
	}
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	defer tempFile.Close()
	defer os.Remove(tempFilePath)
	if _, err := io.Copy(tempFile, src); err != nil {
		return nil, err
	}
	// Process the document and get the text
	pages, err := s.pdfService.ExtractPageContent(&types.ExtractPageContentRequest{
		ToolUse:  req.ToolUse,
		FilePath: tempFilePath,
		FromPage: req.FromPage,
		ToPage:   req.ToPage,
	})

	if err != nil {
		return nil, err
	}
	return &types.DemoGetTextResponse{
		Pages: pages,
	}, nil

}
