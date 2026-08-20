# Kota Siaga Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the copied starter-kit into a public Go backend that exposes independent location, weather, warning, earthquake, and hospital features backed by API Indonesia.

**Architecture:** Keep only the useful starter-kit foundations: Gin, typed `utils.GetEnv`, response envelopes, logging, CORS, recovery, request IDs, Redis-backed rate limiting, and shared JSON cache helpers. Each feature owns its handler, service, DTO, upstream method, cache policy, mapper, and tests; no aggregate `siaga` service exists.

**Tech Stack:** Go 1.26, Gin, `net/http`, `encoding/json`, Redis, `httptest`, existing `pkg/response`, `pkg/logger`, `middlewares`, and `utils`.

No commit or push steps are included; the repository owner handles both.

---

## Task 1: Reduce the copied starter-kit to a public-only service

**Files:**

- Modify: `go.mod`
- Modify: `main.go`
- Modify: `main_test.go`
- Modify: `internal/router/router.go`
- Modify: `pkg/config/startup.go`
- Modify: `pkg/config/startup_test.go`
- Modify: `middlewares/request_logger.go`
- Modify: `middlewares/misc_test.go`
- Modify: `utils/misc_test.go`
- Modify: `.env.example`
- Modify: `README.md`
- Modify: `Dockerfile`
- Modify: `entrypoint.sh`
- Delete: `cmd/`
- Delete: `migrations/`
- Delete: `infrastructure/database/postgresql.go`
- Delete: `infrastructure/media/`
- Delete: `internal/authscope/`
- Delete: `internal/cache/location/`
- Delete: `internal/cache/permission/`
- Delete: `internal/domain/`
- Delete: `internal/interfaces/`
- Delete: `internal/repositories/`
- Delete: `internal/services/`
- Delete: `internal/handlers/http/appconfig/`
- Delete: `internal/handlers/http/audit/`
- Delete: `internal/handlers/http/media/`
- Delete: `internal/handlers/http/menu/`
- Delete: `internal/handlers/http/permission/`
- Delete: `internal/handlers/http/role/`
- Delete: `internal/handlers/http/session/`
- Delete: `internal/handlers/http/user/`
- Delete: `internal/handlers/http/common/audit.go`
- Delete: `internal/handlers/http/common/audit_test.go`
- Delete: `middlewares/auth.go`
- Delete: `middlewares/auth_test.go`
- Delete: `pkg/config/app_conf.go`
- Delete: `pkg/config/config_test.go`
- Delete: `pkg/config/otp.go`
- Delete: `pkg/config/password_reset.go`
- Delete: `pkg/mailer/`
- Delete: `pkg/moduleseed/`
- Delete: `pkg/security/`
- Delete: `pkg/storage/`
- Delete: `utils/audit.go`
- Delete: `utils/audit_test.go`
- Delete: `utils/jwt.go`
- Delete: `utils/jwt_test.go`

- [ ] **Step 1: Record the public startup contract in a failing config test.**

Replace the startup test cases with this contract:

```go
func TestValidateStartupConfigAcceptsPublicRuntime(t *testing.T) {
	t.Setenv("API_INDONESIA_KEY", "aip_live_test")
	t.Setenv("API_INDONESIA_BASE_URL", "https://use.apiindonesia.id")

	if err := ValidateStartupConfig("8080"); err != nil {
		t.Fatalf("expected public runtime config to pass: %v", err)
	}
}

func TestValidateStartupConfigRequiresAPIKey(t *testing.T) {
	t.Setenv("API_INDONESIA_KEY", "")

	if err := ValidateStartupConfig("8080"); err == nil {
		t.Fatal("expected missing API key to fail")
	}
}
```

- [ ] **Step 2: Run the focused test before changing startup code.**

Run: `go test ./pkg/config -run TestValidateStartupConfig -count=1`
Expected: FAIL because the copied starter-kit still requires JWT and database settings and does not validate `API_INDONESIA_KEY`.

- [ ] **Step 3: Rewrite startup and module wiring for the public runtime.**

Change the module declaration from `starter-kit` to `kota-siaga` and update Go imports to the new module path. `main.go` must:

