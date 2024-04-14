package redis

import (
	"context"
	"fmt"

	"github.com/JustForWorld/banner-shift/internal/storage"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	db *redis.Client
}

func New(ctx context.Context, addr, user, password string, db, protocol int) (*Storage, error) {
	const op = "storage.redis.New"

	url := fmt.Sprintf("redis://%v:%v@%v/%v?protocol=%v", user, password, addr, db, protocol)

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	dbm := redis.NewClient(opts)
	if status := dbm.Ping(ctx); status.Err() != nil {
		return nil, status.Err()
	}

	return &Storage{db: dbm}, nil
}

func (s *Storage) SetBanner(ctx context.Context, tagID, featureID int64, content []byte) error {
	const op = "storage.redis.SetBanner"

	key := fmt.Sprintf("%d:%d", tagID, featureID)
	err := s.db.Set(ctx, key, content, 0).Err()
	if err != nil {
		return fmt.Errorf("%s: %w: %w", op, storage.ErrBannerInvalidData, err)
	}

	return nil
}

func (s *Storage) GetBanner(ctx context.Context, key string) ([]byte, error) {
	const op = "storage.redis.GetBanner"

	value, err := s.db.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %w", op, storage.ErrBannerNotFound, err)
	}

	return []byte(value), nil
}
