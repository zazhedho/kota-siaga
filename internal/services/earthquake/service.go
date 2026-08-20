package earthquake

import (
	"context"
	"errors"

	sharedcache "kota-siaga/internal/cache"
	earthquakecache "kota-siaga/internal/cache/earthquake"
	"kota-siaga/internal/dto"
	"kota-siaga/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var ErrEarthquakeClient = errors.New("earthquake upstream client is not configured")

type UpstreamClient interface {
	ListLatest(context.Context) ([]dto.Earthquake, error)
}

type Service struct {
	Client UpstreamClient
	Redis  *redis.Client
}

type EarthquakeService = Service

func NewService(client UpstreamClient, redisClient *redis.Client) *Service {
	return &Service{Client: client, Redis: redisClient}
}

func NewEarthquakeService(client UpstreamClient, redisClient *redis.Client) *Service {
	return NewService(client, redisClient)
}

func (s *Service) ListLatest(ctx context.Context) ([]dto.Earthquake, error) {
	if s == nil || s.Client == nil {
		return nil, ErrEarthquakeClient
	}

	key := earthquakecache.Key()
	var cached []dto.Earthquake
	hit, err := sharedcache.GetJSON(ctx, s.Redis, key, &cached)
	if err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Earthquake cache read failed; continuing with upstream request")
	} else if hit {
		return cached, nil
	}

	result, err := s.Client.ListLatest(ctx)
	if err != nil {
		return nil, err
	}
	if err := sharedcache.SetJSON(ctx, s.Redis, key, result, earthquakecache.TTL()); err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Earthquake cache write failed; continuing without cache")
	}
	return result, nil
}
