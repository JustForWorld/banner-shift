package storage

import "errors"

var (
	ErrBannerNotFound    = errors.New("banner not found")
	ErrBannerInvalidData = errors.New("banner received invalid data")
	ErrBannerExists      = errors.New("banner exists")
	ErrBannerNotExists   = errors.New("banner not exists")
	ErrBannerNotAdd      = errors.New("cannot be added")
)
