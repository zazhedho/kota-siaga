package locationservice

import (
	"context"
	"errors"
	"fmt"

	sharedcache "kota-siaga/internal/cache"
	locationcache "kota-siaga/internal/cache/location"
	"kota-siaga/internal/dto"
	"kota-siaga/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidPagination = errors.New("invalid pagination")
	ErrInvalidParentID   = errors.New("invalid parent ID")
	ErrLocationClient    = errors.New("location upstream client is not configured")
)

type UpstreamClient interface {
	ListProvinces(context.Context, int, int) (dto.Page[dto.Province], error)
	ListCities(context.Context, string, int, int) (dto.Page[dto.City], error)
	ListDistricts(context.Context, string, int, int) (dto.Page[dto.District], error)
	ListVillages(context.Context, string, int, int) (dto.Page[dto.Village], error)
}

type Service struct {
	Client UpstreamClient
	Redis  *redis.Client
}

type LocationService = Service

func NewService(client UpstreamClient, redisClient *redis.Client) *Service {
	return &Service{Client: client, Redis: redisClient}
}

func NewLocationService(client UpstreamClient, redisClient *redis.Client) *Service {
	return NewService(client, redisClient)
}

func ValidatePagination(page, perPage int) error {
	if page < 1 {
		return fmt.Errorf("%w: page must be at least 1", ErrInvalidPagination)
	}
	if perPage < 1 || perPage > 100 {
		return fmt.Errorf("%w: per_page must be between 1 and 100", ErrInvalidPagination)
	}
	return nil
}

func ValidateParentID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: parent ID must be a non-empty numeric string", ErrInvalidParentID)
	}
	for _, char := range id {
		if char < '0' || char > '9' {
			return fmt.Errorf("%w: parent ID must be a non-empty numeric string", ErrInvalidParentID)
		}
	}
	return nil
}

func (s *Service) ListProvinces(ctx context.Context, page, perPage int) (dto.Page[dto.Province], error) {
	if err := ValidatePagination(page, perPage); err != nil {
		return dto.Page[dto.Province]{}, err
	}
	if err := s.validateClient(); err != nil {
		return dto.Page[dto.Province]{}, err
	}

	key := locationcache.ProvinceKey(page, perPage)
	var cached dto.Page[dto.Province]
	if readCache(ctx, s.Redis, key, &cached) {
		return cached, nil
	}

	result, err := s.Client.ListProvinces(ctx, page, perPage)
	if err != nil {
		return dto.Page[dto.Province]{}, err
	}
	writeCache(ctx, s.Redis, key, result)
	return result, nil
}

func (s *Service) ListCities(ctx context.Context, provinceID string, page, perPage int) (dto.Page[dto.City], error) {
	if err := ValidatePagination(page, perPage); err != nil {
		return dto.Page[dto.City]{}, err
	}
	if err := ValidateParentID(provinceID); err != nil {
		return dto.Page[dto.City]{}, err
	}
	if err := s.validateClient(); err != nil {
		return dto.Page[dto.City]{}, err
	}

	key := locationcache.CityKey(provinceID, page, perPage)
	var cached dto.Page[dto.City]
	if readCache(ctx, s.Redis, key, &cached) {
		return cached, nil
	}

	result, err := s.Client.ListCities(ctx, provinceID, page, perPage)
	if err != nil {
		return dto.Page[dto.City]{}, err
	}
	writeCache(ctx, s.Redis, key, result)
	return result, nil
}

func (s *Service) ListDistricts(ctx context.Context, regencyID string, page, perPage int) (dto.Page[dto.District], error) {
	if err := ValidatePagination(page, perPage); err != nil {
		return dto.Page[dto.District]{}, err
	}
	if err := ValidateParentID(regencyID); err != nil {
		return dto.Page[dto.District]{}, err
	}
	if err := s.validateClient(); err != nil {
		return dto.Page[dto.District]{}, err
	}

	key := locationcache.DistrictKey(regencyID, page, perPage)
	var cached dto.Page[dto.District]
	if readCache(ctx, s.Redis, key, &cached) {
		return cached, nil
	}

	result, err := s.Client.ListDistricts(ctx, regencyID, page, perPage)
	if err != nil {
		return dto.Page[dto.District]{}, err
	}
	writeCache(ctx, s.Redis, key, result)
	return result, nil
}

func (s *Service) ListVillages(ctx context.Context, districtID string, page, perPage int) (dto.Page[dto.Village], error) {
	if err := ValidatePagination(page, perPage); err != nil {
		return dto.Page[dto.Village]{}, err
	}
	if err := ValidateParentID(districtID); err != nil {
		return dto.Page[dto.Village]{}, err
	}
	if err := s.validateClient(); err != nil {
		return dto.Page[dto.Village]{}, err
	}

	key := locationcache.VillageKey(districtID, page, perPage)
	var cached dto.Page[dto.Village]
	if readCache(ctx, s.Redis, key, &cached) {
		return cached, nil
	}

	result, err := s.Client.ListVillages(ctx, districtID, page, perPage)
	if err != nil {
		return dto.Page[dto.Village]{}, err
	}
	writeCache(ctx, s.Redis, key, result)
	return result, nil
}

func (s *Service) validateClient() error {
	if s == nil || s.Client == nil {
		return ErrLocationClient
	}
	return nil
}

func readCache[T any](ctx context.Context, client *redis.Client, key string, out *T) bool {
	hit, err := sharedcache.GetJSON(ctx, client, key, out)
	if err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Location cache read failed; continuing with upstream request")
		return false
	}
	return hit
}

func writeCache[T any](ctx context.Context, client *redis.Client, key string, value T) {
	if err := sharedcache.SetJSON(ctx, client, key, value, locationcache.TTL()); err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Location cache write failed; continuing without cache")
	}
}
