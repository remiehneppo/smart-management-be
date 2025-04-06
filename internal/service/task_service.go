package service

import (
	"context"
	"time"

	"github.com/remiehneppo/be-task-management/internal/repository"
	"github.com/remiehneppo/be-task-management/types"
)

type TaskService interface {
	GetTasksAssignedToUser(ctx context.Context, page, limit int64) (items []*types.TaskResponse, total int64, err error)
	GetTaskCreatedByUser(ctx context.Context, page, limit int64) (items []*types.TaskResponse, total int64, err error)
	GetTaskByID(ctx context.Context, id string) (*types.TaskResponse, error)
	CreateTask(ctx context.Context, task *types.CreateTaskRequest) error
	UpdateTask(ctx context.Context, req types.UpdateTaskRequest) error
	DeleteTask(ctx context.Context, id string) error
	FilterTasks(ctx context.Context, page, limit int64, filter types.TaskFilter) (items []*types.TaskResponse, total int64, err error)
	AddReport(ctx context.Context, req types.CreateReportRequest) error
	DeleteReport(ctx context.Context, req *types.DeleteReportRequest) error
	UpdateReport(ctx context.Context, req *types.UpdateReportRequest) error
	FeedbackReport(ctx context.Context, req *types.FeedbackRequest) error
}

type taskService struct {
	taskRepo   repository.TaskRepository
	userRepo   repository.UserRepository
	reportRepo repository.ReportRepository
}

func NewTaskService(taskRepo repository.TaskRepository, reportRepo repository.ReportRepository, userRepo repository.UserRepository) TaskService {
	return &taskService{
		taskRepo:   taskRepo,
		userRepo:   userRepo,
		reportRepo: reportRepo,
	}
}

func (s *taskService) GetTasksAssignedToUser(ctx context.Context, page, limit int64) (items []*types.TaskResponse, total int64, err error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, 0, types.ErrInvalidCredentials
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	tasks, total, err := s.taskRepo.PaginateWithFilter(ctx, page, limit, types.TaskFilter{
		Assignee:  userID,
		Workspace: user.Workspace,
	})
	if err != nil {
		return nil, 0, err
	}
	userIDs := make([]string, 0)
	userIDsMap := make(map[string]bool)
	for _, task := range tasks {
		if userIDsMap[task.Creator] {
			continue
		}
		userIDs = append(userIDs, task.Creator)
		userIDsMap[task.Creator] = true
	}
	usersMap, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	taskIDs := make([]string, 0)
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}
	reportsMap, err := s.reportRepo.FindByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, 0, err
	}
	tasksRes := make([]*types.TaskResponse, 0)
	for _, task := range tasks {
		creator, ok := usersMap[task.Creator]
		if !ok {
			return nil, 0, types.ErrInvalidUser
		}
		taskRes := s.convertTaskToTaskRes(task, creator.FullName, user.FullName)
		reports := make([]*types.ReportResponse, 0)
		for _, report := range reportsMap[task.ID] {
			reportRes := &types.ReportResponse{
				ID:       report.ID,
				Creator:  report.Creator,
				Report:   report.Report,
				Feedback: report.Feedback,
			}
			reports = append(reports, reportRes)
		}
		taskRes.Reports = reports
		tasksRes = append(tasksRes, taskRes)
	}

	return tasksRes, total, nil
}

func (s *taskService) GetTaskCreatedByUser(ctx context.Context, page, limit int64) (items []*types.TaskResponse, total int64, err error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, 0, types.ErrInvalidCredentials
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	tasks, total, err := s.taskRepo.PaginateWithFilter(ctx, page, limit, types.TaskFilter{
		Creator:   userID,
		Workspace: user.Workspace,
	})
	if err != nil {
		return nil, 0, err
	}
	assigneeIds := make([]string, 0)
	assigneeIdsMap := make(map[string]bool)
	for _, task := range tasks {
		if assigneeIdsMap[task.Assignee] {
			continue
		}
		assigneeIds = append(assigneeIds, task.Assignee)
		assigneeIdsMap[task.Assignee] = true
	}
	usersMap, err := s.userRepo.FindByIDs(ctx, assigneeIds)
	if err != nil {
		return nil, 0, err
	}
	taskIDs := make([]string, 0)
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}
	reportsMap, err := s.reportRepo.FindByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, 0, err
	}
	tasksRes := make([]*types.TaskResponse, 0)
	for _, task := range tasks {
		assgnee, ok := usersMap[task.Assignee]
		taskRes := s.convertTaskToTaskRes(task, user.FullName, assgnee.FullName)
		if !ok {
			return nil, 0, types.ErrInvalidUser
		}
		reports := make([]*types.ReportResponse, 0)
		for _, report := range reportsMap[task.ID] {
			reportRes := &types.ReportResponse{
				ID:       report.ID,
				Creator:  report.Creator,
				Report:   report.Report,
				Feedback: report.Feedback,
			}
			reports = append(reports, reportRes)
		}
		taskRes.Reports = reports
		tasksRes = append(tasksRes, taskRes)
	}
	return tasksRes, total, nil
}

