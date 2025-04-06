package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/be-task-management/internal/logger"
	"github.com/remiehneppo/be-task-management/internal/service"
	"github.com/remiehneppo/be-task-management/types"
)

type TaskHandler interface {
	GetTasksAssignedToUser(ctx *gin.Context)
	GetTasksCreatedByUser(ctx *gin.Context)
	GetTaskByID(ctx *gin.Context)
	CreateTask(ctx *gin.Context)
	UpdateTask(ctx *gin.Context)
	DeleteTask(ctx *gin.Context)
	FilterTasks(ctx *gin.Context)
	AddReportTask(ctx *gin.Context)
	UpdateReportTask(ctx *gin.Context)
	DeleteReport(ctx *gin.Context)
	FeedbackReport(ctx *gin.Context)
}

type taskHandler struct {
	taskService service.TaskService
	logger      *logger.Logger
}

func NewTaskHandler(
	taskService service.TaskService,
	logger *logger.Logger,
) TaskHandler {
	return &taskHandler{
		taskService: taskService,
		logger:      logger,
	}
}

// GetTasksAssignedToUser godoc
// @Summary Get tasks assigned to the current user
// @Description Returns a paginated list of tasks assigned to the authenticated user
// @Tags tasks
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Security BearerAuth
// @Router /tasks/assigned [get]
func (h *taskHandler) GetTasksAssignedToUser(ctx *gin.Context) {
	page, limit := GetPaginationParams(ctx)
	tasks, total, err := h.taskService.GetTasksAssignedToUser(ctx, page, limit)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.PaginatedResponse{
		Status:  true,
		Message: "Tasks retrieved successfully",
		Data: types.PaginatedData{
			Items: tasks,
			Total: total,
			Limit: limit,
			Page:  page,
		},
	}
	ctx.JSON(200, res)
}

// GetTasksCreatedByUser godoc
// @Summary Get tasks created by the current user
// @Description Returns a paginated list of tasks created by the authenticated user
// @Tags tasks
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Security BearerAuth
// @Router /tasks/created [get]
func (h *taskHandler) GetTasksCreatedByUser(ctx *gin.Context) {
	page, limit := GetPaginationParams(ctx)
	tasks, total, err := h.taskService.GetTaskCreatedByUser(ctx, page, limit)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.PaginatedResponse{
		Status:  true,
		Message: "Tasks retrieved successfully",
		Data: types.PaginatedData{
			Items: tasks,
			Total: total,
			Limit: limit,
			Page:  page,
		},
	}
	ctx.JSON(200, res)
}

// GetTaskByID godoc
// @Summary Get a task by ID
// @Description Returns a specific task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 404 {object} types.Response
// @Security BearerAuth
// @Router /tasks/{id} [get]
func (h *taskHandler) GetTaskByID(ctx *gin.Context) {
	id := ctx.Param("id")
	task, err := h.taskService.GetTaskByID(ctx, id)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.Response{
		Status:  true,
		Message: "Task retrieved successfully",
		Data:    task,
	}
	ctx.JSON(200, res)
}

// CreateTask godoc
// @Summary Create a new task
// @Description Creates a new task with the provided information
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body types.CreateTaskRequest true "Task information"
// @Success 201 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Security BearerAuth
// @Router /tasks/create [post]
func (h *taskHandler) CreateTask(ctx *gin.Context) {
	req := &types.CreateTaskRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	err := h.taskService.CreateTask(ctx, req)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.Response{
		Status:  true,
		Message: "Task created successfully",
	}
	ctx.JSON(201, res)
}

// UpdateTask godoc
// @Summary Update an existing task
// @Description Updates a task with the provided information
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body types.UpdateTaskRequest true "Updated task information"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 404 {object} types.Response
// @Security BearerAuth
// @Router /tasks/update [post]
func (h *taskHandler) UpdateTask(ctx *gin.Context) {
	req := &types.UpdateTaskRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	err := h.taskService.UpdateTask(ctx, *req)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.Response{
		Status:  true,
		Message: "Task updated successfully",
	}
	ctx.JSON(200, res)
}

