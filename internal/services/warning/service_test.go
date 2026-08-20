package warningservice

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	warningcache "kota-siaga/internal/cache/warning"
	"kota-siaga/internal/dto"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

type upstreamFake struct {
	warnings []dto.Warning
	err      error
	calls    int
	province string
}

func (f *upstreamFake) ListWarnings(_ context.Context, province string) ([]dto.Warning, error) {
	f.calls++
	f.province = province
	return f.warnings, f.err
}

func TestServiceRejectsEmptyProvinceBeforeUpstream(t *testing.T) {
	fake := &upstreamFake{}
	service := NewService(fake, nil)

	for _, province := range []string{"", "   "} {
		_, err := service.ListWarnings(context.Background(), province)
		if !errors.Is(err, ErrInvalidProvince) {
			t.Fatalf("province %q: expected ErrInvalidProvince, got %v", province, err)
		}
	}
	if fake.calls != 0 {
		t.Fatalf("invalid province reached upstream %d times", fake.calls)
	}
}

func TestServiceTrimsProvinceForUpstreamAndReturnsFreshWarnings(t *testing.T) {
	warnings := []dto.Warning{{ID: "fresh", Province: "Jawa Barat", IsActive: true}}
	fake := &upstreamFake{warnings: warnings}
	service := NewService(fake, nil)

	got, err := service.ListWarnings(context.Background(), "  Jawa Barat  ")
	if err != nil {
		t.Fatalf("ListWarnings() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "fresh" {
		t.Fatalf("unexpected warnings: %#v", got)
	}
	if fake.calls != 1 || fake.province != "Jawa Barat" {
		t.Fatalf("unexpected upstream call: calls=%d province=%q", fake.calls, fake.province)
	}
}

func TestServiceReturnsMissingClientSentinel(t *testing.T) {
	service := NewService(nil, nil)

	_, err := service.ListWarnings(context.Background(), "Jawa Barat")
	if !errors.Is(err, ErrWarningClient) {
		t.Fatalf("expected ErrWarningClient, got %v", err)
	}
}

func TestServiceUsesCachedWarningsWithoutUpstreamCall(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := warningcache.Key("Jawa Barat")
	mock.ExpectGet(key).SetVal(`[{"id":"cached","province":"Jawa Barat","is_active":true}]`)

	fake := &upstreamFake{err: errors.New("upstream must not be called")}
	service := NewService(fake, redisClient)
	got, err := service.ListWarnings(context.Background(), " Jawa Barat ")
	if err != nil {
		t.Fatalf("cached request error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "cached" {
		t.Fatalf("unexpected cached result: %#v", got)
	}
	if fake.calls != 0 {
		t.Fatalf("cache hit called upstream %d times", fake.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestServiceFallsBackAfterRedisErrorAndCachesUpstreamResult(t *testing.T) {
	t.Setenv("WARNING_CACHE_TTL", "90m")
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := warningcache.Key("Jawa Barat")
	mock.ExpectGet(key).SetErr(errors.New("redis unavailable"))
	warnings := []dto.Warning{{ID: "fresh", Province: "Jawa Barat", IsActive: true}}
	payload, err := json.Marshal(warnings)
	if err != nil {
		t.Fatalf("marshal expected payload: %v", err)
	}
	mock.ExpectSet(key, string(payload), 90*time.Minute).SetVal("OK")

	service := NewService(&upstreamFake{warnings: warnings}, redisClient)
	got, err := service.ListWarnings(context.Background(), " Jawa Barat ")
	if err != nil || len(got) != 1 || got[0].ID != "fresh" {
		t.Fatalf("unexpected Redis fallback result: got=%+v err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestServiceDoesNotCacheUpstreamFailure(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := warningcache.Key("Jawa Barat")
	mock.ExpectGet(key).SetErr(redis.Nil)
	wantErr := errors.New("upstream body and API key stay private")

	service := NewService(&upstreamFake{err: wantErr}, redisClient)
	_, err := service.ListWarnings(context.Background(), "Jawa Barat")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected upstream error %v, got %v", wantErr, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected cache write after upstream failure: %v", err)
	}
}
