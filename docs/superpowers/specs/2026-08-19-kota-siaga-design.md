# Kota Siaga — Backend MVP Design

Date: 2026-08-19
Status: Proposed; implementation starts only after this document is reviewed.

## Objective

Build a public, read-only dashboard for one Indonesian village/city. The dashboard combines the selected location with:

- three-day weather forecast;
- active weather warnings for the province;
- the latest earthquakes;
- hospitals in the selected city/regency.

The backend keeps the API Indonesia key private, normalizes upstream responses, and caches repeated reads so the free API quota is usable for a portfolio demo.

## Deliberate scope

Included:

- a new `kota-siaga` application derived from `starter-kit`;
- thin location-selection proxy endpoints backed by API Indonesia;
- independent location, weather, warning, earthquake, and hospital endpoints;
- Redis-backed cache when Redis is available;
- input validation, upstream timeouts, rate limiting, structured errors, and tests.

Explicitly deferred:

- user registration, login, JWT, RBAC, sessions, and admin screens;
- push notifications or subscriptions;
- maps, routing, distance calculations, and nearest-hospital claims;
- historical analytics and local persistence of upstream data;
- prayer times, holidays, schools, and campuses.

These can be added when a real user workflow requires them. The MVP should prove the data integration first.

## Language boundary

The backend uses English consistently:

- package, file, type, function, variable, and JSON field names;
- comments, test names, logs, configuration documentation, and backend README content;
- API response `message` values and machine-readable error codes.

Backend errors stay stable and machine-readable so the frontend can translate them. The backend does not implement localization, accept a locale to change messages, or return bilingual labels.

The frontend owns Bahasa Indonesia and English translation resources and chooses the displayed language. Official values returned by API Indonesia—such as region names and weather descriptions—remain source data and are not renamed by the backend.

## Architecture

`kota-siaga` is a sibling project copied from `starter-kit`; the template itself remains unchanged. Reuse only the parts that lower maintenance:

- `utils.GetEnv` and `utils.DurationFromEnv` for typed environment configuration;
- the existing response envelope, log ID, logger, CORS, recovery, request context, and rate-limit middleware;
- the existing router and handler/service conventions;
- the existing Redis infrastructure for optional response caching.

Do not register or expose the starter-kit auth, RBAC, session, media, audit, configuration, or local-location routes. Auth- and database-dependent startup validation must not make `JWT_KEY` or PostgreSQL a requirement for this public application. No database table is needed for the MVP.

The fork does not run the starter-kit migrations. Dashboard readings and location data from API Indonesia are not persisted in PostgreSQL; Redis is only an optional cache.

The product is still a single Kota Siaga dashboard, but the backend is organized by data feature. There is no `siaga` mega-handler or mega-service and no aggregate endpoint that couples all upstream sources.

Request flow per feature:

```text
frontend
  -> GET /api/weather?adm4=...
  -> handler validation
  -> WeatherService
       -> API Indonesia client
       -> weather cache
  -> normalized weather response
```

The same flow applies independently to locations, warnings, earthquakes, and hospitals. The frontend calls the feature endpoints in parallel and can render the other cards when one source fails. The frontend never calls API Indonesia directly.

## Public API surface

Keep the route surface small:

- `GET /healthcheck`
- `GET /api/locations/province`
- `GET /api/locations/city?province_id=...`
- `GET /api/locations/district?kabupaten_id=...`
- `GET /api/locations/village?kecamatan_id=...`
- `GET /api/weather?adm4=...`
- `GET /api/warnings?provinsi=...`
- `GET /api/earthquakes/latest`
- `GET /api/hospitals?kabupaten_id=...&page=1&per_page=20`

The location routes are thin read-through proxies, not a replacement location database. Each route has its own handler and service and uses the same API Indonesia client. `adm4` is validated at the weather boundary as a non-empty code of at most 20 characters containing digits and dots. The frontend gets the province and city/regency IDs from the selected location; the hospital list is displayed as hospitals in the city/regency, not “nearest,” because the list contract does not provide enough information for a distance calculation.

Each feature returns an application-owned DTO rather than a raw upstream envelope. For example, weather owns only weather fields:

```json
{
  "data": [],
  "fetched_at": "2026-08-19T00:00:00Z"
}
```

An upstream failure is returned only for the affected feature. The dashboard decides how to display that card-level error.

## Reusable module boundaries

Each feature owns its handler, service, DTO, cache keys, mapper, and tests. Shared transport, configuration, response, and cache serialization remain outside feature services:

```text
internal/
  integrations/apiindonesia/
    client.go       shared base URL, x-api-key, timeout, JSON envelope, status errors
    locations.go    province, city, district, and village methods
    weather.go      weather endpoint methods
    warnings.go     active warning endpoint methods
    earthquakes.go  latest earthquake endpoint methods
    hospitals.go    hospital list endpoint methods
  cache/
    json.go         shared Redis JSON get/set helpers
    location/       location keys and TTLs
    weather/        weather keys and TTLs
    warning/        warning keys and TTLs
    earthquake/     earthquake keys and TTLs
    hospital/       hospital keys and TTLs
  dto/
    location.go
    weather.go
    warning.go
    earthquake.go
    hospital.go
  handlers/http/
    location/handler.go
    weather/handler.go
    warning/handler.go
    earthquake/handler.go
    hospital/handler.go
  services/
    location/service.go
    weather/service.go
    warning/service.go
    earthquake/service.go
    hospital/service.go
  mappers/
    location.go
    weather.go
    warning.go
    earthquake.go
    hospital.go
```