1. load `.env`;
2. set the Jakarta timezone;
3. validate `PORT`, API Indonesia configuration, and optional Redis configuration;
4. initialize Redis without failing when Redis is unavailable;
5. create the shared API client;
6. create the router and register only feature routes;
7. run the HTTP server without migrations or PostgreSQL.

Remove `-migrate`, `DATABASE_URL`, `JWT_KEY`, and all auth/admin/media/session route registration. Keep `middlewares/request_logger.go` request-ID logging but remove its auth-scope lookup.

- [ ] **Step 4: Remove copied-only files and dependencies.**

Delete the listed auth, RBAC, database, migration, media, mailer, and storage paths. Keep `internal/handlers/http/common/json.go`, `pkg/response`, `pkg/logger`, `pkg/messages`, core `utils`, and the non-auth middleware. Run `go mod tidy` after the new feature imports exist; it must remove PostgreSQL, GORM, JWT, migration, mailer, and storage dependencies that no remaining package imports.

- [ ] **Step 5: Trim runtime configuration and container startup.**

`.env.example` must contain only English configuration for:

```text
APP_NAME=KOTA-SIAGA
APP_ENV=development
PORT=8080
GIN_MODE=release
API_INDONESIA_KEY=
API_INDONESIA_BASE_URL=https://use.apiindonesia.id
API_INDONESIA_TIMEOUT=5s
LOCATION_CACHE_TTL=24h
WEATHER_CACHE_TTL=6h
WARNING_CACHE_TTL=1h
EARTHQUAKE_CACHE_TTL=1h
HOSPITAL_CACHE_TTL=24h
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
PUBLIC_RATE_LIMIT=30
PUBLIC_RATE_WINDOW=1m
LOG_LEVEL=5
LOG_FORMAT=json
NODE=api
```

Remove migration copying from `Dockerfile` and make `entrypoint.sh` execute the binary directly. Rewrite `README.md` in English with the public routes, server-side API-key rule, local run command, and the backend/frontend language boundary.

- [ ] **Step 6: Run the baseline verification.**

Run: `go test ./...`
Expected: PASS with no database, migration, auth, or starter-kit route package remaining in the build graph.

Run: `go vet ./...`
Expected: PASS.

---

## Task 2: Add shared API configuration, upstream client, and Redis JSON cache

**Files:**

- Create: `pkg/config/api_indonesia.go`
- Modify: `pkg/config/startup.go`
- Create: `pkg/config/api_indonesia_test.go`
- Create: `internal/integrations/apiindonesia/client.go`
- Create: `internal/integrations/apiindonesia/client_test.go`
- Create: `internal/cache/json.go`
- Create: `internal/cache/json_test.go`
- Create: `internal/dto/page.go`

- [ ] **Step 1: Define the shared types and config contract.**

Use English names and the existing typed environment helper:

```go
type APIIndonesiaConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func LoadAPIIndonesiaConfig() APIIndonesiaConfig
```

Defaults: base URL `https://use.apiindonesia.id`, timeout `5s`. `API_INDONESIA_KEY` is required by startup validation. Define a shared page type:

```go
type Page[T any] struct {
	Data       []T  `json:"data"`
	Total      int  `json:"total"`
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	TotalPages int  `json:"total_pages"`
}
```

- [ ] **Step 2: Write failing client tests using `httptest.Server`.**

Cover these exact cases:

- `GET` uses the configured base URL and path;
- every request sends `x-api-key`;
- a successful `{ "data": ... }` envelope decodes into the caller output;
- a non-2xx `{ "error": { "code": "QUOTA_EXCEEDED" } }` response returns an `UpstreamError` with status and code but no secret;
- the request context cancels a slow server;
- query values are encoded with `url.Values`.

- [ ] **Step 3: Implement the smallest shared client.**

Use `net/http` only:

```go
type Client struct {
	BaseURL    *url.URL
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(config APIIndonesiaConfig) (*Client, error)
func (c *Client) GetJSON(ctx context.Context, path string, query url.Values, out any) error
```

`GetJSON` must create a request with context, add `x-api-key`, set `Accept: application/json`, enforce the configured timeout through `HTTPClient`, decode the shared envelope, and return safe typed errors. No feature service may create raw upstream requests.

