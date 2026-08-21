package hospitalservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	sharedcache "kota-siaga/internal/cache"
	hospitalcache "kota-siaga/internal/cache/hospital"
	"kota-siaga/internal/dto"
	"kota-siaga/pkg/logger"

	"github.com/redis/go-redis/v9"
)

const maxSearchLength = 100

var (
	ErrInvalidPagination  = errors.New("invalid pagination")
	ErrInvalidKabupatenID = errors.New("invalid kabupaten ID")
	ErrInvalidSearch      = errors.New("invalid hospital search")
	ErrHospitalClient     = errors.New("hospital upstream client is not configured")
)

type UpstreamClient interface {
	ListHospitals(context.Context, string, string, int, int) (dto.Page[dto.Hospital], error)
}

type Service struct {
	Client UpstreamClient
	Redis  *redis.Client
}

type HospitalService = Service

func NewService(client UpstreamClient, redisClient *redis.Client) *Service {
	return &Service{Client: client, Redis: redisClient}
}

func NewHospitalService(client UpstreamClient, redisClient *redis.Client) *Service {
	return NewService(client, redisClient)
}

func ValidatePagination(page, perPage int) error {
	if page < 1 {
		return fmt.Errorf("%w: page must be at least 1", ErrInvalidPagination)
	}
	if perPage < 1 || perPage > 200 {
		return fmt.Errorf("%w: per_page must be between 1 and 200", ErrInvalidPagination)
	}
	return nil
}

func ValidateKabupatenID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: kabupaten ID must be a non-empty numeric string", ErrInvalidKabupatenID)
	}
	for _, char := range id {
		if char < '0' || char > '9' {
			return fmt.Errorf("%w: kabupaten ID must be a non-empty numeric string", ErrInvalidKabupatenID)
		}
	}
	return nil
}

func NormalizeSearch(value string) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > maxSearchLength {
		return "", fmt.Errorf("%w: search must be 100 characters or fewer", ErrInvalidSearch)
	}
	return value, nil
}

func (s *Service) ListHospitals(ctx context.Context, kabupatenID, search string, page, perPage int) (dto.Page[dto.Hospital], error) {
	search, err := NormalizeSearch(search)
	if err != nil {
		return dto.Page[dto.Hospital]{}, err
	}
	if err := ValidateKabupatenID(kabupatenID); err != nil {
		return dto.Page[dto.Hospital]{}, err
	}
	if err := ValidatePagination(page, perPage); err != nil {
		return dto.Page[dto.Hospital]{}, err
	}
	if s == nil || s.Client == nil {
		return dto.Page[dto.Hospital]{}, ErrHospitalClient
	}

	key := hospitalcache.Key(kabupatenID, search, page, perPage)
	var cached dto.Page[dto.Hospital]
	hit, err := sharedcache.GetJSON(ctx, s.Redis, key, &cached)
	if err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Hospital cache read failed; continuing with upstream request")
	} else if hit {
		return cached, nil
	}

	result, err := s.Client.ListHospitals(ctx, kabupatenID, search, page, perPage)
	if err != nil {
		return dto.Page[dto.Hospital]{}, err
	}
	if err := sharedcache.SetJSON(ctx, s.Redis, key, result, hospitalcache.TTL()); err != nil {
		logger.WriteLog(logger.LogLevelWarn, "Hospital cache write failed; continuing without cache")
	}
	return result, nil
}
