package repository

import (
	"context"

	"github.com/remiehneppo/be-task-management/internal/database"
	"github.com/remiehneppo/be-task-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const ReportCollection = "reports"

var defaultReportSort = bson.M{"created_at": -1} // sort by created_at descending

type ReportRepository interface {
	Save(ctx context.Context, report *types.Report) error
	FindByID(ctx context.Context, id string) (*types.Report, error)
	FindByTaskID(ctx context.Context, taskID string) ([]*types.Report, error)
	FindByTaskIDs(ctx context.Context, taskIDs []string) (map[string][]*types.Report, error)
	FilterReports(ctx context.Context, page int64, limit int64, filter types.ReportFilter) ([]*types.Report, int64, error)
	Count(ctx context.Context) (int64, error)
	CountWithFilter(ctx context.Context, filter types.ReportFilter) (int64, error)
	Delete(ctx context.Context, id string) error
	DeleteByTaskID(ctx context.Context, taskID string) error
	Update(ctx context.Context, id string, report *types.Report) error
	FindAll(ctx context.Context) ([]*types.Report, error)
}

type reportRepository struct {
	database   database.Database
	collection string
}

func NewReportRepository(db database.Database) ReportRepository {
	return &reportRepository{
		database:   db,
		collection: ReportCollection,
	}
}

func (r *reportRepository) Save(ctx context.Context, report *types.Report) error {
	return r.database.Save(ctx, r.collection, report)
}
func (r *reportRepository) FindByID(ctx context.Context, id string) (*types.Report, error) {
	report := &types.Report{}
	err := r.database.FindByID(ctx, r.collection, id, report)
	if err != nil {
		return nil, err
	}
	return report, nil
}
func (r *reportRepository) FindByTaskID(ctx context.Context, taskID string) ([]*types.Report, error) {
	reports := make([]*types.Report, 0)
	filter := bson.M{"task_id": taskID}
	err := r.database.Query(ctx, r.collection, filter, 0, 0, defaultReportSort, &reports)
	if err != nil {
		return nil, err
	}
	return reports, nil
}
func (r *reportRepository) FindByTaskIDs(ctx context.Context, taskIDs []string) (map[string][]*types.Report, error) {
	reports := make([]*types.Report, 0)
	filter := bson.M{"task_id": bson.M{"$in": taskIDs}}
	err := r.database.Query(ctx, r.collection, filter, 0, 0, defaultReportSort, &reports)
	if err != nil {
		return nil, err
	}
	reportsMap := make(map[string][]*types.Report)
	for _, report := range reports {
		reportsMap[report.TaskID] = append(reportsMap[report.TaskID], report)
	}
	return reportsMap, nil
}
func (r *reportRepository) FilterReports(ctx context.Context, page int64, limit int64, filter types.ReportFilter) ([]*types.Report, int64, error) {

	pipelineMongo := r.pipelineFromReportFilter(filter)
	// Add skip and limit stages for pagination
	var skip int64 = 0
	if page > 0 {
		skip = (page - 1) * limit
	}

	// Add pagination stages if limit > 0
	if limit > 0 {
		pipelineMongo = append(pipelineMongo, bson.M{"$skip": skip})
		pipelineMongo = append(pipelineMongo, bson.M{"$limit": limit})
	}
	// Execute aggregation pipeline
	reports := make([]*types.Report, 0)
	err := r.database.Aggregate(ctx, r.collection, pipelineMongo, &reports)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	total, err := r.CountWithFilter(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil

}
func (r *reportRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.database.Count(ctx, r.collection, nil)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (r *reportRepository) CountWithFilter(ctx context.Context, filter types.ReportFilter) (int64, error) {
	pipelineMongo := r.pipelineFromReportFilter(filter)
	// Add count stage
	pipelineMongo = append(pipelineMongo, bson.M{"$count": "total"})
	// Execute aggregation pipeline
	countData := make([]int64, 0)
	err := r.database.Aggregate(ctx, r.collection, pipelineMongo, &countData)
	if err != nil {
		return 0, err
	}
	if len(countData) == 0 {
		return 0, nil
	}
	return countData[0], nil
}
func (r *reportRepository) Delete(ctx context.Context, id string) error {
	err := r.database.Delete(ctx, r.collection, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *reportRepository) DeleteByTaskID(ctx context.Context, taskID string) error {
	err := r.database.DeleteMany(ctx, r.collection, bson.M{"task_id": taskID})
	if err != nil {
		return err
	}
	return nil
}

func (r *reportRepository) Update(ctx context.Context, id string, report *types.Report) error {
	err := r.database.Update(ctx, r.collection, id, report)
	if err != nil {
		return err
	}
	return nil
}
func (r *reportRepository) FindAll(ctx context.Context) ([]*types.Report, error) {
	reports := make([]*types.Report, 0)
	err := r.database.Query(ctx, r.collection, nil, 0, 0, defaultReportSort, &reports)
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *reportRepository) pipelineFromReportFilter(filter types.ReportFilter) []bson.M {
	mongoFilter := bson.M{}
	if filter.TaskID != "" {
		mongoFilter["task_id"] = filter.TaskID
	}
	if filter.Creator != "" {
		mongoFilter["creator"] = filter.Creator
	}
	if filter.CreatedFrom != 0 {
		mongoFilter["created_at"] = bson.M{"$gte": filter.CreatedFrom}
	}
	if filter.CreatedTo != 0 {
		mongoFilter["created_at"] = bson.M{"$lte": filter.CreatedTo}
	}
	if filter.Workspace != "" {
		mongoFilter["task.workspace"] = filter.Workspace // Stage 1: Lookup Task collection
	}

	lookupStage := bson.M{
		"$lookup": bson.M{
			"from":         "tasks",
			"localField":   "task_id",
			"foreignField": "_id",
			"as":           "task",
		},
	}
	// Stage 2: Unwind task array
	unwindStage := bson.M{
		"$unwind": "$task",
	}

	matchStage := bson.M{
		"$match": mongoFilter,
	}

	projectStage := bson.M{
		"$project": bson.M{
			"_id":        1,
			"task_id":    1,
			"creator":    1,
			"report":     1,
			"created_at": 1,
			"updated_at": 1,
		},
	}

	pipelineMongo := []bson.M{
		lookupStage,
		unwindStage,
		matchStage,
		projectStage,
		defaultReportSort,
	}

	return pipelineMongo
}
