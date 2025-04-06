package types

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
