# Biblios frontend

React 19 and TypeScript single-page application for the Biblios catalogue,
personal and cooperative libraries, reading activity, reviews, notifications,
and moderation workflows.

## API routing

The browser uses the same-origin `/api/v1` base path by default. This avoids
baking a developer machine's `localhost` into a deployed bundle.

* In local Docker development, Vite proxies `/api` to `http://backend:8080`.
* When Vite runs directly on the host, the default proxy target is
  `http://localhost:8081`.
* The production Nginx image proxies `/api` to the Compose backend service.
* `VITE_API_URL` may override the browser-facing base with a root-relative path
  or an absolute HTTPS URL. Plain HTTP, protocol-relative, malformed, and
  credential-bearing values fail validation.

See `.env.example` for the supported frontend variables. Never commit a real
environment file or credentials.

## Local Docker development

From the repository root:

```bash
docker compose up --build
```

The development image installs dependencies with `npm ci`. A named
`frontend_node_modules` volume caches them. On every start, the entrypoint
compares a fingerprint of `package.json` and `package-lock.json` with the
installed volume and automatically reruns `npm ci` when they differ. Ordinary
dependency changes therefore do not require deleting volumes manually.

The application is available at `http://localhost:5173`. Check readiness with:

```bash
docker compose ps
docker compose logs --tail=100 frontend
```

If the dependency volume itself is damaged, replace only the frontend
dependency state; do not casually run `docker compose down -v`, because that
can remove the PostgreSQL data volume.

## Direct host development

```bash
npm ci
npm run dev
```

Vite proxies `/api` to `http://localhost:8081` unless
`VITE_API_PROXY_TARGET` is explicitly set.

## Production frontend image

The Dockerfile's `production` target performs a lockfile-reproducible build and
copies only `dist/` plus Nginx configuration into the runtime image. The
runtime:

* does not contain Node, npm, Vite, source files, or `node_modules`;
* runs Nginx as its unprivileged `nginx` user on port 8080;
* serves SPA routes through `index.html` fallback;
* serves hashed assets with immutable caching;
* proxies `/api` to the backend service;
* exposes `/healthz` and includes a container health check.

Build it directly:

```bash
docker build --target production \
  --build-arg VITE_API_URL=/api/v1 \
  --build-arg APP_VERSION=0.1.0 \
  --build-arg VCS_REF=<git-commit> \
  --build-arg BUILD_DATE=<ISO-8601-date> \
  --tag biblios-frontend:production \
  frontend
```

Exercise the production frontend with the root Compose stack:

```bash
FRONTEND_PUBLIC_PORT=8080 docker compose \
  -f docker-compose.yml \
  -f docker-compose.frontend-production.yml \
  up --build frontend
```

Open `http://localhost:8080` and verify both `/healthz` and an authenticated API
journey. The override removes development bind/dependency volumes, so the
production container serves only immutable image contents.

This frontend override does not certify the remaining services for production.
TLS ingress, secrets, datastore isolation, backend image hardening,
observability, backups, and deployment/rollback validation are tracked
separately.

## Quality checks

```bash
npm run lint
npx tsc --noEmit
npm test
npm run build
```

All checks must run from a clean `npm ci` installation in CI. Production image
validation must additionally inspect the configured user, health check, static
SPA fallback, and `/api` proxy behavior.
