package database

import "context"

type Database interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error

	Save(ctx context.Context, collection string, data interface{}) error
	FindByID(ctx context.Context, collection string, id string, data interface{}) error
	FindAll(ctx context.Context, collection string, sort interface{}, data interface{}) error
	Update(ctx context.Context, collection string, id string, data interface{}) error
	Delete(ctx context.Context, collection string, id string) error
	DeleteMany(ctx context.Context, collection string, filter interface{}) error
	Query(ctx context.Context, collection string, filter interface{}, skip int64, limit int64, sort interface{}, data interface{}) error
	Aggregate(ctx context.Context, collection string, pipeline interface{}, data interface{}) error
	Count(ctx context.Context, collection string, filter interface{}) (int64, error)
}