- [ ] **Step 4: Write cache tests before implementation.**

Use `redismock` to cover cache hit, cache miss, malformed JSON, nil Redis, Redis error, key preservation, and configured TTL. The cache helper must never turn Redis failure into an upstream feature failure.

- [ ] **Step 5: Implement shared JSON cache helpers.**

Create generic helpers with one responsibility:

```go
func GetJSON[T any](ctx context.Context, client *redis.Client, key string, out *T) (bool, error)
func SetJSON(ctx context.Context, client *redis.Client, key string, value any, ttl time.Duration) error
```

Feature cache packages will own only key construction and TTL loading through `utils.GetEnv`; serialization and Redis timeouts stay in this shared file.

- [ ] **Step 6: Run shared-package checks.**

Run: `go test ./pkg/config ./internal/integrations/apiindonesia ./internal/cache -count=1`
Expected: PASS.

---

## Task 3: Implement the location feature

**Files:**

- Create: `internal/integrations/apiindonesia/locations.go`
- Create: `internal/dto/location.go`
- Create: `internal/cache/location/cache.go`
- Create: `internal/cache/location/cache_test.go`
- Create: `internal/mappers/location.go`
- Create: `internal/services/location/service.go`
- Create: `internal/services/location/service_test.go`
- Create: `internal/handlers/http/location/handler.go`
- Create: `internal/handlers/http/location/handler_test.go`

- [ ] **Step 1: Define the upstream and application contracts.**

Implement client methods for the documented endpoints:

```go
ListProvinces(ctx context.Context, page, perPage int) (Page[Province], error)
ListCities(ctx context.Context, provinceID string, page, perPage int) (Page[City], error)
ListDistricts(ctx context.Context, regencyID string, page, perPage int) (Page[District], error)
ListVillages(ctx context.Context, districtID string, page, perPage int) (Page[Village], error)
```

Map upstream fields such as `province_id`, `regency_id`, and `postal_code` into English application DTOs. Preserve official `name` and `alt_name` values unchanged.

- [ ] **Step 2: Write service tests for query forwarding and cache behavior.**

Test one successful call per level, one invalid parent ID, one cache hit, and one upstream failure. The service must validate IDs as non-empty numeric strings before calling the client.

- [ ] **Step 3: Implement location cache, service, and handlers.**

Expose:

```text
GET /api/locations/province?page=1&per_page=20
GET /api/locations/city?province_id=32&page=1&per_page=20
GET /api/locations/district?kabupaten_id=3273&page=1&per_page=20
GET /api/locations/village?kecamatan_id=3273010&page=1&per_page=20
```

Use the existing response envelope with English messages and stable error codes. Cache keys must include the endpoint and all query inputs, for example `locations:city:32:1:20`.

- [ ] **Step 4: Run the location tests.**

Run: `go test ./internal/integrations/apiindonesia ./internal/cache/location ./internal/services/location ./internal/handlers/http/location -count=1`
Expected: PASS.

---

## Task 4: Implement the weather feature

**Files:**

- Create: `internal/integrations/apiindonesia/weather.go`
- Create: `internal/dto/weather.go`
- Create: `internal/cache/weather/cache.go`
- Create: `internal/cache/weather/cache_test.go`
- Create: `internal/mappers/weather.go`
- Create: `internal/services/weather/service.go`
- Create: `internal/services/weather/service_test.go`
- Create: `internal/handlers/http/weather/handler.go`
- Create: `internal/handlers/http/weather/handler_test.go`

- [ ] **Step 1: Define the weather DTO and source method.**

Implement `GetWeather(ctx context.Context, adm4 string) ([]WeatherForecast, error)` for `GET /api/v1/cuaca?adm4=...`. Use English field names such as `Adm4`, `LocalDatetime`, `Weather`, `WeatherCode`, `TemperatureC`, `HumidityPercent`, `PrecipitationMM`, `WindSpeed`, and `Source`.

- [ ] **Step 2: Write failing service and handler tests.**

Cover malformed `adm4`, a successful three-day response, cache hit, upstream `429`, and a context timeout. The handler must return `400` for invalid input and an English stable error code for upstream failure.