func (s *taskService) GetTaskByID(ctx context.Context, id string) (*types.TaskResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, types.ErrInvalidCredentials
	}
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if task.Workspace != user.Workspace {
		return nil, types.ErrTaskNotInWorkspace
	}
	users, err := s.userRepo.FindByIDs(ctx, []string{task.Creator, task.Assignee})
	if err != nil {
		return nil, err
	}
	if len(users) != 2 {
		return nil, types.ErrInvalidUser
	}
	if _, ok := users[task.Creator]; !ok {
		return nil, types.ErrInvalidUser
	}
	if _, ok := users[task.Assignee]; !ok {
		return nil, types.ErrInvalidUser
	}
	taskRes := s.convertTaskToTaskRes(task, users[task.Creator].FullName, users[task.Assignee].FullName)
	taskRes.Creator = users[task.Creator].FullName
	taskRes.Assignee = users[task.Assignee].FullName
	reports, err := s.reportRepo.FindByTaskID(ctx, task.ID)
	if err != nil {
		return nil, err
	}
	reportsRes := make([]*types.ReportResponse, 0)
	for _, report := range reports {
		reportRes := &types.ReportResponse{
			ID:       report.ID,
			Creator:  report.Creator,
			Report:   report.Report,
			Feedback: report.Feedback,
		}
		reportsRes = append(reportsRes, reportRes)
	}
	taskRes.Reports = reportsRes
	return taskRes, nil
}

func (s *taskService) CreateTask(ctx context.Context, req *types.CreateTaskRequest) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return types.ErrInvalidCredentials
	}
	creator, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	assignee, err := s.userRepo.FindByID(ctx, req.Assignee)
	if err != nil {
		return err
	}
	err = s.validateCreateQuestPermission(ctx, creator, assignee)
	if err != nil {
		return err
	}
	task := &types.Task{
		Title:       req.Title,
		Description: req.Description,
		Workspace:   creator.Workspace,
		StartAt:     req.StartAt,
		Deadline:    req.Deadline,
		Creator:     userID,
		Assignee:    req.Assignee,
		Status:      types.TASK_STATUS_OPEN,
		CreateAt:    time.Now().Unix(),
		UpdateAt:    time.Now().Unix(),
		Progress:    0,
	}
	err = s.taskRepo.Save(ctx, task)
	if err != nil {
		return err
	}
	return nil
}

func (s *taskService) UpdateTask(ctx context.Context, req types.UpdateTaskRequest) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return types.ErrInvalidCredentials
	}
	taskInDB, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return err
	}
	if taskInDB.Creator != userID || taskInDB.Assignee != userID {
		return types.ErrTaskNotCreatorOrAssignee
	}
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return err
	}
	creator, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Assignee != "" {
		assignee, err := s.userRepo.FindByID(ctx, req.Assignee)
		if err != nil {
			return err
		}
		err = s.validateCreateQuestPermission(ctx, creator, assignee)
		if err != nil {
			return err
		}
		task.Assignee = req.Assignee
	}
	if req.Deadline != 0 {
		task.Deadline = req.Deadline
	}
	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Progress != 0 {
		if req.Progress < 0 || req.Progress > 100 {
			return types.ErrInvalidProgress
		}
		task.Progress = req.Progress
	}
	if req.StartAt != 0 {
		task.StartAt = req.StartAt
	}
	task.ID = ""
	err = s.taskRepo.Update(ctx, req.TaskID, task)
	if err != nil {
		return err
	}
	return nil
}

