# Kota Siaga Backend

Public Go backend foundation for Kota Siaga. It keeps Gin, configuration, response envelopes, logging, CORS, recovery, request IDs, optional Redis, rate limiting, and shared utilities. PostgreSQL, migrations, authentication, RBAC, sessions, media, mail, audit, and copied local-location persistence are not part of this service.

## Public routes

The application currently exposes these public GET routes:

- `GET /healthcheck`
- `GET /api/locations/province?page=1&per_page=20`
- `GET /api/locations/city?province_id=32&page=1&per_page=20`
- `GET /api/locations/district?kabupaten_id=3273&page=1&per_page=20`
- `GET /api/locations/village?kecamatan_id=3273010&page=1&per_page=20`
- `GET /api/weather?adm4=32.73.01.1001`
- `GET /api/warnings?provinsi=Jawa+Barat`
- `GET /api/earthquakes/latest`
- `GET /api/hospitals?kabupaten_id=3273&page=1&per_page=20`

## Runtime configuration

Copy `.env.example` to `.env`, set `API_INDONESIA_KEY`, then run:

```bash
go run .
```

`API_INDONESIA_KEY` is required and must stay server-side. Never expose it in browser code, frontend configuration, logs, or API responses. `API_INDONESIA_BASE_URL` defaults to `https://use.apiindonesia.id`.

Redis is optional. If unavailable, the service continues without Redis-backed features.

## Language boundary

Backend package names, identifiers, logs, response messages, and error codes are English. The frontend owns Indonesian/English translation and presentation. Official upstream names and warning text remain unchanged.

## Container

```bash
docker build -t kota-siaga-backend .
docker run --env-file .env -p 8080:8080 kota-siaga-backend
```
