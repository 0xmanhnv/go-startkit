# AppSecHub

AppSecHub is a Go starter kit for building HTTP services, following a clear layered architecture (Domain → UseCase → Interface/HTTP → Infrastructure) with foundational security practices (password hashing, JWT, migrations, environment-driven config).

## Request flow (example: login)

```txt
HTTP Request (Gin)
    ↓
Handler (UserHandler)
    ↓
DTO → UseCase
    ↓
UseCase (UserUsecase)
    ↓
Call Repository Interface (UserRepository)
    ↓
Repository implementation (Postgres)
    ↓
Domain Entity/ValueObject handles business logic (validation, password hashing...)
    ↓
Return result (token / error)
    ↓
Handler returns HTTP Response
```

## System requirements
- Go 1.22+
- Docker 24+ and Docker Compose v2

## Environment configuration
- Copy `.env.example` to `.env` (for local), or inject variables via CI/CD for containers.
- Important variables: `ENV`, `HTTP_PORT`, `DB_*`, `MIGRATIONS_PATH`, `JWT_SECRET`, `JWT_EXPIRE_SEC`.
- Optional RBAC policy from YAML: set `RBAC_POLICY_PATH` to a YAML file path (see `configs/rbac.policy.yaml`).
- Optional seeding (init admin user):
  - `SEED_ENABLE=true`
  - `SEED_USER_EMAIL=admin@example.com`
  - `SEED_USER_PASSWORD=ChangeMe!123`
  - `SEED_USER_FIRST_NAME=Admin` (optional)
  - `SEED_USER_LAST_NAME=User` (optional)
  - `SEED_USER_ROLE=admin` (optional)
 - Optional JWT hardening:
   - `JWT_ISSUER=appsechub`
   - `JWT_AUDIENCE=appsechub-clients`
   - `JWT_LEEWAY_SEC=30`
 - Optional HTTP security & rate limit:
   - `HTTP_SECURITY_HEADERS=true` (enable common security headers; use behind TLS)
   - `HTTP_LOGIN_RATELIMIT_RPS=1`
   - `HTTP_LOGIN_RATELIMIT_BURST=5`

## Development (hot reload)
1) Docker + Air (recommended):
   - `docker compose -f docker-compose.dev.yml up`
   - Source code is bind-mounted; the app rebuilds automatically on changes.
2) Local (Go and Air installed):
   - Install Air: `go install github.com/cosmtrek/air@latest`
   - Run: `air -c ./.air.toml`

Default API base URL: `http://localhost:8080`

## Production (reference)
- Build & run: `docker compose up -d --build`
- Set environment variables securely on host/CI (especially `JWT_SECRET`, `DB_PASSWORD`).
- Do not expose the DB port publicly; keep it on an internal network.

## Database migrations
- Migrations reside in `migrations/` and run automatically on startup.
- Optionally manage them with the `golang-migrate` CLI.

## Sample endpoints
- `POST /v1/auth/register` – create a new user
  - JSON body: `first_name`, `last_name`, `email`, `password`, `role`.
- `POST /v1/auth/login` – accepts `email/password`, returns JWT.
- Admin example (requires JWT and RBAC permission): `GET /v1/admin/stats`
  - Send header: `Authorization: Bearer <JWT>`
  - Login may be rate limited (HTTP 429) based on `HTTP_LOGIN_RATELIMIT_*`.

## Project layout
- `internal/domain` – Entities, ValueObjects, business rules
- `internal/application` – DTOs, UseCases, Services
- `internal/interfaces/http` – Router, Handlers, Middleware (Gin)
- `internal/infras` – DB migrator, security (JWT/bcrypt), storage (Postgres)
- `migrations` – SQL migrations
- `build` – Dockerfile and build assets

## Import rules (Clean Architecture / DDD)

| Layer | Directory | Allowed to import | Must NOT import | Notes/Examples |
|------|-----------|-------------------|------------------|----------------|
| Domain (Core) | `internal/domain/...` | Standard library only (and very small third-party if truly pure) | `application`, `interfaces`, `infras`, `config` | Contains Entities, ValueObjects, Domain Errors, Repository Interfaces. |
| Application (UseCases) | `internal/application/...` | `internal/domain/...`, application-local interfaces (e.g. `application/service`) | `interfaces/http/...`, `infras/...` | Orchestrates domain logic; depends inward on domain; define `TokenService` etc. as interfaces. |
| Interfaces (HTTP) | `internal/interfaces/http/...` | `internal/application/...`, interface-local helpers (e.g. `response`), third-party frameworks (Gin) | `infras/...` | Accept dependencies via wiring; do not call infra directly. Map domain/application errors → HTTP. |
| Infrastructure | `internal/infras/...` | `internal/domain/...`, application service interfaces (e.g. implement `TokenService`) | `interfaces/http/...`, application DTOs/usecases | Implements repositories (Postgres), security (JWT/Bcrypt), migrator. No knowledge of HTTP. |
| Config | `internal/config/...` | Standard library + config libs | n/a | Read env/config; imported by `main` and infra where needed. |
| Composition Root | `cmd/api` | All layers | n/a | Wire everything: build DSN, create repos/services/usecases/handlers, create router. |

Dependency flow must be inward only: `interfaces → application → domain`; `infras → domain` (and application service interfaces). `cmd/api` is the only place that sees and wires across all layers.

## Documentation
- Codebase review & pending fixes: `docs/review.md`
- Best-practice starter checklist: `docs/starter-kit-checklist.md`
  - Step-by-step roadmap: `docs/step-by-step.md`
  - RBAC policy file example: `configs/rbac.policy.yaml` (configure via `RBAC_POLICY_PATH`)
  - Authorization usage: include header `Authorization: Bearer <JWT>` for protected routes

## Security recommendations
- Password hashing with bcrypt (integrated); increase cost appropriately for production.
- Manage `JWT_SECRET` with a secret manager; use short TTL; consider refresh tokens/rotation.
- Enable CORS appropriately, add security headers (HSTS, X-Content-Type-Options, Referrer-Policy...).
- Structured logging with a `request-id`; never log secrets or sensitive data.