`apiindonesia.Client` owns all outbound HTTP details. Endpoint methods call one shared request helper; no service or handler creates raw upstream requests. The client uses the standard library `net/http` and the existing typed environment helper through one config loader. Do not add a generic HTTP framework or a broad provider abstraction for a single upstream.

The shared cache helper owns Redis serialization; each feature owns only its key and TTL policy. Location responses may be cached in Redis, but never written to PostgreSQL. Use separate source keys:

- weather: 6 hours or the upstream cache window;
- warnings: 1 hour;
- earthquakes: 1 hour;
- hospitals: 24 hours.

TTL values are configurable through environment variables using `GetEnv`/`DurationFromEnv`. A missing Redis client is a cache miss, not a request failure.

Use narrow interfaces only at dependency boundaries needed for tests (the API client and cache). Do not create interfaces for plain DTO mappers or one-line helpers. A feature service must not contain another feature's mapping or cache logic.

## API Indonesia integration

Use the documented REST base URL `https://use.apiindonesia.id` and `x-api-key` authentication.

- locations: the documented `/api/v1/wilayah/provinsi`, `/kabupaten`, `/kecamatan`, and `/kelurahan` endpoints;
- weather: `GET /api/v1/cuaca?adm4=...`;
- warnings: `GET /api/v1/peringatan-dini?provinsi=...`;
- earthquakes: `GET /api/v1/gempa/terkini`;
- hospitals: `GET /api/v1/rumah-sakit?kabupaten_id=...&page=1&per_page=20`.

Configuration:

- `API_INDONESIA_KEY` — required at startup;
- `API_INDONESIA_BASE_URL` — defaults to `https://use.apiindonesia.id`;
- `API_INDONESIA_TIMEOUT` — duration parsed by `GetEnv`, default `5s`;
- `LOCATION_CACHE_TTL`, `WEATHER_CACHE_TTL`, `WARNING_CACHE_TTL`, `EARTHQUAKE_CACHE_TTL`, `HOSPITAL_CACHE_TTL` — optional TTL overrides.

Default cache TTLs are `6h`, `1h`, `1h`, and `24h` respectively. The API key is validated during startup; a missing key prevents the service from starting.

The shared client must:

- attach the API key on every upstream data request;
- honor request context and a finite timeout;
- decode the documented `data`/`error` envelope;
- map non-2xx responses to a typed upstream error containing status and safe error code;
- never include the API key in logs or public errors.

The API contract and quota assumptions are based on the [API Indonesia documentation](https://docs.apiindonesia.id/).

## Error handling and resilience

- Missing or malformed feature parameters (`adm4`, `provinsi`, `kabupaten_id`, or pagination): `400` with the existing response envelope.
- Unknown location: `404` from the location feature.
- API key/configuration failure: fail startup; runtime upstream configuration errors return a safe `503`; never expose credentials.
- Upstream timeout, `429`, `402`, or `5xx`: log the safe upstream code and return an error for that feature only.
- Redis failure: log at warning level and continue with the upstream request.
- All external calls use the incoming request context and a finite timeout.
- Apply the existing IP rate-limit middleware to the public feature routes; default to 30 requests per minute per IP and make the limit/window configurable.

The endpoints are public, but public does not mean unprotected: input validation, secret handling, timeout, rate limiting, CORS, and safe error responses remain mandatory.

## Testing and validation

Add the smallest tests that protect the non-trivial behavior:

- `httptest.Server` tests for API key/header injection, envelope decoding, non-2xx mapping, and timeout behavior;
- cache tests for key names, TTL selection, cache hit, miss, malformed payload, nil Redis, and Redis errors;
- location proxy tests for query forwarding and normalized upstream responses;
- service tests for each feature's parameter mapping, successful response mapping, upstream failure, and upstream call parameters;
- handler tests for missing/invalid query values and normalized response/error status;
- one router smoke test that confirms only the public feature surface is registered.

Before implementation is considered complete, run:

```text
go test ./...
go vet ./...
make lint
```

## Acceptance criteria

The MVP is complete when:

1. a user can select a village through the API Indonesia-backed location proxy endpoints;
2. each feature endpoint returns its own normalized data when the upstream source is available;
3. a failed feature returns a safe feature-level error without coupling or disabling the other feature requests;
4. repeated requests use Redis when available and still work without Redis;
5. the API key never reaches the browser, logs, or public errors;
6. no auth/admin/local-location route, database dependency, or starter-kit migration is part of the deployed MVP;
7. all backend source, tests, logs, and API messages use English;
8. tests and static checks pass.