- [ ] **Step 3: Implement the weather feature.**

Expose `GET /api/weather?adm4=32.73.01.1001`. Cache by `weather:<adm4>` with a default `6h` TTL from `WEATHER_CACHE_TTL`. Keep upstream-to-application field mapping in `internal/mappers/weather.go` so the service remains orchestration-only.

- [ ] **Step 4: Run the weather tests.**

Run: `go test ./internal/services/weather ./internal/handlers/http/weather -count=1`
Expected: PASS.

---

## Task 5: Implement the warning feature

**Files:**

- Create: `internal/integrations/apiindonesia/warnings.go`
- Create: `internal/dto/warning.go`
- Create: `internal/cache/warning/cache.go`
- Create: `internal/cache/warning/cache_test.go`
- Create: `internal/mappers/warning.go`
- Create: `internal/services/warning/service.go`
- Create: `internal/services/warning/service_test.go`
- Create: `internal/handlers/http/warning/handler.go`
- Create: `internal/handlers/http/warning/handler_test.go`

- [ ] **Step 1: Define the warning contract.**

Implement `ListWarnings(ctx context.Context, province string) ([]Warning, error)` for `GET /api/v1/peringatan-dini?provinsi=...`. Use English fields `ID`, `Event`, `Urgency`, `Severity`, `Certainty`, `Area`, `Province`, `Effective`, `Expires`, `Headline`, `Description`, `Instruction`, `Source`, and `IsActive`.

- [ ] **Step 2: Write tests for province validation, successful mapping, cache, and upstream errors.**

The service must reject an empty province before making an HTTP call. The handler must expose `GET /api/warnings?provinsi=Jawa+Barat` with English response metadata while preserving the official warning text.

- [ ] **Step 3: Implement warning cache and HTTP surface.**

Cache by normalized province with a default `1h` TTL from `WARNING_CACHE_TTL`. Use the shared endpoint rate limiter and response envelope.

- [ ] **Step 4: Run the warning tests.**

Run: `go test ./internal/services/warning ./internal/handlers/http/warning -count=1`
Expected: PASS.

---

## Task 6: Implement the earthquake feature

**Files:**

- Create: `internal/integrations/apiindonesia/earthquakes.go`
- Create: `internal/dto/earthquake.go`
- Create: `internal/cache/earthquake/cache.go`
- Create: `internal/cache/earthquake/cache_test.go`
- Create: `internal/mappers/earthquake.go`
- Create: `internal/services/earthquake/service.go`
- Create: `internal/services/earthquake/service_test.go`
- Create: `internal/handlers/http/earthquake/handler.go`
- Create: `internal/handlers/http/earthquake/handler_test.go`

- [ ] **Step 1: Define the latest-earthquake contract.**

Implement `ListLatest(ctx context.Context) ([]Earthquake, error)` for `GET /api/v1/gempa/terkini`. Map `datetime`, `magnitude`, `depth_km`, `lat`, `lng`, `region`, `potential`, `is_felt`, `felt_areas`, and `source` to English field names.

- [ ] **Step 2: Write tests for successful mapping, cache hit, and upstream failure.**

The route has no user query parameters. A malformed upstream payload must return a safe `502`-class response and must not be cached.

- [ ] **Step 3: Implement the feature.**

Expose `GET /api/earthquakes/latest`. Cache the latest list under a fixed feature key with a default `1h` TTL from `EARTHQUAKE_CACHE_TTL`.

- [ ] **Step 4: Run the earthquake tests.**

Run: `go test ./internal/services/earthquake ./internal/handlers/http/earthquake -count=1`
Expected: PASS.

---

## Task 7: Implement the hospital feature

**Files:**

- Create: `internal/integrations/apiindonesia/hospitals.go`
- Create: `internal/dto/hospital.go`
- Create: `internal/cache/hospital/cache.go`
- Create: `internal/cache/hospital/cache_test.go`
- Create: `internal/mappers/hospital.go`
- Create: `internal/services/hospital/service.go`
- Create: `internal/services/hospital/service_test.go`
- Create: `internal/handlers/http/hospital/handler.go`
- Create: `internal/handlers/http/hospital/handler_test.go`

