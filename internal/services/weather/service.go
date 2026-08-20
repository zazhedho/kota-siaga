package weatherservice

import (
	"context"
	"errors"
	"fmt"

	sharedcache "kota-siaga/internal/cache"
	weathercache "kota-siaga/internal/cache/weather"
	"kota-siaga/internal/dto"
	"kota-siaga/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidADM4   = errors.New("invalid adm4")
	ErrWeatherClient = errors.New("weather upstream client is not configured")
)

type UpstreamClient interface {
	GetWeather(context.Context, string) ([]dto.WeatherForecast, error)
}

type Service struct {
	Client UpstreamClient
	Redis  *redis.Client
}

type WeatherService = Service

func NewService(client UpstreamClient, redisClient *redis.Client) *Service {
	return &Service{Client: client, Redis: redisClient}
}

func NewWeatherService(client UpstreamClient, redisClient *redis.Client) *Service {
	return NewService(client, redisClient)
}

func ValidateADM4(adm4 string) error {
	if len(adm4) == 0 || len(adm4) > 20 {
		return fmt.Errorf("%w: adm4 must be 1-20 characters", ErrInvalidADM4)
	}

	hasDigit := false
	for i := 0; i < len(adm4); i++ {
		char := adm4[i]
		if char == '.' {
			continue
		}
		if char < '0' || char > '9' {
			return fmt.Errorf("%w: adm4 must contain only digits and dots", ErrInvalidADM4)
		}
		hasDigit = true
	}
	if !hasDigit {
		return fmt.Errorf("%w: adm4 must contain a digit", ErrInvalidADM4)
	}
	return nil
}

func (s *Service) GetWeather(ctx context.Context, adm4 string) ([]dto.WeatherForecast, error) {
	if err := ValidateADM4(adm4); err != nil {
		return nil, err
	}
	if s == nil || s.Client == nil {
		return nil, ErrWeatherClient
	}

	key := weathercache.Key(adm4)
	var cached []dto.WeatherForecast
	hit, err := sharedcache.GetJSON(ctx, s.Redis, key, &cached)
	if err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Weather cache read failed; continuing with upstream request")
	} else if hit {
		return cached, nil
	}

	result, err := s.Client.GetWeather(ctx, adm4)
	if err != nil {
		return nil, err
	}
	if err := sharedcache.SetJSON(ctx, s.Redis, key, result, weathercache.TTL()); err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Weather cache write failed; continuing without cache")
	}
	return result, nil
}
