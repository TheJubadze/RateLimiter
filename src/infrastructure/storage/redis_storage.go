package redisstorage

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/TheJubadze/RateLimiter/pkg/logger"
	"github.com/go-redis/redis/v8"
)

type RedisBucketStorage struct {
	logger logger.Logger
	client *redis.Client
}

func NewRedisBucketStorage(logger logger.Logger, client *redis.Client) *RedisBucketStorage {
	return &RedisBucketStorage{
		logger: logger,
		client: client,
	}
}

func (r *RedisBucketStorage) CheckRateLimit(ctx context.Context, key string, capacity int, leakRate time.Duration) (bool, error) {
	now := time.Now().Unix()

	// Get the last request count and timestamp from Redis
	pipe := r.client.TxPipeline()

	// Fetch bucket count and last updated time (if present)
	countCmd := pipe.Get(ctx, key+":count")
	lastLeakCmd := pipe.Get(ctx, key+":lastLeak")

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, err
	}

	// Parse the count and last leak timestamp (defaults to 0 if not present)
	count, _ := strconv.ParseInt(countCmd.Val(), 10, 64)
	lastLeak, _ := strconv.ParseInt(lastLeakCmd.Val(), 10, 64)

	// Calculate the time since last leak
	if lastLeak == 0 {
		lastLeak = now
	}

	elapsed := now - lastLeak
	leakedRequests := int64(math.Floor(float64(elapsed) / leakRate.Seconds() * float64(capacity)))

	// Subtract the leaked requests from the current count
	count -= leakedRequests
	if count < 0 {
		count = 0
	}

	// If the bucket is still below capacity, allow the request and update the count
	if count < int64(capacity) {
		count++

		pipe := r.client.TxPipeline()

		// Update the count and the last leak timestamp
		pipe.Set(ctx, key+":count", count, 0)
		pipe.Set(ctx, key+":lastLeak", now, 0)

		_, err := pipe.Exec(ctx)
		if err != nil {
			return false, err
		}

		r.logger.Printf("Key: %s, count: %d, lastLeak: %s", key, count, time.Unix(lastLeak, 0).Format("2006-01-02 15:04:05"))
		return true, nil
	}

	// If the count exceeds the capacity, reject the request
	r.logger.Printf("Key: %s, count: %d - rate limit exceeded", key, count)
	return false, nil
}

func (r *RedisBucketStorage) ResetBucket(ctx context.Context, key string) error {
	// Reset the count and lastLeak for the bucket
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, key+":count")
	pipe.Del(ctx, key+":lastLeak")

	_, err := pipe.Exec(ctx)
	return err
}
