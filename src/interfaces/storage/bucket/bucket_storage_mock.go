package bucket

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockBucketStorage struct {
	mock.Mock
}

func (m *MockBucketStorage) CheckRateLimit(ctx context.Context, key string, capacity int, leakRate time.Duration) (bool, error) {
	args := m.Called(ctx, key, capacity, leakRate)
	return args.Bool(0), args.Error(1)
}

func (m *MockBucketStorage) ResetBucket(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}