// DeleteTask godoc
// @Summary Delete a task
// @Description Deletes a task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 404 {object} types.Response
// @Security BearerAuth
// @Router /tasks/delete/{id} [post]
func (h *taskHandler) DeleteTask(ctx *gin.Context) {
	id := ctx.Param("id")
	err := h.taskService.DeleteTask(ctx, id)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.Response{
		Status:  true,
		Message: "Task deleted successfully",
	}
	ctx.JSON(200, res)
}

// FilterTasks godoc
// @Summary Filter tasks based on criteria
// @Description Returns a paginated list of tasks matching the filter criteria
// @Tags tasks
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Param deadlineFrom query int64 false "Filter by deadline starting from (unix timestamp)"
// @Param deadlineTo query int64 false "Filter by deadline up to (unix timestamp)"
// @Param startFrom query int64 false "Filter by start date from (unix timestamp)"
// @Param startTo query int64 false "Filter by start date to (unix timestamp)"
// @Param reportFrom query int64 false "Filter by report date from (unix timestamp)"
// @Param reportTo query int64 false "Filter by report date to (unix timestamp)"
// @Param title query string false "Filter by task title (partial match)"
// @Param status query string false "Filter by task status"
// @Param assignee query string false "Filter tasks assigned to current user"
// @Param creator query string false "Filter tasks created by current user"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Security BearerAuth
// @Router /tasks/filter [get]
func (h *taskHandler) FilterTasks(ctx *gin.Context) {
	page, limit := GetPaginationParams(ctx)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		res := types.Response{
			Status:  false,
			Message: types.ErrInvalidCredentials.Error(),
		}
		ctx.JSON(401, res)
		return
	}
	filter := &types.TaskFilter{}
	if ctx.Query("deadlineFrom") != "" {
		deadlineFrom, err := strconv.ParseInt(ctx.Query("deadlineFrom"), 10, 64)
		if err != nil {
			res := types.Response{
				Status:  false,
				Message: "Invalid deadlineFrom parameter",
			}
			ctx.JSON(400, res)
			return
		}
		filter.DeadlineFrom = deadlineFrom
	}
	if ctx.Query("deadlineTo") != "" {
		deadlineTo, err := strconv.ParseInt(ctx.Query("deadlineTo"), 10, 64)
		if err != nil {
			res := types.Response{
				Status:  false,
				Message: "Invalid deadlineTo parameter",
			}
			ctx.JSON(400, res)
			return
		}
		filter.DeadlineTo = deadlineTo
	}
	if ctx.Query("title") != "" {
		filter.Title = ctx.Query("title")
	}

	if ctx.Query("status") != "" {
		filter.Status = ctx.Query("status")
	}

	if ctx.Query("assignee") != "" {
		filter.Assignee = userID
	}
	if ctx.Query("creator") != "" {
		filter.Creator = userID
	}
	if ctx.Query("startFrom") != "" {
		startFrom, err := strconv.ParseInt(ctx.Query("startFrom"), 10, 64)
		if err != nil {
			res := types.Response{
				Status:  false,
				Message: "Invalid startFrom parameter",
			}
			ctx.JSON(400, res)
			return
		}
		filter.StartFrom = startFrom
	}
	if ctx.Query("startTo") != "" {
		startTo, err := strconv.ParseInt(ctx.Query("startTo"), 10, 64)
		if err != nil {
			res := types.Response{
				Status:  false,
				Message: "Invalid startTo parameter",
			}
			ctx.JSON(400, res)
			return
		}
		filter.StartTo = startTo
	}
	if ctx.Query("reportFrom") != "" {
		reportFrom, err := strconv.ParseInt(ctx.Query("reportFrom"), 10, 64)
		if err != nil {
			res := types.Response{
				Status:  false,
				Message: "Invalid reportFrom parameter",
			}
			ctx.JSON(400, res)
			return
		}
		filter.ReportFrom = reportFrom
	}
	if ctx.Query("reportTo") != "" {
		reportTo, err := strconv.ParseInt(ctx.Query("reportTo"), 10, 64)
		if err != nil {
			res := types.Response{
				Status:  false,
				Message: "Invalid reportTo parameter",
			}
			ctx.JSON(400, res)
			return
		}
		filter.ReportTo = reportTo
	}

	tasks, total, err := h.taskService.FilterTasks(ctx, page, limit, *filter)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.PaginatedResponse{
		Status:  true,
		Message: "Tasks retrieved successfully",
		Data: types.PaginatedData{
			Items: tasks,
			Total: total,
			Limit: limit,
			Page:  page,
		},
	}
	ctx.JSON(200, res)
}

