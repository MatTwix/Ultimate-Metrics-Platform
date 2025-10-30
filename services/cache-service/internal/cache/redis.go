package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/pkg/models"
	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

func NewredisCache(client *redis.Client) Cache {
	return &RedisCache{client: client}
}

func (r *RedisCache) SetMetric(ctx context.Context, metric models.Metric) error {
	key := fmt.Sprintf("%s:%s", metric.Source, metric.Name)
	data, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal data for caching: %w", err)
	}

	return r.client.Set(ctx, key, data, 5*time.Minute).Err()
}

func (r *RedisCache) GetMetric(ctx context.Context, source, name string) (*models.Metric, error) {
	key := fmt.Sprintf("%s:%s", source, name)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get data from cache: %w", err)
	}

	var metric models.Metric
	if err := json.Unmarshal([]byte(data), &metric); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data from cache: %w", err)
	}

	return &metric, nil
}
