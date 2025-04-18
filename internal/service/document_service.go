package service

import (
	"context"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/remiehneppo/be-task-management/internal/repository"
	"github.com/remiehneppo/be-task-management/types"
)

var _ DocumentService = (*documentService)(nil)

type DocumentService interface {
	UploadDocument(ctx context.Context, req *types.UploadDocumentRequest, fileHeader *multipart.FileHeader) (*types.UploadDocumentResponse, error)
	SearchDocument(ctx context.Context, req *types.SearchDocumentRequest) (*types.SearchDocumentResponse, error)
	AskAI(ctx context.Context, req *types.AskAIRequest) (*types.AskAIResponse, error)
	ViewDocument(ctx context.Context, req *types.ViewDocumentRequest) (*types.ViewDocumentResponse, error)
}

type documentService struct {
	allowedTypes       []string
	aiService          AIService
	ragService         RAGService
	fileService        FileService
	pdfService         PDFService
	documentVectorRepo repository.DocumentVectorRepository
}

func NewDocumentService(
	aiService AIService,
	ragService RAGService,
	fileService FileService,
	pdfService PDFService,
	documentVectorRepo repository.DocumentVectorRepository,
	allowedTypes []string,
) *documentService {
	return &documentService{
		aiService:          aiService,
		ragService:         ragService,
		fileService:        fileService,
		pdfService:         pdfService,
		documentVectorRepo: documentVectorRepo,
		allowedTypes:       allowedTypes,
	}
}

func (s *documentService) UploadDocument(ctx context.Context, req *types.UploadDocumentRequest, fileHeader *multipart.FileHeader) (*types.UploadDocumentResponse, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !s.isAllowedType(ext) {
		return nil, types.ErrUnsupportedFileType
	}
	uploadFileRes, err := s.fileService.UploadFile(ctx, types.UploadFileRequest{
		FileName:   req.Title + ext,
		FileHeader: fileHeader,
	})
	if err != nil {
		return nil, err
	}
	chunks, err := s.pdfService.ProcessPDF(uploadFileRes.FilePath)
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