// AddReportTask godoc
// @Summary Add a report to a task
// @Description Creates a new report for a specific task
// @Tags reports
// @Accept json
// @Produce json
// @Param report body types.CreateReportRequest true "Report information"
// @Success 200 {object} types.Response "Report added successfully"
// @Failure 400 {object} types.Response "Invalid request format or validation error"
// @Failure 401 {object} types.Response "Unauthorized"
// @Failure 404 {object} types.Response "Task not found"
// @Security BearerAuth
// @Router /tasks/report/add [post]
func (h *taskHandler) AddReportTask(ctx *gin.Context) {
	var req types.CreateReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	err := h.taskService.AddReport(ctx, req)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.Response{
		Status:  true,
		Message: "Report added successfully",
	}
	ctx.JSON(200, res)

}

// DeleteReport godoc
// @Summary Delete a report
// @Description Deletes an existing report by ID
// @Tags reports
// @Accept json
// @Produce json
// @Param request body types.DeleteReportRequest true "Delete report request with report ID"
// @Success 200 {object} types.Response "Report deleted successfully"
// @Failure 400 {object} types.Response "Invalid request format"
// @Failure 401 {object} types.Response "Unauthorized or not the report creator"
// @Failure 404 {object} types.Response "Report not found"
// @Security BearerAuth
// @Router /tasks/report/delete [post]
func (h *taskHandler) DeleteReport(ctx *gin.Context) {
	req := &types.DeleteReportRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	err := h.taskService.DeleteReport(ctx, req)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.Response{
		Status:  true,
		Message: "Report deleted successfully",
	}
	ctx.JSON(200, res)
}

// UpdateReportTask godoc
// @Summary Update an existing report
// @Description Updates the report information for a task
// @Tags reports
// @Accept json
// @Produce json
// @Param report body types.UpdateReportRequest true "Updated report information"
// @Success 200 {object} types.Response "Report updated successfully"
// @Failure 400 {object} types.Response "Invalid request format or validation error"
// @Failure 401 {object} types.Response "Unauthorized or not the report creator"
// @Failure 404 {object} types.Response "Report not found"
// @Security BearerAuth
// @Router /tasks/report/update [post]
func (h *taskHandler) UpdateReportTask(ctx *gin.Context) {
	req := &types.UpdateReportRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	err := h.taskService.UpdateReport(ctx, req)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.Response{
		Status:  true,
		Message: "Report updated successfully",
	}
	ctx.JSON(200, res)
}

// FeedbackReport godoc
// @Summary Provide feedback on a report
// @Description Adds feedback to a specific report
// @Tags reports
// @Accept json
// @Produce json
// @Param feedback body types.FeedbackRequest true "Feedback information"
// @Success 200 {object} types.Response "Feedback added successfully"
// @Failure 400 {object} types.Response "Invalid request format or validation error"
// @Failure 401 {object} types.Response "Unauthorized"
// @Failure 404 {object} types.Response "Report not found"
// @Security BearerAuth
// @Router /tasks/report/feedback [post]
func (h *taskHandler) FeedbackReport(ctx *gin.Context) {
	req := &types.FeedbackRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	err := h.taskService.FeedbackReport(ctx, req)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	res := types.Response{
		Status:  true,
		Message: "Feedback added successfully",
	}
	ctx.JSON(200, res)
}

func GetPaginationParams(c *gin.Context) (page int64, limit int64) {
	// Default values
	page = 1
	limit = 10

	// Try to parse page parameter
	pageStr := c.DefaultQuery("page", "1")
	if pageVal, err := strconv.ParseInt(pageStr, 10, 64); err == nil && pageVal > 0 {
		page = pageVal
	}

	// Try to parse limit parameter
	limitStr := c.DefaultQuery("limit", "10")
	if limitVal, err := strconv.ParseInt(limitStr, 10, 64); err == nil && limitVal > 0 {
		limit = limitVal
	}

	return page, limit
}
