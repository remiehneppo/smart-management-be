package service

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

var _ LockService = (*lockService)(nil)

type LockService interface {
	Lock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, key string) error
	WaitIfLocked(ctx context.Context, key string)
}

type lockService struct {
	pool *redsync.Redsync
}

func NewLockService(redisClient *redis.Client) *lockService {
	return &lockService{
		pool: redsync.New(goredis.NewPool(redisClient)),
	}
}

func (r *lockService) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	mutex := r.pool.NewMutex(key, redsync.WithExpiry(expiration))
	if err := mutex.LockContext(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (r *lockService) ReleaseLock(ctx context.Context, key string) error {
	_, err := r.pool.NewMutex(key).UnlockContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *lockService) WaitIfLocked(ctx context.Context, key string) {
	mutex := r.pool.NewMutex(key)
	for {
		if err := mutex.LockContext(ctx); err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
}
