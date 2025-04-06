package repository

import (
	"context"

	"github.com/remiehneppo/be-task-management/internal/database"
	"github.com/remiehneppo/be-task-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var _ UserRepository = &userRepository{}

type UserRepository interface {
	Save(ctx context.Context, user *types.User) error
	FindByID(ctx context.Context, id string) (*types.User, error)
	FindByIDs(ctx context.Context, ids []string) (map[string]*types.User, error)
	FindByUsername(ctx context.Context, username string) (*types.User, error)
	FindAll(ctx context.Context) ([]*types.User, error)
	Update(ctx context.Context, id string, user *types.User) error
	Delete(ctx context.Context, id string) error
	FindByWorkspace(ctx context.Context, workspace string) ([]*types.User, error)
	FindByWorkspaceAndRole(ctx context.Context, workspace string, role string) ([]*types.User, error)
	Paginate(ctx context.Context, page int64, limit int64) ([]*types.User, int64, error)
	Count(ctx context.Context) (int64, error)
}

type userRepository struct {
	database   database.Database
	collection string
}

func NewUserRepository(db database.Database) UserRepository {
	return &userRepository{
		database:   db,
		collection: "users",
	}
}

func (r *userRepository) Save(ctx context.Context, user *types.User) error {
	return r.database.Save(ctx, r.collection, user)
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*types.User, error) {
	user := &types.User{}
	err := r.database.FindByID(ctx, r.collection, id, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByIDs(ctx context.Context, ids []string) (map[string]*types.User, error) {
	objIds := make([]bson.ObjectID, len(ids))
	for i, id := range ids {
		objId, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		objIds[i] = objId
	}
	filter := bson.M{
		"_id": bson.M{
			"$in": objIds,
		},
	}
	users := make([]*types.User, 0)
	err := r.database.Query(ctx, r.collection, filter, 0, 0, nil, &users)
	if err != nil {
		return nil, err
	}
	usersMap := make(map[string]*types.User)
	for _, user := range users {
		usersMap[user.ID] = user
	}
	return usersMap, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*types.User, error) {
	var users []*types.User
	err := r.database.Query(ctx, r.collection, bson.M{"username": username}, 0, 0, nil, &users)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, types.ErrUserNotFound
	}
	return users[0], nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]*types.User, error) {
	var users []*types.User
	err := r.database.FindAll(ctx, r.collection, nil, users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, id string, user *types.User) error {

	return r.database.Update(ctx, r.collection, id, user)
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.database.Delete(ctx, r.collection, id)
}

func (r *userRepository) FindByWorkspace(ctx context.Context, workspace string) ([]*types.User, error) {
	filter := bson.M{
		"workspace": workspace,
	}
	users := make([]*types.User, 0)
	err := r.database.Query(ctx, r.collection, filter, 0, 0, nil, &users)
	if err != nil {
		return nil, err
	}
	// Remove password from users
	for _, user := range users {
		user.Password = ""
	}
	// This is important for security reasons
	return users, nil
}

func (r *userRepository) FindByWorkspaceAndRole(ctx context.Context, workspace string, role string) ([]*types.User, error) {
	users := make([]*types.User, 0)
	err := r.database.Query(ctx, r.collection, bson.M{
		"workspace": workspace,
		"role":      role,
	}, 0, 0, nil, users)
	if err != nil {
		return nil, err
	}
	// Remove password from users
	for _, user := range users {
		user.Password = ""
	}
	return users, nil
}

func (r *userRepository) Paginate(ctx context.Context, page int64, limit int64) ([]*types.User, int64, error) {
	users := make([]*types.User, 0)
	err := r.database.Query(ctx, r.collection, nil, page*limit, limit, nil, users)
	if err != nil {
		return nil, 0, err
	}
	// Remove password from users
	for _, user := range users {
		user.Password = ""
	}
	count, err := r.database.Count(ctx, r.collection, nil)
	if err != nil {
		return nil, 0, err
	}
	return users, count, nil

}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.database.Count(ctx, r.collection, nil)
	if err != nil {
		return 0, err
	}
	return count, nil
}
