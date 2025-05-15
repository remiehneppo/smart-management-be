package repository

import (
	"context"

	"github.com/remiehneppo/be-task-management/internal/database"
	"github.com/remiehneppo/be-task-management/types"
)

var PendingDocumentCollection = "pending_documents"

var _ PendingDocumentRepository = (*pendingDocumentRepository)(nil)

type PendingDocumentRepository interface {
	Save(ctx context.Context, pendingDocument *types.PendingDocument) error
	FindByID(ctx context.Context, id string) (*types.PendingDocument, error)
	FindAll(ctx context.Context, page, limit int64) ([]*types.PendingDocument, int64, error)
	Remove(ctx context.Context, id string) error
}

type pendingDocumentRepository struct {
	database   database.Database
	collection string
}

func NewPendingDocumentRepository(db database.Database) *pendingDocumentRepository {
	return &pendingDocumentRepository{
		database:   db,
		collection: PendingDocumentCollection,
	}
}

func (r *pendingDocumentRepository) Save(ctx context.Context, pendingDocument *types.PendingDocument) error {
	return r.database.Save(ctx, r.collection, pendingDocument)
}

func (r *pendingDocumentRepository) FindByID(ctx context.Context, id string) (*types.PendingDocument, error) {
	pendingDocument := &types.PendingDocument{}
	err := r.database.FindByID(ctx, r.collection, id, pendingDocument)
	if err != nil {
		return nil, err
	}
	return pendingDocument, nil
}

func (r *pendingDocumentRepository) FindAll(ctx context.Context, page, limit int64) ([]*types.PendingDocument, int64, error) {
	pendingDocuments := make([]*types.PendingDocument, 0)
	totalCount, err := r.database.Count(ctx, r.collection, nil)
	if err != nil {
		return nil, 0, err
	}

	err = r.database.Query(ctx, r.collection, nil, page*limit, limit, nil, &pendingDocuments)
	if err != nil {
		return nil, 0, err
	}
	return pendingDocuments, totalCount, nil
}

func (r *pendingDocumentRepository) Remove(ctx context.Context, id string) error {
	return r.database.Delete(ctx, r.collection, id)
}
