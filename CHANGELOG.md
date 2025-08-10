## Changelog

## Unreleased
- Observability: add `/metrics` (Prometheus), enable pprof in dev, optional OpenTelemetry tracing for HTTP and DB.
- CI/CD: add GitHub Actions (build, lint, unit + integration tests, govulncheck), optional Trivy image scan.

## 1.0.0 - 2025-08-10

### Added
- Layered architecture with composition root wiring (`cmd/api`).
- Health and readiness endpoints: `GET /healthz`, `GET /readyz` (DB ping with timeout).
- Auth: JWT with role in claims, auth middleware, RBAC with YAML policy (`RBAC_POLICY_PATH`).
- Auth endpoints: `POST /v1/auth/register`, `POST /v1/auth/login`, `GET /v1/auth/me`, `POST /v1/auth/change-password`.
- Feature-flagged refresh flow (Redis-backed): `POST /v1/auth/refresh`, `POST /v1/auth/logout` when `AUTH_REFRESH_ENABLED=true`.
- Rate limit for `/v1/auth/login` (in-memory); optional Redis distributed limiter when `REDIS_ADDR` set; 429 includes `Retry-After`, `X-RateLimit-*`.
- Security headers middleware (CSP, HSTS when HTTPS), CORS from env, trusted proxies config.
- Response envelope and error mapping helpers; friendly JSON binding/validation errors; request body size limit via `HTTP_MAX_BODY_BYTES`.
- Postgres migrations and repository; connection pool tuning via env.
- Seeding initial admin via `SEED_*` env.
- OpenAPI 3.0 schema (`/openapi.json`) and ReDoc (`/swagger`) in dev-only.
- DevEx: `Makefile` targets, `.golangci.yml`, `scripts/test_integration.sh`.
- Docker: multi-stage `build/Dockerfile` (distroless runtime), `docker-compose.dev.yml`, `docker-compose.yml`, `docker-compose.test.yml`.

### Changed
- Global structured logger initialization (`logger.Init`), use `logger.L()` across the app.
- JWT hardening: set `nbf=now`, validate `iss/aud/nbf/exp` with leeway; restrict to HS256.
- Router consolidated to a single `NewRouter(...)`; graceful shutdown and HTTP server timeouts.

### Fixed
- Prevent double route registration; clearer JSON error messages; corrected `NotBefore` validation.

### Security
- Do not expose `/swagger` or `/openapi.json` in production (only when `ENV=dev`).
- Notes on secrets management for `JWT_SECRET` and `DB_PASSWORD`.

### Docs
- README updated (architecture, ports glossary, feature flags, DB pool tuning, DevEx shortcuts).
- OpenAPI `info.version` bumped to `1.0.0`.


