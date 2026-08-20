package weatherservice

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"kota-siaga/internal/cache/weather"
	"kota-siaga/internal/dto"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

type upstreamFake struct {
	forecasts []dto.WeatherForecast
	err       error
	calls     int
	adm4      string
}

func (f *upstreamFake) GetWeather(_ context.Context, adm4 string) ([]dto.WeatherForecast, error) {
	f.calls++
	f.adm4 = adm4
	return f.forecasts, f.err
}

func TestValidateADM4RejectsMalformedValues(t *testing.T) {
	for _, value := range []string{
		"",
		"abc",
		"32.73.01.1001x",
		"123456789012345678901",
	} {
		if err := ValidateADM4(value); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
	if err := ValidateADM4("32.73.01.1001"); err != nil {
		t.Fatalf("expected valid adm4, got %v", err)
	}
}

func TestServiceReturnsThreeDayForecastAndCallsUpstreamOnce(t *testing.T) {
	forecasts := []dto.WeatherForecast{
		{ID: "1", Adm4: "32.73.01.1001", LocalDatetime: "2026-08-20T07:00:00+07:00"},
		{ID: "2", Adm4: "32.73.01.1001", LocalDatetime: "2026-08-21T07:00:00+07:00"},
		{ID: "3", Adm4: "32.73.01.1001", LocalDatetime: "2026-08-22T07:00:00+07:00"},
	}
	fake := &upstreamFake{forecasts: forecasts}
	service := NewService(fake, nil)

	got, err := service.GetWeather(context.Background(), "32.73.01.1001")
	if err != nil {
		t.Fatalf("GetWeather() error = %v", err)
	}
	if len(got) != 3 || got[2].ID != "3" {
		t.Fatalf("unexpected three-day result: %+v", got)
	}
	if fake.calls != 1 || fake.adm4 != "32.73.01.1001" {
		t.Fatalf("unexpected upstream calls: calls=%d adm4=%q", fake.calls, fake.adm4)
	}
}

func TestServiceUsesCachedForecastWithoutUpstreamCall(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := weathercache.Key("32.73.01.1001")
	mock.ExpectGet(key).SetVal(`[{"id":"cached","adm4":"32.73.01.1001","local_datetime":"2026-08-20T07:00:00+07:00"}]`)

	fake := &upstreamFake{err: errors.New("upstream must not be called")}
	service := NewService(fake, redisClient)
	got, err := service.GetWeather(context.Background(), "32.73.01.1001")
	if err != nil {
		t.Fatalf("cached request error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "cached" {
		t.Fatalf("unexpected cached result: %+v", got)
	}
	if fake.calls != 0 {
		t.Fatalf("cache hit called upstream %d times", fake.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestServiceFallsBackAfterRedisErrorAndCachesUpstreamResult(t *testing.T) {
	t.Setenv("WEATHER_CACHE_TTL", "90m")
	redisClient, mock := redismock.NewClientMock()
	defer redisClient.Close()
	key := weathercache.Key("32.73.01.1001")
	mock.ExpectGet(key).SetErr(errors.New("redis unavailable"))
	forecasts := []dto.WeatherForecast{{ID: "fresh", Adm4: "32.73.01.1001", TemperatureC: 24.5}}
	payload, err := json.Marshal(forecasts)
	if err != nil {
		t.Fatalf("marshal expected payload: %v", err)
	}
	mock.ExpectSet(key, string(payload), 90*time.Minute).SetVal("OK")

	service := NewService(&upstreamFake{forecasts: forecasts}, redisClient)
	got, err := service.GetWeather(context.Background(), "32.73.01.1001")
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
	key := weathercache.Key("32.73.01.1001")
	mock.ExpectGet(key).SetErr(redis.Nil)
	wantErr := errors.New("upstream body and API key stay private")

	service := NewService(&upstreamFake{err: wantErr}, redisClient)
	_, err := service.GetWeather(context.Background(), "32.73.01.1001")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected upstream error %v, got %v", wantErr, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected cache write after upstream failure: %v", err)
	}
}

func TestServicePropagatesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := NewService(&upstreamFake{err: context.Canceled}, nil)

	_, err := service.GetWeather(ctx, "32.73.01.1001")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
