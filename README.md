# Kota Siaga Backend

Public Go backend foundation for Kota Siaga. It keeps Gin, configuration, response envelopes, logging, CORS, recovery, request IDs, optional Redis, rate limiting, and shared utilities. PostgreSQL, migrations, authentication, RBAC, sessions, media, mail, audit, and copied local-location persistence are not part of this service.

## Public routes

The application currently exposes these public GET routes:

- `GET /healthcheck`
- `GET /api/locations/province?page=1&per_page=20`
- `GET /api/locations/city?province_id=32&page=1&per_page=20`
- `GET /api/locations/district?kabupaten_id=3273&page=1&per_page=20`
- `GET /api/locations/village?kecamatan_id=327301&page=1&per_page=20`
- `GET /api/weather?adm4=32.73.01.1001`
- `GET /api/warnings?provinsi=Jawa+Barat`
- `GET /api/earthquakes/latest`
- `GET /api/hospitals?kabupaten_id=3273&page=1&per_page=20&search=hasan`
  The optional `search` query searches hospital names within the selected city/regency; it does not provide distance or nearest-hospital ranking.

## Runtime configuration

Copy `.env.example` to `.env`, set `API_INDONESIA_KEY`, `SATUSEHAT_CLIENT_ID`, and `SATUSEHAT_CLIENT_SECRET`, then run:

```bash
go run .
```

`API_INDONESIA_KEY`, `SATUSEHAT_CLIENT_ID`, and `SATUSEHAT_CLIENT_SECRET` are required server-side credentials. Never expose them in browser code, frontend configuration, logs, or API responses. Weather and warning requests use API Indonesia. Hospital requests use SATUSEHAT Master Sarana Index through the production defaults `SATUSEHAT_BASE_URL=https://api-satusehat.kemkes.go.id/masterdata` and `SATUSEHAT_AUTH_BASE_URL=https://api-satusehat.kemkes.go.id/oauth2/v1`. Earthquake requests use BMKG open data through `BMKG_BASE_URL`, which defaults to `https://data.bmkg.go.id` and does not require an API key. Location hierarchy requests use `LOCATION_SERVICE_BASE_URL`, which defaults to `https://indonesia.imyourz.com` and does not require an API key.

Redis is optional. If unavailable, the service continues without Redis-backed features.

## Language boundary

Backend package names, identifiers, logs, response messages, and error codes are English. The frontend owns Indonesian/English translation and presentation. Official upstream names and warning text remain unchanged.

## Container

```bash
docker build -t kota-siaga-backend .
docker run --env-file .env -p 8080:8080 kota-siaga-backend
```
