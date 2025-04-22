package types

import "mime/multipart"

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type PaginatedRequest struct {
	Page  int64 `json:"page" binding:"required"`
	Limit int64 `json:"limit" binding:"required"`
}

type UpdateTaskRequest struct {
	TaskID      string `json:"task_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartAt     int64  `json:"start_at"`
	Deadline    int64  `json:"deadline"`
	Assignee    string `json:"assignee"`
	Progress    int    `json:"progress"`
	Status      string `json:"status"`
}

type GetTasksAssignedToUserRequest struct {
}

type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	StartAt     int64  `json:"start_at" binding:"required"`
	Deadline    int64  `json:"deadline" binding:"required"`
	Assignee    string `json:"assignee" binding:"required"`
}

type UpdateReportRequest struct {
	ReportID string `json:"report_id" binding:"required"`
	Report   string `json:"report" binding:"required"`
}

type DeleteReportRequest struct {
	ReportID string `json:"report_id" binding:"required"`
}

type CreateReportRequest struct {
	TaskID     string `json:"task_id" binding:"required"`
	Report     string `json:"report" binding:"required"`
	ReportFile string `json:"report_file"`
}

type FeedbackRequest struct {
	ReportID string `json:"report_id" binding:"required"`
	Feedback string `json:"feedback" binding:"required"`
}

type UploadRequest struct {
	Title  string   `json:"title" binding:"required"`
	Source string   `json:"source"`
	Tags   []string `json:"tags" binding:"required"`
}

type ChatRequest struct {
	ChatId string `json:"chat_id" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
}

type ChatStatelessRequest struct {
	Messages []Message `json:"messages" binding:"required"`
}

type PaginateMessagesRequest struct {
	ChatId string `json:"chat_id" binding:"required"`
	Page   int64  `json:"page" binding:"required"`
	Limit  int64  `json:"limit" binding:"required"`
}

type UploadFileRequest struct {
	FileName   string                `json:"file_name" binding:"required"`
	FileHeader *multipart.FileHeader `json:"file" binding:"required"`
}

type UploadDocumentRequest struct {
	Title   string   `json:"title" binding:"required"`
	Tags    []string `json:"tags" binding:"required"`
	ToolUse string   `json:"tool_use" binding:"required"`
}

type SearchDocumentRequest struct {
	Title string   `json:"title,omitempty"`
	Tags  []string `json:"tags,omitempty"`
	Query string   `json:"query" binding:"required"`
	Limit int      `json:"limit" binding:"required"`
}

type AskAIRequest struct {
	Question string   `json:"question" binding:"required"`
	Title    string   `json:"title,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Query    string   `json:"query" binding:"required"`
	Limit    int      `json:"limit" binding:"required"`
}

type ViewDocumentRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}

type DemoGetTextRequest struct {
	ToolUse  string `json:"tool_use" binding:"required"`
	FromPage int    `json:"from_page" binding:"required"`
	ToPage   int    `json:"to_page" binding:"required"`
}

type ProcessPDFRequest struct {
	ToolUse  string `json:"tool_use" binding:"required"`
	FilePath string `json:"file_path" binding:"required"`
}

type ExtractPageContentRequest struct {
	ToolUse  string `json:"tool_use" binding:"required"`
	FilePath string `json:"file_path" binding:"required"`
	FromPage int    `json:"from_page" binding:"required"`
	ToPage   int    `json:"to_page" binding:"required"`
}