func (s *taskService) DeleteTask(ctx context.Context, id string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return types.ErrInvalidCredentials
	}
	taskInDB, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if taskInDB.Creator != userID {
		return types.ErrTaskNotCreator
	}
	err = s.taskRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	err = s.reportRepo.DeleteByTaskID(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *taskService) FilterTasks(ctx context.Context, page, limit int64, filter types.TaskFilter) (items []*types.TaskResponse, total int64, err error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, 0, types.ErrInvalidCredentials
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	filter.Workspace = user.Workspace
	tasks, total, err := s.taskRepo.PaginateWithFilter(ctx, page, limit, filter)
	if err != nil {
		return nil, 0, err
	}
	userIDs := make([]string, 0)
	userIDsMap := make(map[string]bool)
	for _, task := range tasks {
		if !userIDsMap[task.Creator] {
			userIDs = append(userIDs, task.Creator)
			userIDsMap[task.Creator] = true
		}
		if !userIDsMap[task.Assignee] {
			userIDs = append(userIDs, task.Assignee)
			userIDsMap[task.Assignee] = true
		}
	}
	usersMap, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}
	taskIDs := make([]string, 0)
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}
	reportsMap, err := s.reportRepo.FindByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, 0, err
	}
	tasksRes := make([]*types.TaskResponse, 0)
	for _, task := range tasks {
		creator, ok := usersMap[task.Creator]
		if !ok {
			return nil, 0, types.ErrInvalidUser
		}
		assignee, ok := usersMap[task.Assignee]
		if !ok {
			return nil, 0, types.ErrInvalidUser
		}
		taskRes := s.convertTaskToTaskRes(task, creator.FullName, assignee.FullName)
		reports := make([]*types.ReportResponse, 0)
		for _, report := range reportsMap[task.ID] {
			reportRes := &types.ReportResponse{
				ID:       report.ID,
				Creator:  report.Creator,
				Report:   report.Report,
				Feedback: report.Feedback,
			}
			reports = append(reports, reportRes)
		}
		taskRes.Reports = reports
		tasksRes = append(tasksRes, taskRes)
	}
	return tasksRes, total, nil
}

func (s *taskService) AddReport(ctx context.Context, req types.CreateReportRequest) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return types.ErrInvalidCredentials
	}
	taskInDB, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return err
	}
	if taskInDB.Assignee != userID {
		return types.ErrTaskNotAssignee
	}
	reportObj := &types.Report{
		TaskID:     req.TaskID,
		Creator:    userID,
		Report:     req.Report,
		CreatedAt:  time.Now().Unix(),
		ReportFile: req.ReportFile,
	}
	err = s.reportRepo.Save(ctx, reportObj)
	if err != nil {
		return err
	}
	return nil
}
func (s *taskService) DeleteReport(ctx context.Context, req *types.DeleteReportRequest) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return types.ErrInvalidCredentials
	}
	reportInDB, err := s.reportRepo.FindByID(ctx, req.ReportID)
	if err != nil {
		return err
	}
	if reportInDB.Creator != userID {
		return types.ErrReportNotCreator
	}
	err = s.reportRepo.Delete(ctx, req.ReportID)
	if err != nil {
		return err
	}
	return nil
}

func (s *taskService) UpdateReport(ctx context.Context, req *types.UpdateReportRequest) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return types.ErrInvalidCredentials
	}
	reportInDB, err := s.reportRepo.FindByID(ctx, req.ReportID)
	if err != nil {
		return err
	}
	if reportInDB.Creator != userID {
		return types.ErrReportNotCreator
	}
	reportInDB.Report = req.Report
	reportInDB.ID = ""
	err = s.reportRepo.Update(ctx, req.ReportID, reportInDB)
	if err != nil {
		return err
	}
	return nil
}

func (s *taskService) validateCreateQuestPermission(ctx context.Context, creator, assignee *types.User) error {
	if creator.Workspace != assignee.Workspace {
		return types.ErrQuestAssignNotWorkspaceMember
	}
	if types.MAPPING_ROLE_TO_MANAGEMENT_LEVEL[creator.WorkspaceRole] < types.MAPPING_ROLE_TO_MANAGEMENT_LEVEL[assignee.WorkspaceRole] {
		return types.ErrQuestAssignNoPermission
	}
	return nil
}

func (s *taskService) ReportTaskById(ctx context.Context, id string, report string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return types.ErrInvalidCredentials
	}
	taskInDB, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if taskInDB.Assignee != userID {
		return types.ErrTaskNotAssignee
	}
	reportObj := &types.Report{
		TaskID:    id,
		Creator:   userID,
		Report:    report,
		CreatedAt: time.Now().Unix(),
	}
	err = s.reportRepo.Save(ctx, reportObj)
	if err != nil {
		return err
	}
	return nil
}

func (s *taskService) FeedbackReport(ctx context.Context, req *types.FeedbackRequest) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return types.ErrInvalidCredentials
	}
	reportInDB, err := s.reportRepo.FindByID(ctx, req.ReportID)
	if err != nil {
		return err
	}
	if reportInDB.Creator == userID {
		return types.ErrReportNotCreator
	}
	reportInDB.Feedback = req.Feedback
	reportInDB.ID = ""
	err = s.reportRepo.Update(ctx, req.ReportID, reportInDB)
	if err != nil {
		return err
	}
	return nil
}

func (s *taskService) convertTaskToTaskRes(task *types.Task, creator, assignee string) *types.TaskResponse {
	taskRes := &types.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Workspace:   task.Workspace,
		Creator:     creator,
		Deadline:    task.Deadline,
		Assignee:    assignee,
		Status:      task.Status,
		CreateAt:    task.CreateAt,
		UpdateAt:    task.UpdateAt,
	}
	return taskRes
}
