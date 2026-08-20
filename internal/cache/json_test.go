package cache

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

type cacheTestValue struct {
	Name string `json:"name"`
}

func TestGetJSONReturnsDecodedCacheHit(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()
	mock.ExpectGet("cache:hit").SetVal(`{"name":"cached"}`)

	var got cacheTestValue
	hit, err := GetJSON(context.Background(), client, "cache:hit", &got)
	if err != nil {
		t.Fatalf("GetJSON() error = %v", err)
	}
	if !hit || got.Name != "cached" {
		t.Fatalf("expected decoded cache hit, hit=%v value=%#v", hit, got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestGetJSONTreatsRedisNilAsMiss(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()
	mock.ExpectGet("cache:miss").SetErr(redis.Nil)

	var got cacheTestValue
	hit, err := GetJSON(context.Background(), client, "cache:miss", &got)
	if err != nil {
		t.Fatalf("GetJSON() error = %v", err)
	}
	if hit {
		t.Fatal("expected cache miss")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestGetJSONReturnsMalformedPayloadError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()
	mock.ExpectGet("cache:malformed").SetVal(`{"name":`)

	var got cacheTestValue
	hit, err := GetJSON(context.Background(), client, "cache:malformed", &got)
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
	if hit {
		t.Fatal("expected malformed payload not to be a usable hit")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestJSONCacheHandlesNilRedis(t *testing.T) {
	var got cacheTestValue
	hit, err := GetJSON(context.Background(), nil, "cache:nil", &got)
	if err != nil || hit {
		t.Fatalf("expected nil client miss without error, hit=%v err=%v", hit, err)
	}
	if err := SetJSON(context.Background(), nil, "cache:nil", cacheTestValue{Name: "ignored"}, time.Minute); err != nil {
		t.Fatalf("expected nil client set no-op, got %v", err)
	}
}

func TestGetJSONReturnsRedisError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()
	mock.ExpectGet("cache:error").SetErr(errors.New("redis unavailable"))

	var got cacheTestValue
	if _, err := GetJSON(context.Background(), client, "cache:error", &got); err == nil {
		t.Fatal("expected Redis error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestSetJSONPreservesKeyAndTTL(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()
	key := "cache:exact:key"
	ttl := 17 * time.Minute
	mock.ExpectSet(key, `{"name":"stored"}`, ttl).SetVal("OK")

	if err := SetJSON(context.Background(), client, key, cacheTestValue{Name: "stored"}, ttl); err != nil {
		t.Fatalf("SetJSON() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func TestSetJSONRejectsNonPositiveTTLForRedis(t *testing.T) {
	for _, ttl := range []time.Duration{0, -time.Second} {
		client, mock := redismock.NewClientMock()
		err := SetJSON(context.Background(), client, "cache:invalid-ttl", cacheTestValue{Name: "ignored"}, ttl)
		_ = client.Close()
		if err == nil || !strings.Contains(err.Error(), "positive") {
			t.Fatalf("expected non-positive TTL %v validation error, got %v", ttl, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unexpected Redis command for TTL %v: %v", ttl, err)
		}
	}
}
