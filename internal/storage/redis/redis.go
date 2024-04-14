package redis

import (
	"context"
	"fmt"

	"github.com/JustForWorld/banner-shift/internal/storage"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	db  *redis.Client
	ctx context.Context
}

func New(addr, user, password string, db, protocol int) (*Storage, error) {
	const op = "storage.redis.New"

	url := fmt.Sprintf("redis://%v:%v@%v/%v?protocol=%v", user, password, addr, db, protocol)

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: redis.NewClient(opts)}, nil
}

func (s *Storage) SetBanner(tagID, featureID int64, content interface{}) error {
	const op = "storage.redis.SetBanner"

	key := fmt.Sprintf("banner:%d:%d", tagID, featureID)
	err := s.db.Set(s.ctx, key, content, 0).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
	}

	return nil
}
