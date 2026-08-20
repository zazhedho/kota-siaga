package warningservice

import (
	"context"
	"errors"
	"strings"

	sharedcache "kota-siaga/internal/cache"
	warningcache "kota-siaga/internal/cache/warning"
	"kota-siaga/internal/dto"
	"kota-siaga/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidProvince = errors.New("invalid province")
	ErrWarningClient   = errors.New("warning upstream client is not configured")
)

type UpstreamClient interface {
	ListWarnings(context.Context, string) ([]dto.Warning, error)
}

type Service struct {
	Client UpstreamClient
	Redis  *redis.Client
}

type WarningService = Service

func NewService(client UpstreamClient, redisClient *redis.Client) *Service {
	return &Service{Client: client, Redis: redisClient}
}

func NewWarningService(client UpstreamClient, redisClient *redis.Client) *Service {
	return NewService(client, redisClient)
}

func ValidateProvince(province string) error {
	if strings.TrimSpace(province) == "" {
		return ErrInvalidProvince
	}
	return nil
}

func (s *Service) ListWarnings(ctx context.Context, province string) ([]dto.Warning, error) {
	if err := ValidateProvince(province); err != nil {
		return nil, err
	}
	if s == nil || s.Client == nil {
		return nil, ErrWarningClient
	}

	province = strings.TrimSpace(province)
	key := warningcache.Key(province)
	var cached []dto.Warning
	hit, err := sharedcache.GetJSON(ctx, s.Redis, key, &cached)
	if err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Warning cache read failed; continuing with upstream request")
	} else if hit {
		return cached, nil
	}

	result, err := s.Client.ListWarnings(ctx, province)
	if err != nil {
		return nil, err
	}
	if err := sharedcache.SetJSON(ctx, s.Redis, key, result, warningcache.TTL()); err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Warning cache write failed; continuing without cache")
	}
	return result, nil
}
