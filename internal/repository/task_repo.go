package repository

import (
	"context"

	"github.com/remiehneppo/be-task-management/internal/database"
	"github.com/remiehneppo/be-task-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const TaskCollection = "tasks"

var defaultSort = bson.M{"deadline": -1} // sort by deadline descending

type TaskRepository interface {
	Save(ctx context.Context, task *types.Task) error
	FindByID(ctx context.Context, id string) (*types.Task, error)
	FindAll(ctx context.Context) ([]*types.Task, error)
	Update(ctx context.Context, id string, task *types.Task) error
	Delete(ctx context.Context, id string) error
	FindByWorkspace(ctx context.Context, workspace string) ([]*types.Task, error)
	FindByWorkspaceAndStatus(ctx context.Context, workspace string, status string) ([]*types.Task, error)
	Paginate(ctx context.Context, page int64, limit int64) ([]*types.Task, int64, error)
	PaginateWithFilter(ctx context.Context, page int64, limit int64, filter types.TaskFilter) ([]*types.Task, int64, error)
	Count(ctx context.Context) (int64, error)
	CountWithFilter(ctx context.Context, filter types.TaskFilter) (int64, error)
}

type taskRepository struct {
	database   database.Database
	collection string
}

func NewTaskRepository(db database.Database) TaskRepository {
	return &taskRepository{
		database:   db,
		collection: TaskCollection,
	}
}

func (r *taskRepository) Save(ctx context.Context, task *types.Task) error {
	return r.database.Save(ctx, r.collection, task)
}
func (r *taskRepository) FindByID(ctx context.Context, id string) (*types.Task, error) {
	var task types.Task
	err := r.database.FindByID(ctx, r.collection, id, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}
func (r *taskRepository) FindAll(ctx context.Context) ([]*types.Task, error) {
	var tasks []*types.Task
	err := r.database.FindAll(ctx, r.collection, defaultSort, tasks) // sort by deadline descending
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
func (r *taskRepository) Update(ctx context.Context, id string, task *types.Task) error {
	err := r.database.Update(ctx, r.collection, id, task)
	if err != nil {
		return err
	}
	return nil
}
func (r *taskRepository) Delete(ctx context.Context, id string) error {
	err := r.database.Delete(ctx, r.collection, id)
	if err != nil {
		return err
	}
	return nil
}
func (r *taskRepository) FindByWorkspace(ctx context.Context, workspace string) ([]*types.Task, error) {
	mongoFilter := bson.M{
		"workspace": workspace,
	}
	var tasks []*types.Task
	err := r.database.Query(ctx, r.collection, mongoFilter, 0, 0, defaultSort, tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
func (r *taskRepository) FindByWorkspaceAndStatus(ctx context.Context, workspace string, status string) ([]*types.Task, error) {
	mongoFilter := bson.M{
		"workspace": workspace,
		"status":    status,
	}
	var tasks []*types.Task
	err := r.database.Query(ctx, r.collection, mongoFilter, 0, 0, defaultSort, tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
func (r *taskRepository) Paginate(ctx context.Context, page int64, limit int64) ([]*types.Task, int64, error) {
	var tasks []*types.Task
	err := r.database.Query(ctx, r.collection, nil, page*limit, limit, defaultSort, tasks)
	if err != nil {
		return nil, 0, err
	}
	count, err := r.database.Count(ctx, r.collection, nil)
	if err != nil {
		return nil, 0, err
	}
	return tasks, count, nil
}
func (r *taskRepository) PaginateWithFilter(ctx context.Context, page int64, limit int64, filter types.TaskFilter) ([]*types.Task, int64, error) {

	match, lookup := r.pipelineFromTaskFilter(filter)

	projectStage := bson.M{
		"$project": bson.M{
			"_id":         1,
			"title":       1,
			"description": 1,
			"workspace":   1,
			"creator":     1,
			"deadline":    1,
			"assignee":    1,
			"status":      1,
			"created_at":  1,
			"start_at":    1,
			"updated_at":  1,
			"progress":    1,
		},
	}

	facetStage := bson.M{
		"$facet": bson.M{
			"metadata": []bson.M{
				match,
				{"$count": "total"},
			},
			"data": []bson.M{
				match,
				projectStage,
				{"$sort": defaultSort},
				{"$skip": (page - 1) * limit},
				{"$limit": limit},
			},
		},
	}
	if lookup != nil {
		facetStage["$facet"].(bson.M)["data"] = append(facetStage["$facet"].(bson.M)["data"].([]bson.M), lookup)
		facetStage["$facet"].(bson.M)["metadata"] = append(facetStage["$facet"].(bson.M)["metadata"].([]bson.M), lookup)
	}
	// Định nghĩa kiểu dữ liệu để match với cấu trúc JSON
	type AggregationResult struct {
		Metadata []struct {
			Total int64 `json:"total"`
		} `json:"metadata"`
		Data []types.Task `json:"data"`
	}
	aggregationResult := []AggregationResult{}
	err := r.database.Aggregate(
		ctx,
		r.collection,
		[]bson.M{facetStage},
		&aggregationResult,
	)
	if err != nil {
		return nil, 0, err
	}

	tasks := make([]*types.Task, 0)
	for _, item := range aggregationResult[0].Data {
		task := &types.Task{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			Workspace:   item.Workspace,
			Creator:     item.Creator,
			Deadline:    item.Deadline,
			Assignee:    item.Assignee,
			Status:      item.Status,
			Progress:    item.Progress,
			CreateAt:    item.CreateAt,
			StartAt:     item.StartAt,
			UpdateAt:    item.UpdateAt,
		}
		tasks = append(tasks, task)
	}
	total := int64(0)
	if len(aggregationResult[0].Metadata) > 0 {
		total = aggregationResult[0].Metadata[0].Total
	}

	// var total int64 = 0
	// if metadataArr, ok := facetResult["metadata"].(bson.A); ok && len(metadataArr) > 0 {
	// 	if metadata, ok := metadataArr[0].(bson.M); ok {
	// 		if totalVal, ok := metadata["total"]; ok {
	// 			total = totalVal.(int64)
	// 		}
	// 	}
	// } else {
	// 	return nil, 0, nil
	// }

	// tasks := make([]*types.Task, 0)
	// if dataArr, ok := facetResult["data"].([]bson.M); ok {
	// 	for _, item := range dataArr {

	// 		if !ok {
	// 			continue
	// 		}

	// 		task := &types.Task{
	// 			ID:          item["_id"].(string),
	// 			Title:       item["title"].(string),
	// 			Description: item["description"].(string),
	// 			Workspace:   item["workspace"].(string),
	// 			Creator:     item["creator"].(string),
	// 			Deadline:    item["deadline"].(int64),
	// 			Assignee:    item["assignee"].(string),
	// 			Status:      item["status"].(string),
	// 			Progress:    int(item["progress"].(int32)),
	// 			CreateAt:    item["created_at"].(int64),
	// 			StartAt:     item["start_at"].(int64),
	// 			UpdateAt:    item["updated_at"].(int64),
	// 		}
	// 		tasks = append(tasks, task)
	// 	}
	// }

	return tasks, total, nil
}
func (r *taskRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.database.Count(ctx, r.collection, nil)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (r *taskRepository) CountWithFilter(ctx context.Context, filter types.TaskFilter) (int64, error) {
	match, lookup := r.pipelineFromTaskFilter(filter)

	countStage := bson.M{"$count": "total"}
	pipeline := []bson.M{match, lookup, countStage}
	var countResult bson.M
	err := r.database.Aggregate(ctx, r.collection, pipeline, countResult)
	if err != nil {
		return 0, err
	}

	var total int64 = 0
	if totalVal, ok := countResult["total"]; ok {
		total = totalVal.(int64)

	}

	return total, nil
}

func (r *taskRepository) newMongoFilter(filter types.TaskFilter) bson.M {
	mongoFilter := bson.M{}
	if filter.Title != "" {
		mongoFilter["title"] = bson.M{"$regex": filter.Title}
	}
	if filter.Workspace != "" {
		mongoFilter["workspace"] = filter.Workspace
	}
	if filter.Creator != "" {
		mongoFilter["creator"] = filter.Creator
	}
	if filter.DeadlineFrom != 0 {
		mongoFilter["deadline"] = bson.M{"$gte": filter.DeadlineFrom}
	}
	if filter.DeadlineTo != 0 {
		mongoFilter["deadline"] = bson.M{"$lte": filter.DeadlineTo}
	}
	if filter.Assignee != "" {
		mongoFilter["assignee"] = filter.Assignee
	}
	if filter.Status != "" {
		mongoFilter["status"] = filter.Status
	}

	return mongoFilter
}

func (r *taskRepository) pipelineFromTaskFilter(filter types.TaskFilter) (match bson.M, lookup bson.M) {
	match = bson.M{}
	if filter.Title != "" {
		match["title"] = bson.M{"$regex": filter.Title}
	}
	if filter.Workspace != "" {
		match["workspace"] = filter.Workspace
	}
	if filter.Creator != "" {
		match["creator"] = filter.Creator
	}
	if filter.DeadlineFrom != 0 || filter.DeadlineTo != 0 {
		deadlineFilter := bson.M{}
		if filter.DeadlineFrom != 0 {
			deadlineFilter["$gte"] = filter.DeadlineFrom
		}
		if filter.DeadlineTo != 0 {
			deadlineFilter["$lte"] = filter.DeadlineTo
		}
		match["deadline"] = deadlineFilter
	}
	if filter.Assignee != "" {
		match["assignee"] = filter.Assignee
	}
	if filter.Status != "" {
		match["status"] = filter.Status
	}
	if filter.StartFrom != 0 || filter.StartTo != 0 {
		startFilter := bson.M{}
		if filter.StartFrom != 0 {
			startFilter["$gte"] = filter.StartFrom
		}
		if filter.StartTo != 0 {
			startFilter["$lte"] = filter.StartTo
		}
		match["start_at"] = startFilter
	}
	if filter.ReportFrom != 0 || filter.ReportTo != 0 {
		lookup = bson.M{
			"$lookup": bson.M{
				"from":         "reports",
				"localField":   "_id",
				"foreignField": "task_id",
				"as":           "reports",
			},
		}
		//
		reportFilter := bson.M{}
		if filter.ReportFrom != 0 {
			reportFilter["$gte"] = filter.ReportFrom
		}
		if filter.ReportTo != 0 {
			reportFilter["$lte"] = filter.ReportTo
		}
		match["reports"] = bson.M{
			"$elemMatch": bson.M{
				"created_at": reportFilter,
			},
		}
	}
	return bson.M{"$match": match}, lookup
}
