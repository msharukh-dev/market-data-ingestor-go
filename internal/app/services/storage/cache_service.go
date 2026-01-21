package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	common "ws_ingestor/internal/app/common/exception_handler"
	"ws_ingestor/internal/app/common/logger"
	"ws_ingestor/internal/app/models"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type CacheService struct {
	Client *redis.Client
	logger *logrus.Logger
}

func NewRedis(addr, password string, db int) (*CacheService, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, common.NewCustomError(common.ErrCacheConnect, "Failed to connect to Redis", err)
	}

	return &CacheService{Client: rdb, logger: logger.GetLogger()}, nil
}

func (c *CacheService) InsertBatch(ctx context.Context, batch []models.MarketData, ttl time.Duration) error {
	pipe := c.Client.Pipeline()

	for _, data := range batch {
		if data.Timestamp == 0 {
			continue // Skip entries with zero timestamp
		}

		key := data.Name
		value, err := json.Marshal(data)
		if err != nil {
			c.logger.Error(fmt.Sprintf("Failed to marshal data for %s: %v", key, err))
			continue
		}

		pipe.Set(ctx, key, value, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Failed to execute Redis pipeline: %v", err))
		return err
	}
	return nil
}

func (c *CacheService) Close() {
	c.Client.Close()
}

func (c *CacheService) GetAllData(ctx context.Context) ([]models.MarketData, error) {
	var allData []models.MarketData

	iter := c.Client.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		value, err := c.Client.Get(ctx, key).Result()
		if err != nil {
			c.logger.Error(fmt.Sprintf("Failed to get value for key %s: %v", key, err))
			continue
		}

		var data models.MarketData
		if err := json.Unmarshal([]byte(value), &data); err != nil {
			c.logger.Error(fmt.Sprintf("Failed to unmarshal data for key %s: %v", key, err))
			continue
		}

		allData = append(allData, data)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return allData, nil
}
