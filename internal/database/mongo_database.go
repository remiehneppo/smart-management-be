package database

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var _ Database = &mongoDatabase{}

type mongoDatabase struct {
	uri         string
	database    string
	mongoClient *mongo.Client
}

func NewMongoDatabase(uri, database string) Database {
	return &mongoDatabase{
		uri:      uri,
		database: database,
	}
}

func (m *mongoDatabase) Connect(ctx context.Context) error {
	client, err := mongo.Connect(options.Client().
		ApplyURI(m.uri).SetBSONOptions(
		&options.BSONOptions{
			ObjectIDAsHexString: true,
		},
	))
	if err != nil {
		panic(err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		panic(err)
	}
	m.mongoClient = client
	return nil
}

func (m *mongoDatabase) Disconnect(ctx context.Context) error {
	if err := m.mongoClient.Disconnect(ctx); err != nil {
		panic(err)
	}
	return nil
}

func (m *mongoDatabase) Save(ctx context.Context, collection string, data interface{}) error {
	coll := m.mongoClient.Database(m.database).Collection(collection)
	_, err := coll.InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func (m *mongoDatabase) FindByID(ctx context.Context, collection string, id string, data interface{}) error {
	coll := m.mongoClient.Database(m.database).Collection(collection)
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	result := coll.FindOne(ctx, bson.M{"_id": objId})
	if result.Err() != nil {
		return result.Err()
	}
	if err := result.Decode(data); err != nil {
		return err
	}
	return nil
}

func (m *mongoDatabase) FindAll(ctx context.Context, collection string, sort interface{}, data interface{}) error {
	coll := m.mongoClient.Database(m.database).Collection(collection)
	ops := options.Find()
	if sort != nil {
		ops.SetSort(sort)
	}
	cursor, err := coll.Find(ctx, bson.D{}, ops)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, data); err != nil {
		return err
	}
	return nil
}

func (m *mongoDatabase) Update(ctx context.Context, collection string, id string, data interface{}) error {
	coll := m.mongoClient.Database(m.database).Collection(collection)
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = coll.UpdateOne(ctx, map[string]interface{}{"_id": objId}, bson.M{"$set": data})
	if err != nil {
		return err
	}
	return nil
}

func (m *mongoDatabase) Delete(ctx context.Context, collection string, id string) error {
	coll := m.mongoClient.Database(m.database).Collection(collection)
	objId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = coll.DeleteOne(ctx, map[string]interface{}{"_id": objId})
	if err != nil {
		return err
	}
	return nil
}

func (m *mongoDatabase) DeleteMany(ctx context.Context, collection string, filter interface{}) error {
	coll := m.mongoClient.Database(m.database).Collection(collection)
	_, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

func (m *mongoDatabase) Query(ctx context.Context, collection string, filter interface{}, skip int64, limit int64, sort interface{}, data interface{}) error {
	coll := m.mongoClient.Database(m.database).Collection(collection)
	ops := options.Find()
	if sort != nil {
		ops.SetSort(sort)
	}
	if skip > 0 {
		ops.SetSkip(skip)
	}
	if limit > 0 {
		ops.SetLimit(limit)
	}
	if filter == nil {
		filter = bson.D{} // Use an empty filter if nil is provided
	}
	cursor, err := coll.Find(ctx, filter, ops)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, data)
}

func (m *mongoDatabase) Aggregate(ctx context.Context, collection string, pipeline interface{}, data interface{}) error {
	coll := m.mongoClient.Database(m.database).Collection(collection)
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, data)
}

func (m *mongoDatabase) Count(ctx context.Context, collection string, filter interface{}) (int64, error) {
	coll := m.mongoClient.Database(m.database).Collection(collection)

	// Handle nil filter by using empty bson.D{} which matches all documents
	if filter == nil {
		filter = bson.D{}
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	// Optional: Add logging for debugging
	// log.Printf("Count operation on collection '%s' with filter %+v returned: %d", collection, filter, count)

	return count, nil
}

func (m *mongoDatabase) GetClient() *mongo.Client {
	return m.mongoClient
}
