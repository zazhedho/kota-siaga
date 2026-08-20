package earthquake

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	earthquakecache "kota-siaga/internal/cache/earthquake"
	"kota-siaga/internal/dto"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

type upstreamFake struct {
	earthquakes []dto.Earthquake
	err         error
	calls       int
}

func (f *upstreamFake) ListLatest(_ context.Context) ([]dto.Earthquake, error) {
	f.calls++
	return f.earthquakes, f.err
}

func TestServiceReturnsLatestEarthquakesWithoutQueryValidation(t *testing.T) {
	earthquakes := []dto.Earthquake{{ID: "fresh", Region: "South of Java"}}
	fake := &upstreamFake{earthquakes: earthquakes}
	service := NewService(fake, nil)

	got, err := service.ListLatest(context.Background())
	if err != nil {
		t.Fatalf("ListLatest() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "fresh" {
		t.Fatalf("unexpected latest earthquakes: %#v", got)
	}
	if fake.calls != 1 {
		t.Fatalf("expected one upstream call, got %d", fake.calls)
	}
}

func TestServiceUsesCachedEarthquakesWithoutUpstreamCall(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	mock.ExpectGet(earthquakecache.Key()).SetVal(`[{"id":"cached","region":"Central Java"}]`)

	fake := &upstreamFake{err: errors.New("upstream must not be called")}
	service := NewService(fake, redisClient)
	got, err := service.ListLatest(context.Background())
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

func TestServiceFailsOpenAfterRedisReadErrorAndCachesSuccessfulResult(t *testing.T) {
	t.Setenv("EARTHQUAKE_CACHE_TTL", "90m")
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := earthquakecache.Key()
	mock.ExpectGet(key).SetErr(errors.New("redis unavailable"))
	earthquakes := []dto.Earthquake{{ID: "fresh", Magnitude: 5.4}}
	payload, err := json.Marshal(earthquakes)
	if err != nil {
		t.Fatalf("marshal expected payload: %v", err)
	}
	mock.ExpectSet(key, string(payload), 90*time.Minute).SetVal("OK")

	service := NewService(&upstreamFake{earthquakes: earthquakes}, redisClient)
	got, err := service.ListLatest(context.Background())
	if err != nil || len(got) != 1 || got[0].ID != "fresh" {
		t.Fatalf("unexpected Redis fallback result: got=%+v err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestServiceFailsOpenAfterRedisWriteError(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := earthquakecache.Key()
	mock.ExpectGet(key).SetErr(redis.Nil)
	earthquakes := []dto.Earthquake{{ID: "fresh"}}
	payload, err := json.Marshal(earthquakes)
	if err != nil {
		t.Fatalf("marshal expected payload: %v", err)
	}
	mock.ExpectSet(key, string(payload), time.Hour).SetErr(errors.New("redis unavailable"))

	service := NewService(&upstreamFake{earthquakes: earthquakes}, redisClient)
	got, err := service.ListLatest(context.Background())
	if err != nil || len(got) != 1 || got[0].ID != "fresh" {
		t.Fatalf("expected fresh result despite Redis write failure: got=%+v err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestServiceDoesNotCacheUpstreamFailure(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := earthquakecache.Key()
	mock.ExpectGet(key).SetErr(redis.Nil)
	wantErr := errors.New("upstream body and API key stay private")

	service := NewService(&upstreamFake{err: wantErr}, redisClient)
	_, err := service.ListLatest(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected upstream error %v, got %v", wantErr, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected cache write after upstream failure: %v", err)
	}
}

func TestServiceReturnsMissingClientSentinel(t *testing.T) {
	service := NewService(nil, nil)
	_, err := service.ListLatest(context.Background())
	if !errors.Is(err, ErrEarthquakeClient) {
		t.Fatalf("expected ErrEarthquakeClient, got %v", err)
	}
}
