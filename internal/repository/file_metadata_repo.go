package repository

import (
	"context"

	"github.com/remiehneppo/be-task-management/internal/database"
	"github.com/remiehneppo/be-task-management/types"
)

var fileMetadataCollection = "file_metadata"

var _ FileMetadataRepository = (*fileMetadataRepository)(nil)

type FileMetadataRepository interface {
	GetFileMetadata(ctx context.Context, fileID string) (*types.FileMetadata, error)
	GetFileList(ctx context.Context, page, limit int64) ([]*types.FileMetadata, int64, error)
	CreateFileMetadata(ctx context.Context, fileMetadata *types.FileMetadata) error
	UpdateFileMetadata(ctx context.Context, fileMetadata *types.FileMetadata) error
	DeleteFileMetadata(ctx context.Context, fileID string) error
	GetFileByName(ctx context.Context, fileName string) (*types.FileMetadata, error)
}

type fileMetadataRepository struct {
	database   database.Database
	collection string
}

func NewFileMetadataRepository(db database.Database) *fileMetadataRepository {
	return &fileMetadataRepository{
		database:   db,
		collection: fileMetadataCollection,
	}
}

func (r *fileMetadataRepository) GetFileMetadata(ctx context.Context, fileID string) (*types.FileMetadata, error) {
	fileMetadata := &types.FileMetadata{}
	err := r.database.FindByID(ctx, r.collection, fileID, fileMetadata)
	if err != nil {
		return nil, err
	}
	return fileMetadata, nil
}

func (r *fileMetadataRepository) GetFileList(ctx context.Context, page, limit int64) ([]*types.FileMetadata, int64, error) {
	fileMetadataList := make([]*types.FileMetadata, 0)
	totalCount, err := r.database.Count(ctx, r.collection, nil)
	if err != nil {
		return nil, 0, err
	}

	err = r.database.Query(ctx, r.collection, nil, page*limit, limit, nil, &fileMetadataList)
	if err != nil {
		return nil, 0, err
	}
	return fileMetadataList, totalCount, nil
}

func (r *fileMetadataRepository) CreateFileMetadata(ctx context.Context, fileMetadata *types.FileMetadata) error {
	err := r.database.Save(ctx, r.collection, fileMetadata)
	if err != nil {
		return err
	}
	return nil
}

func (r *fileMetadataRepository) UpdateFileMetadata(ctx context.Context, fileMetadata *types.FileMetadata) error {
	err := r.database.Update(ctx, r.collection, fileMetadata.ID, fileMetadata)
	if err != nil {
		return err
	}
	return nil
}

func (r *fileMetadataRepository) DeleteFileMetadata(ctx context.Context, fileID string) error {
	err := r.database.Delete(ctx, r.collection, fileID)
	if err != nil {
		return err
	}
	return nil
}

func (r *fileMetadataRepository) GetFileByName(ctx context.Context, fileName string) (*types.FileMetadata, error) {
	filesMetadata := []*types.FileMetadata{}
	err := r.database.Query(
		ctx,
		r.collection,
		map[string]interface{}{"file_name": fileName},
		0,
		0,
		nil,
		filesMetadata,
	)
	if err != nil {
		return nil, err
	}
	return filesMetadata[0], nil
}
