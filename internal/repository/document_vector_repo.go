package repository

import (
	"context"
	"fmt"

	"github.com/remiehneppo/be-task-management/types"
	"github.com/remiehneppo/be-task-management/utils"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

var _ DocumentVectorRepository = (*documentVectorRepository)(nil)

var DefaultDocumentClass = &models.Class{
	Class: "Document",
	Properties: []*models.Property{
		{Name: "title", DataType: []string{"text"}},
		{Name: "content", DataType: []string{"text"}},
		{Name: "page_number", DataType: []string{"int"}},
		{Name: "chunk_number", DataType: []string{"int"}},
		{Name: "tags", DataType: []string{"text[]"}},
	},
}

type DocumentVectorRepository interface {
	SaveBatchDocumentVector(ctx context.Context, metadata *types.DocumentMetadata, document []*types.DocumentChunk) error
	SaveDocumentVector(ctx context.Context, metadata *types.DocumentMetadata, document *types.DocumentChunk) error
	SearchDocumentVector(ctx context.Context, metadata *types.DocumentMetadata, queries []string, limit int) ([]*types.ChunkDocumentResponse, error)
}

type documentVectorRepository struct {
	batchSize int
	client    *weaviate.Client
	class     *models.Class
}

func NewDocumentVectorRepository(ctx context.Context, client *weaviate.Client, documentClass *models.Class, batchSize int) *documentVectorRepository {
	schema, err := client.Schema().Getter().Do(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to get schema: %v", err))
	}
	hasDocumentClass := false
	for _, class := range schema.Classes {
		if class.Class == documentClass.Class {
			hasDocumentClass = true
			break
		}
	}
	if !hasDocumentClass {
		err = client.Schema().ClassCreator().WithClass(documentClass).Do(context.Background())
		if err != nil {
			panic(fmt.Sprintf("failed to create class %s: %v", documentClass.Class, err))
		}
	}
	return &documentVectorRepository{
		client:    client,
		class:     documentClass,
		batchSize: batchSize,
	}
}

func (r *documentVectorRepository) SaveBatchDocumentVector(ctx context.Context, metadata *types.DocumentMetadata, documents []*types.DocumentChunk) error {
	total := len(documents)
	for i := 0; i < total; i += r.batchSize {
		end := i + r.batchSize
		if end > total {
			end = total
		}
		batcher := r.client.Batch().ObjectsBatcher()

		for j := i; j < end; j++ {
			properties := map[string]interface{}{
				"title":        metadata.Title,
				"content":      documents[j].Content,
				"page_number":  documents[j].Page,
				"chunk_number": documents[j].Chunk,
				"tags":         metadata.Tags,
			}
			batcher.WithObjects(
				&models.Object{
					Class:      r.class.Class,
					Properties: properties,
				},
			)
		}
		_, err := batcher.Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to save batch document vector: %w", err)
		}
	}
	return nil
}

func (r *documentVectorRepository) SaveDocumentVector(ctx context.Context, metadata *types.DocumentMetadata, document *types.DocumentChunk) error {
	properties := map[string]interface{}{
		"title":        metadata.Title,
		"content":      document.Content,
		"page_number":  document.Page,
		"chunk_number": document.Chunk,
		"tags":         metadata.Tags,
	}
	creator := r.client.Data().Creator().
		WithClassName(r.class.Class).
		WithProperties(properties)
	_, err := creator.Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to save document vector: %w", err)
	}
	return nil
}

func (r *documentVectorRepository) SearchDocumentVector(ctx context.Context, metadata *types.DocumentMetadata, queries []string, limit int) ([]*types.ChunkDocumentResponse, error) {
	fields := []graphql.Field{
		{Name: "title"},
		{Name: "content"},
		{Name: "page_number"},
		{Name: "chunk_number"},
		{Name: "tags"},
		{Name: "_additional", Fields: []graphql.Field{
			{Name: "id"},
		},
		},
	}
	getBuilder := r.client.GraphQL().Get().
		WithClassName(r.class.Class).
		WithFields(fields...)
	nearVector := r.client.GraphQL().NearTextArgBuilder().
		WithConcepts(queries).
		WithCertainty(0.7)
	whereFilter := buildMetadataFilter(metadata)

	if limit > 0 {
		getBuilder.WithLimit(limit)
	}
	if whereFilter != nil {
		getBuilder.WithWhere(whereFilter)
	}
	getBuilder.WithNearText(nearVector)

	result, err := getBuilder.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search document vector: %w", err)
	}
	if result.Errors != nil {
		return nil, fmt.Errorf("failed to search document vector: %v", result.Errors)
	}
	var docs []*types.ChunkDocumentResponse
	if data, ok := result.Data["Get"].(map[string]interface{})[r.class.Class].([]interface{}); ok {
		for _, item := range data {
			if doc, ok := item.(map[string]interface{}); ok {
				id, ok := doc["_additional"].(map[string]interface{})["id"].(string)
				if !ok {
					id = ""
				}
				docs = append(docs, &types.ChunkDocumentResponse{
					ID:          id,
					Title:       doc["title"].(string),
					Content:     doc["content"].(string),
					PageNumber:  int(doc["page_number"].(float64)),
					ChunkNumber: int(doc["chunk_number"].(float64)),
					Tags:        utils.ParseStringArray(doc["tags"]),
				})
			}
		}
	}
	return docs, nil
}

func buildMetadataFilter(metadata *types.DocumentMetadata) *filters.WhereBuilder {

	var whereFilter *filters.WhereBuilder

	if metadata.Title != "" {
		whereFilter = filters.Where().WithPath([]string{"title"}).
			WithOperator(filters.Equal).
			WithValueString(metadata.Title)
	}

	if metadata.Source != "" {
		sourceFilter := filters.Where().
			WithPath([]string{"source"}).
			WithOperator(filters.Equal).
			WithValueString(metadata.Source)
		if whereFilter == nil {
			whereFilter = sourceFilter
		} else {
			whereFilter = whereFilter.WithOperator(filters.And).WithOperands([]*filters.WhereBuilder{sourceFilter})
		}

	}

	if len(metadata.Tags) > 0 {
		for _, tag := range metadata.Tags {
			tagFilter := filters.Where().
				WithPath([]string{"tags"}).
				WithOperator(filters.ContainsAny).
				WithValueString(tag)
			if whereFilter == nil {
				whereFilter = tagFilter
			} else {
				whereFilter = whereFilter.WithOperator(filters.And).WithOperands([]*filters.WhereBuilder{tagFilter})
			}
		}
	}

	return whereFilter
}
