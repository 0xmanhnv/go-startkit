## Changelog (Unreleased)

### Added
- Feature-flagged refresh token flow with rotation/revocation (Redis-backed)
  - Endpoints: POST /v1/auth/refresh, POST /v1/auth/logout (enabled only when AUTH_REFRESH_ENABLED=true)
  - Config: REFRESH_TTL_SEC, AUTH_REFRESH_ENABLED
- Redis distributed rate limiter middleware for /v1/auth/login (activates when REDIS_ADDR is set)
- API Docs (dev-only): GET /openapi.json (OpenAPI 3.0), GET /swagger (ReDoc viewer)
- HTTP error mapper `response.FromError` for consistent error→HTTP mapping
- Application sentinel errors (apperr): ErrInvalidCredentials, ErrInvalidRefreshToken, ErrRefreshStoreNotConfigured
- DB pool tuning via config: DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS, DB_CONN_MAX_LIFETIME_SEC, DB_CONN_MAX_IDLE_TIME_SEC
- DevEx: Makefile (build/run/test/lint/tools/dev-up/prod-up), .golangci.yml

### Changed
- Logger: global singleton (logger.Init, logger.L()); removed per-file/new logger creation
- Config: singleton load (sync.Once); inject cfg into constructors; avoid Load() in leaf code
- Router: consolidated to NewRouter(userHandler, cfg, authMiddleware...) only
- JWT: set NotBefore=now; validate iss/aud/nbf with leeway; HS256 allowed methods
- Bcrypt: cost injected via cfg.Security.BcryptCost
- Auth middleware: JWTAuth accepts TokenValidator func for IoC
- Handlers: standardized JSON binding error messages; unified error mapping via response helpers and mapper
- Security headers: CSP added; HSTS only on HTTPS/X-Forwarded-Proto=https
- CORS: applied only via config (removed default duplicate CORS)
- DB connection: defaults for pool; overrides applied from config in bootstrap
- Refresh token TTL now read from `REFRESH_TTL_SEC` via `RefreshUseCase` (removes hard-coded 7d)
- Rate limit 429 responses now include `Retry-After`, `X-RateLimit-*` headers (both in-memory and Redis limiter)
- Recovery: replace gin.Recovery with JSONRecovery to standardize 500 error envelope

### Fixed
- Double route registration panic (removed NewRouterWithConfig recursion)
- Clearer JSON binding errors (empty body, malformed JSON, invalid types)
- Invalid jwt.WithNotBefore usage (manual NotBefore check)

### Removed
- docs/step-by-step.md and docs/starter-kit-checklist.md (duplicated guidance; consolidated into README and docs/review.md)

### Security
- Rate-limit login (in-memory or Redis for multi-instance)
- Do not expose /swagger or /openapi.json in production (mounted only when ENV=dev)
- Secrets guidance: manage JWT_SECRET/DB_PASSWORD via secret manager in prod

### Docs
- README: architecture, ports glossary, naming, feature flags (AUTH_REFRESH_ENABLED), DB pool tuning, API Docs (dev-only), DevEx shortcuts
- docs/review.md: reduced to Pending (Observability, Testing & CI, Secrets)

### DevEx
- Makefile targets for build/run/test/lint/dev/prod
- .golangci.yml aligned to schema (with version), conservative linter set

### Compose/Docker
- docker-compose.dev.yml: app+db+redis; Air hot reload; sensible defaults
- docker-compose.yml: app+db+redis, healthchecks; environment injection; reminders for secrets
- build/Dockerfile: multi-stage, distroless runtime

### Observability (Planned)
- Next: /metrics (Prometheus), pprof (dev-only), optional OpenTelemetry for HTTP/DB

### Testing & CI (Planned)
- Next: GH Actions (build/test/lint/govulncheck), Trivy scan, unit/HTTP/integration tests, coverage thresholds


