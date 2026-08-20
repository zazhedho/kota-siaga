# Kota Siaga Frontend

Public disaster preparedness and healthcare monitoring dashboard for Indonesia.

## Getting Started

### Prerequisites

- Node.js (v18+)
- Backend service running on `http://localhost:8080`

### Local Development Setup

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

The application will be available at `http://localhost:5173`.

### Environment Configuration

In `.env.local`:
```text
VITE_API_BASE_URL=/api
```

- In development, Vite proxies requests from `/api` to the Go backend (`http://localhost:8080`).
- **Security Notice:** The frontend never connects directly to API Indonesia and never holds `API_INDONESIA_KEY`. All requests are proxied securely through the Go backend.

### Available Scripts

- `npm run dev`: Starts the local Vite development server.
- `npm run build`: Compiles and bundles production assets into `dist/`.
- `npm test`: Runs Vitest unit and integration test suites.
- `npm run test:watch`: Runs tests in watch mode.
- `npm run lint`: Checks codebase against ESLint rules.
- `npm run preview`: Previews the production build locally.
