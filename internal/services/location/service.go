package locationservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	sharedcache "kota-siaga/internal/cache"
	locationcache "kota-siaga/internal/cache/location"
	"kota-siaga/internal/dto"
	"kota-siaga/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidPagination   = errors.New("invalid pagination")
	ErrInvalidParentID     = errors.New("invalid parent ID")
	ErrInvalidSearchQuery  = errors.New("invalid location search query")
	ErrInvalidSearchLimit  = errors.New("invalid location search limit")
	ErrInvalidLocationCode = errors.New("invalid location code")
	ErrLocationClient      = errors.New("location upstream client is not configured")
)

const (
	DefaultSearchLimit = 10
	MaxSearchLimit     = 10
	MinSearchLength    = 3
)

type UpstreamClient interface {
	ListProvinces(context.Context, int, int) (dto.Page[dto.Province], error)
	ListCities(context.Context, string, int, int) (dto.Page[dto.City], error)
	ListDistricts(context.Context, string, int, int) (dto.Page[dto.District], error)
	ListVillages(context.Context, string, int, int) (dto.Page[dto.Village], error)
	SearchLocations(context.Context, string, int) ([]dto.LocationSearchItem, error)
	ResolveLocation(context.Context, string) (dto.LocationPath, error)
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

func ValidateSearchQuery(query string) error {
	if utf8.RuneCountInString(strings.TrimSpace(query)) < MinSearchLength {
		return fmt.Errorf("%w: query must contain at least %d characters", ErrInvalidSearchQuery, MinSearchLength)
	}
	return nil
}

func ValidateSearchLimit(limit int) error {
	if limit < 1 || limit > MaxSearchLimit {
		return fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidSearchLimit, MaxSearchLimit)
	}
	return nil
}

func ValidateLocationCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("%w: code is required", ErrInvalidLocationCode)
	}

	if strings.Contains(code, ".") {
		parts := strings.Split(code, ".")
		if len(parts) != 3 && len(parts) != 4 {
			return fmt.Errorf("%w: code must contain three or four numeric segments", ErrInvalidLocationCode)
		}
		for _, part := range parts {
			if !isNumeric(part) {
				return fmt.Errorf("%w: code must contain only numeric segments", ErrInvalidLocationCode)
			}
		}
		return nil
	}

	if (len(code) != 6 && len(code) != 7 && len(code) != 10) || !isNumeric(code) {
		return fmt.Errorf("%w: code must be a district or village code", ErrInvalidLocationCode)
	}
	return nil
}

func ValidateVillageCode(code string) error {
	return ValidateLocationCode(code)
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func (s *Service) SearchLocations(ctx context.Context, query string, limit int) ([]dto.LocationSearchItem, error) {
	if err := ValidateSearchQuery(query); err != nil {
		return nil, err
	}
	if err := ValidateSearchLimit(limit); err != nil {
		return nil, err
	}
	if err := s.validateClient(); err != nil {
		return nil, err
	}

	items, err := s.Client.SearchLocations(ctx, strings.TrimSpace(query), limit)
	if err != nil {
		return nil, err
	}

	results := make([]dto.LocationSearchItem, 0, len(items))
	for _, item := range items {
		level := strings.ToLower(strings.TrimSpace(item.Level))
		if level != "district" && level != "village" {
			continue
		}
		results = append(results, item)
		if len(results) == limit {
			break
		}
	}
	return results, nil
}

func (s *Service) ResolveLocation(ctx context.Context, code string) (dto.LocationPath, error) {
	if err := ValidateLocationCode(code); err != nil {
		return dto.LocationPath{}, err
	}
	if err := s.validateClient(); err != nil {
		return dto.LocationPath{}, err
	}
	return s.Client.ResolveLocation(ctx, strings.TrimSpace(code))
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
