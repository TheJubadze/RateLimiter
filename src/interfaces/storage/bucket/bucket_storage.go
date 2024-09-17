package bucket

import (
	"context"
	"time"
)

type Storage interface {
	CheckRateLimit(ctx context.Context, key string, limit int, leakRate time.Duration) (bool, error)
	ResetBucket(ctx context.Context, key string) error
}
