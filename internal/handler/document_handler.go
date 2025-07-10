package handler

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/be-task-management/internal/service"
	"github.com/remiehneppo/be-task-management/types"
)

var _ DocumentHandler = (*documentHandler)(nil)

type DocumentHandler interface {
	UploadPDF(ctx *gin.Context)
	BatchUploadPDFAsync(ctx *gin.Context)
	SearchDocument(ctx *gin.Context)
	AskAI(ctx *gin.Context)
	ViewDocument(ctx *gin.Context)
	DemoloadText(ctx *gin.Context)
}

type documentHandler struct {
	documentService service.DocumentService
}

func NewDocumentHandler(documentService service.DocumentService) *documentHandler {
	return &documentHandler{
		documentService: documentService,
	}
}

// UploadPDF godoc
// @Summary Upload a PDF document
// @Description Uploads a PDF file and processes it for further use
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF file to upload"
// @Param metadata formData string true "Document metadata in JSON format"
// @Success 200 {object} types.Response "File uploaded successfully"
// @Failure 400 {object} types.Response "File upload error or invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /documents/upload [post]
func (h *documentHandler) UploadPDF(ctx *gin.Context) {
	// Handle file upload
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "File upload error",
		})
		return
	}
	metadata := ctx.PostForm("metadata")
	if metadata == "" {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request: missing metadata",
		})
		return
	}
	var req types.UploadDocumentRequest
	if err := json.Unmarshal([]byte(metadata), &req); err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid metadata format",
		})
		return
	}
	_, err = h.documentService.UploadDocument(ctx, &req, file)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: err.Error(),
		})
		return
	}
	ctx.JSON(200, types.Response{
		Status:  true,
		Message: "File uploaded successfully",
	})

}

// SearchDocument godoc
// @Summary Search documents
// @Description Searches for documents based on the provided query
// @Tags documents
// @Accept json
// @Produce json
// @Param query body types.SearchDocumentRequest true "Search query"
// @Success 200 {object} types.Response{data=[]types.SearchDocumentResponse} "Search results"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /documents/search [post]
func (h *documentHandler) SearchDocument(ctx *gin.Context) {
	var req types.SearchDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request",
		})
		return
	}
	res, err := h.documentService.SearchDocument(ctx, &req)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Internal server error",
		})
		return
	}
	ctx.JSON(200, types.Response{
		Status:  true,
		Message: "Search results",
		Data:    res,
	})
}

// AskAI godoc
// @Summary Ask AI a question
// @Description Sends a question to the AI and retrieves a response
// @Tags documents
// @Accept json
// @Produce json
// @Param question body types.AskAIRequest true "Question for the AI"
// @Success 200 {object} types.Response{data=types.AskAIResponse} "AI response"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /documents/ask-ai [post]
func (h *documentHandler) AskAI(ctx *gin.Context) {
	var req types.AskAIRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request",
		})
		return
	}
	res, err := h.documentService.AskAI(ctx, &req)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Internal server error",
		})
		return
	}
	ctx.JSON(200, types.Response{
		Status:  true,
		Message: "AI response",
		Data:    res,
	})
}

// ViewDocument godoc
// @Summary View a PDF document
// @Description Streams a PDF document to the client for viewing in the browser
// @Tags documents
// @Accept json
// @Produce application/pdf
// @Param path query string true "Path to the PDF document"
// @Success 200 {file} file "PDF document streamed successfully"
// @Failure 400 {object} types.Response "Invalid request: missing document path"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /documents/view [get]
func (h *documentHandler) ViewDocument(ctx *gin.Context) {
	documentPath := ctx.Query("path")
	if documentPath == "" {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request: missing document path",
		})
		return
	}
	documentRes, err := h.documentService.ViewDocument(ctx, &types.ViewDocumentRequest{
		FilePath: documentPath,
	})
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Internal server error",
		})
		return
	}
	defer documentRes.Document.Close()
	ctx.Header("Content-Disposition", "inline")
	ctx.Header("Content-Type", "application/pdf")

	_, err = io.Copy(ctx.Writer, documentRes.Document)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Internal server error",
		})
		return
	}

}

// DemoloadText godoc
// @Summary Demo load text from a PDF document
// @Description Loads text from a PDF document for demonstration purposes
// @Tags documents
// @Accept json
// @Produce json
// @Param file formData file true "PDF file to load text from"
// @Param metadata formData string true "Document metadata in JSON format"
// @Success 200 {object} types.Response{data=types.DemoGetTextResponse} "Text loaded successfully"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /documents/demo-load-text [post]
func (h *documentHandler) DemoloadText(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "File upload error",
		})
		return
	}
	metadata := ctx.PostForm("metadata")
	if metadata == "" {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request: missing metadata",
		})
		return
	}
	var req types.DemoGetTextRequest
	if err := json.Unmarshal([]byte(metadata), &req); err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid metadata format",
		})
		return
	}
	res, err := h.documentService.DemoGetText(ctx, &req, file)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Internal server error",
		})
		return
	}
	ctx.JSON(200, types.Response{
		Status:  true,
		Message: "Text loaded successfully",
		Data:    res,
	})
}

// BatchUploadPDFAsync godoc
// @Summary Upload multiple PDF documents asynchronously
// @Description Uploads multiple PDF files and processes them asynchronously in the background
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "PDF files to upload (multiple files with same field name)"
// @Param metadata formData string true "Document metadata in JSON format"
// @Success 200 {object} types.Response{data=types.BatchUploadDocumentResponse} "Files uploaded successfully"
// @Failure 400 {object} types.Response "File upload error or invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /documents/batch-upload [post]
func (h *documentHandler) BatchUploadPDFAsync(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request: unable to parse multipart form",
		})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request: no files provided",
		})
		return
	}

	var req types.BatchUploadDocumentRequest
	metadata := ctx.PostForm("metadata")
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &req); err != nil {
			ctx.JSON(400, types.Response{
				Status:  false,
				Message: "Invalid metadata format",
			})
			return
		}
	} else {
		req = types.BatchUploadDocumentRequest{}
	}
	req.Files = files

	res, err := h.documentService.BatchUploadDocumentAsync(ctx, &req)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Internal server error",
		})
		return
	}
	ctx.JSON(200, types.Response{
		Status:  true,
		Message: "Files uploaded successfully",
		Data:    res,
	})
}