- [ ] **Step 1: Define pagination and hospital contracts.**

Implement `ListHospitals(ctx context.Context, regencyID string, page, perPage int) (Page[Hospital], error)` for `GET /api/v1/rumah-sakit`. Validate `kabupaten_id` as numeric, `page >= 1`, and `1 <= per_page <= 200`.

Map the upstream fields to `ID`, `Name`, `Type`, `Class`, `Ownership`, `Address`, `PostalCode`, `Phone`, `BedsTotal`, `ICUBeds`, `ProvinceName`, `RegencyName`, and `IsActive`.

- [ ] **Step 2: Write tests for validation, pagination mapping, cache key isolation, and upstream failure.**

Cache keys must include `kabupaten_id`, `page`, and `per_page`, so one city or page cannot return another city's result.

- [ ] **Step 3: Implement the hospital feature.**

Expose `GET /api/hospitals?kabupaten_id=3273&page=1&per_page=20`. Use a default `24h` TTL from `HOSPITAL_CACHE_TTL`. Do not call the result “nearest”; this endpoint is a city/regency directory only.

- [ ] **Step 4: Run the hospital tests.**

Run: `go test ./internal/services/hospital ./internal/handlers/http/hospital -count=1`
Expected: PASS.

---

## Task 8: Wire routes, rate limits, and backend language policy

**Files:**

- Modify: `internal/router/router.go`
- Create or modify: `internal/router/router_test.go`
- Modify: `main.go`
- Modify: `middlewares/rate_limiter.go` only if the existing middleware cannot be reused as-is
- Modify: `README.md`
- Modify: `.env.example`

- [ ] **Step 1: Write the route registration test.**

Assert that `NewRoutes` registers exactly these public paths:

```text
GET /healthcheck
GET /api/locations/province
GET /api/locations/city
GET /api/locations/district
GET /api/locations/village
GET /api/weather
GET /api/warnings
GET /api/earthquakes/latest
GET /api/hospitals
```

Also assert that copied starter-kit paths such as `/api/user/login`, `/api/roles`, `/api/location/sync`, and `/api/media` are absent.

- [ ] **Step 2: Wire one shared client and Redis instance into the feature constructors.**

Use constructor injection so handlers do not read environment variables and services do not create clients. Register each feature independently:

```go
location.Register(routes.App, client, redisClient)
weather.Register(routes.App, client, redisClient)
warning.Register(routes.App, client, redisClient)
earthquake.Register(routes.App, client, redisClient)
hospital.Register(routes.App, client, redisClient)
```

If a feature constructor needs a dependency not used by another feature, keep it local to that feature. Do not create a `KotaSiagaService` or aggregate route.

- [ ] **Step 3: Apply the shared public rate limiter.**

Use `PUBLIC_RATE_LIMIT` and `PUBLIC_RATE_WINDOW` through `utils.GetEnv`, defaulting to 30 requests per minute per IP. Redis failure must preserve the existing fail-open behavior and log the failure in English.

- [ ] **Step 4: Enforce the English backend contract.**

Check all new and retained backend files for English package names, identifiers, comments, tests, logs, response messages, and error codes. Keep official API values unchanged. Do not add locale handling or bilingual response text; the future frontend owns Indonesian/English translations.

- [ ] **Step 5: Run route and full checks.**

Run: `go test ./internal/router ./middlewares -count=1`
Expected: PASS.

Run: `go test ./...`
Expected: PASS.

Run: `go vet ./...`
Expected: PASS.

Run: `make lint`
Expected: `golangci-lint passed.`

---

## Self-review checklist

- Spec coverage: Tasks 1–2 cover public-only bootstrap, API-key security, `GetEnv`, shared HTTP, timeouts, Redis behavior, and English backend messages. Tasks 3–7 cover each independent API feature. Task 8 covers route registration, rate limiting, bilingual boundary, and final verification.
- No aggregate service or `/api/siaga` route is planned.
- No PostgreSQL, migration, authentication, RBAC, session, media, or local-location persistence remains in the runtime plan.
- All feature services have separate files and tests; shared code is limited to transport, configuration, response, and cache serialization.
- All commands and expected outputs are explicit.
