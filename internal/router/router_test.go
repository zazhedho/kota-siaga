package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	redismock "github.com/go-redis/redismock/v9"

	"github.com/gin-gonic/gin"
)

const routerAtomicRateLimitScript = `
local current = redis.call("INCR", KEYS[1])
if current == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current`

func TestNewRoutesRegistersExactPublicRoutes(t *testing.T) {
	routes := NewRoutes(nil, nil)
	registered := map[string]bool{}
	for _, route := range routes.App.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	want := map[string]bool{
		"GET /healthcheck":            true,
		"GET /api/locations/province": true,
		"GET /api/locations/city":     true,
		"GET /api/locations/district": true,
		"GET /api/locations/village":  true,
		"GET /api/weather":            true,
		"GET /api/warnings":           true,
		"GET /api/earthquakes/latest": true,
		"GET /api/hospitals":          true,
	}
	if len(registered) != len(want) {
		t.Fatalf("expected %d routes, got %d: %v", len(want), len(registered), registered)
	}
	for route := range want {
		if !registered[route] {
			t.Fatalf("expected route %s to be registered", route)
		}
	}
}

func TestNewRoutesExcludesCopiedStarterKitRoutes(t *testing.T) {
	routes := NewRoutes(nil, nil)
	for _, copiedPath := range []string{
		"/api/user/login",
		"/api/roles",
		"/api/location/sync",
		"/api/media",
	} {
		for _, route := range routes.App.Routes() {
			if route.Path == copiedPath {
				t.Fatalf("copied route %s must not be registered", copiedPath)
			}
		}
	}
}

func TestNewRoutesRegistersHealthcheck(t *testing.T) {
	routes := NewRoutes(nil, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	routes.App.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestNewRoutesUsesCustomRecoveryForPanics(t *testing.T) {
	routes := NewRoutes(nil, nil)
	routes.App.GET("/panic", func(_ *gin.Context) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	routes.App.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected JSON recovery response, got headers=%v", rec.Header())
	}
	if !strings.Contains(rec.Body.String(), `"status":false`) {
		t.Fatalf("expected custom recovery envelope, got %s", rec.Body.String())
	}
}

func TestNewRoutesUsesRemoteAddrForPublicRateLimitIdentity(t *testing.T) {
	t.Setenv("PUBLIC_RATE_LIMIT", "1")
	t.Setenv("PUBLIC_RATE_WINDOW", "1m")

	client, mock := redismock.NewClientMock()
	key := "rate_limit:public:1.2.3.4"
	mock.ExpectEval(routerAtomicRateLimitScript, []string{key}, time.Minute.Milliseconds()).SetVal(int64(1))

	routes := NewRoutes(client, nil)
	routes.App.GET("/identity", func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	if routes.App.ForwardedByClientIP {
		t.Fatal("expected forwarded client IP trust to be disabled")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/identity", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	req.Header.Set("X-Forwarded-For", "9.9.9.9")
	routes.App.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}
