package database

import (
	"context"
	"fmt"
	"time"

	"kota-siaga/pkg/logger"
	"kota-siaga/utils"

	"github.com/redis/go-redis/v9"
)

func InitRedis() (*redis.Client, error) {
	var opt *redis.Options
	redisURL := utils.GetEnv("REDIS_URL", "")
	if redisURL != "" {
		var err error
		opt, err = redis.ParseURL(redisURL)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_URL: %w", err)
		}
	} else {
		opt = &redis.Options{
			Addr:         fmt.Sprintf("%s:%s", utils.GetEnv("REDIS_HOST", "localhost"), utils.GetEnv("REDIS_PORT", "6379")),
			Password:     utils.GetEnv("REDIS_PASSWORD", ""),
			DB:           utils.GetEnv("REDIS_DB", 0),
			DialTimeout:  10 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			PoolSize:     10,
			PoolTimeout:  30 * time.Second,
		}
	}
	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		_ = client.Close()
		logger.WriteLog(logger.LogLevelError, fmt.Sprintf("Failed to connect to Redis: %v", err))
		return nil, err
	}

	logger.WriteLog(logger.LogLevelInfo, "Connected to Redis")
	return client, nil
}
