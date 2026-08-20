package database

import (
	"strings"
	"testing"
)

func TestInitRedisReturnsMalformedURLWithoutFallback(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://%")

	client, err := InitRedis()
	if err == nil || !strings.Contains(err.Error(), "REDIS_URL") {
		if client != nil {
			_ = client.Close()
		}
		t.Fatalf("expected malformed REDIS_URL error, client=%#v err=%v", client, err)
	}
	if client != nil {
		t.Fatalf("expected nil client on malformed URL, got %#v", client)
	}
}

func TestInitRedisReturnsPingError(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://127.0.0.1:1/0")

	client, err := InitRedis()
	if err == nil {
		if client != nil {
			_ = client.Close()
		}
		t.Fatal("expected redis connection error")
	}
	if client != nil {
		t.Fatalf("expected nil client on error, got %#v", client)
	}
}
