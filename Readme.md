# AppSecHub

AppSecHub is a Go starter kit for building HTTP services, following a clear layered architecture (Domain → UseCase → Interface/HTTP → Infrastructure) with foundational security practices (password hashing, JWT, migrations, environment-driven config).

## Architecture diagrams
- Project layout (directories): see section "Project layout"
- Layered architecture (imports): see subsection under "Project layout"
- Runtime call flow (protected route): see section "Runtime call flow (authn protected route)"
- Detailed request flow (login): see section "Detailed request flow (login, with rate limit & validation)"
- Protected route flow: see section "Protected route flow (/v1/auth/me with JWTAuth)"

## Request flow (example: login)

```txt
HTTP Request (Gin)
    ↓
Base middlewares: JSONRecovery → RequestID → (Logger+SecurityHeaders if enabled) → CORS
    ↓
Route match (Gin) + optional RateLimit for /v1/auth/login
    ↓
Handler (UserHandler)
    ↓
DTO → UseCase
    ↓
UseCase (UserUsecase)
    ↓
Call Repository Interface (UserRepository) + Ports (e.g., TokenIssuer)
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
- Go 1.24+
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
    - For multi-instance/prod, use distributed limiter (e.g., Redis) instead of in-memory.
    - `HTTP_MAX_BODY_BYTES=1048576` (limit JSON body size; default 1 MiB). Requests exceeding this return 413 with code `payload_too_large`.
  - Optional password hashing:
    - `BCRYPT_COST=12` (4–31). Higher = slower = stronger. Tune per env (dev lower for speed, prod higher ~100–250ms/hash target).
  - Optional DB pool tuning:
    - `DB_MAX_OPEN_CONNS=25`, `DB_MAX_IDLE_CONNS=25`
    - `DB_CONN_MAX_LIFETIME_SEC=900`, `DB_CONN_MAX_IDLE_TIME_SEC=300`
  - Optional refresh tokens (feature flag):
    - `AUTH_REFRESH_ENABLED=false` (enable to expose `/v1/auth/refresh` and `/v1/auth/logout`)
    - `REFRESH_TTL_SEC=604800` (7d default; only used when refresh is enabled; controls rotation TTL)

## Development (hot reload)
1) Docker + Air (recommended):
   - `docker compose -f docker-compose.dev.yml up`
   - Source code is bind-mounted; the app rebuilds automatically on changes.
2) Local (Go and Air installed):
   - Install Air: `go install github.com/cosmtrek/air@latest`
   - Run: `air -c ./.air.toml`

### DevEx shortcuts
- Makefile targets:
  - `make build|run|test|fmt|vet|lint`
  - `make dev-up|dev-down` (compose dev), `make prod-up|prod-down`
  - `make tools` (install Air, golangci-lint)
- Linting: configure via `.golangci.yml` (optional in CI).

Default API base URL: `http://localhost:8080`

### API Docs (dev-only)
- Available only when `ENV=dev`:
  - `GET /swagger` (ReDoc viewer)
  - `GET /openapi.json` (OpenAPI 3.0)
- Do not expose in production.

### Distributed rate limit (production)
- Current limiter is in-memory per instance (OK for dev/single instance).
- For prod with >1 replicas, use Redis-based limiter (configure `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`). Compose includes `redis` service in dev/prod profiles.

### Refresh tokens (JWT hardening)
When `AUTH_REFRESH_ENABLED=true` and Redis configured, the following apply:
- Endpoints exposed: `POST /v1/auth/refresh`, `POST /v1/auth/logout`.
- Refresh TTL controlled by `REFRESH_TTL_SEC` (default 604800).

### Validation middleware (pre-handler)
- Requests are bound and validated via middleware before reaching handlers.
- Errors are mapped to friendly messages; body size is enforced by `HTTP_MAX_BODY_BYTES`.
- Handlers read validated DTOs from context key `"req"`.

### Registration policy
- Public registration (`POST /v1/auth/register`) only permits non-admin roles. Attempts to register as `admin` are rejected with `invalid_request`.
- Access token short TTL; issue refresh tokens with rotation and revocation list (e.g., stored in Redis with TTL).
- Skeleton implemented: application port `RefreshTokenStore`, Redis-backed implementation `internal/infras/auth/redis_refresh_store.go`. Wire and endpoints for refresh/revoke can be added as needed.

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
- `POST /v1/auth/refresh` – exchange refresh token for new access token (only when `AUTH_REFRESH_ENABLED=true`)
- `POST /v1/auth/logout` – revoke refresh token (only when `AUTH_REFRESH_ENABLED=true`)
- Admin example (requires JWT and RBAC permission): `GET /v1/admin/stats`
  - Send header: `Authorization: Bearer <JWT>`
  - Login may be rate limited (HTTP 429) based on `HTTP_LOGIN_RATELIMIT_*`.
    - 429 responses include headers: `Retry-After`, `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`.
    - In distributed mode (Redis), reset aligns to 1s windows.

### Error envelope & Recovery
- All unhandled panics are converted to JSON envelope:
  `{ "error": { "code": "server_error", "message": "internal error" }, "meta": { "request_id": "..." } }`.
- Every response includes `X-Request-Id` for correlation.

## Project layout
```
AppSecHub/
├─ cmd/
│  └─ api/                  # Composition root (main, bootstrap, wiring)
├─ internal/
│  ├─ application/          # Use cases, DTOs, application ports (interfaces)
│  │  ├─ dto/
│  │  └─ usecase/
│  │     └─ userusecase/
│  ├─ domain/               # Entities, value objects, domain services & errors
│  │  └─ user/
│  ├─ interfaces/
│  │  └─ http/              # HTTP adapters: router, handlers, middleware, apidocs
│  │     ├─ apidocs/
│  │     ├─ handler/
│  │     └─ middleware/
│  ├─ infras/               # Infrastructure adapters (implement application ports)
│  │  ├─ auth/              # Refresh token store (Redis)
│  │  ├─ db/                # DSN, migrations, health checks
│  │  ├─ ratelimit/         # Distributed rate limiter (Redis)
│  │  ├─ security/          # JWT service, bcrypt hasher
│  │  └─ storage/
│  │     └─ postgres/       # User repository (Postgres)
│  └─ config/               # Strongly-typed env config & loader
├─ pkg/                     # Shared utilities (logger, rbac, validator)
├─ migrations/              # SQL migrations
├─ build/                   # Dockerfile & build assets
├─ configs/                 # Example config files (RBAC, etc.)
├─ docs/                    # Project docs, reviews
├─ scripts/                 # Helper scripts
└─ docker-compose*.yml      # Dev/Prod composition
```

```
Layered architecture (imports flow)

interfaces/http  →  application (usecases, ports)  →  domain (entities, VOs)
      ▲                         │
      │                         └── ports implemented by
      │                              infrastructure (jwt, db, redis, ...)

cmd/api wires everything together (composition root)
```
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
| Application (UseCases) | `internal/application/...` | `internal/domain/...`, application-local interfaces (e.g. `application/ports`) | `interfaces/http/...`, `infras/...` | Orchestrates domain logic; depends inward on domain; define ports (e.g., `TokenIssuer`) as interfaces. |
| Interfaces (HTTP) | `internal/interfaces/http/...` | `internal/application/...`, interface-local helpers (e.g. `response`), third-party frameworks (Gin) | `infras/...` | Accept dependencies via wiring; do not call infra directly. Map domain/application errors → HTTP. |
| Infrastructure | `internal/infras/...` | `internal/domain/...`, application ports (e.g. implement `TokenIssuer`) | `interfaces/http/...`, application DTOs/usecases | Implements repositories (Postgres), security (JWT/Bcrypt), migrator. No knowledge of HTTP. |
| Config | `internal/config/...` | Standard library + config libs | n/a | Read env/config; imported by `main` and infra where needed. |
| Composition Root | `cmd/api` | All layers | n/a | Wire everything: build DSN, create repos/services/usecases/handlers, create router. |

Dependency flow must be inward only: `interfaces → application → domain`; `infras → domain` (and application service interfaces). `cmd/api` is the only place that sees and wires across all layers.

### Runtime call flow (authn protected route)
```txt
Client
  → Gin router
    → Middlewares: Recovery → RequestID → (Logger/SecurityHeaders) → CORS → (JWTAuth on protected group)
      → Handler (UserHandler)
        → UseCase (UserUsecases)
          → Ports (TokenIssuer, EmailSender, SMSSender, ObjectStorage...) & Repositories
            → Infrastructure implementations (JWT, SMTP, Twilio, Postgres...)
```

### Detailed request flow (login, with rate limit & validation)
```txt
Client
  │
  ├─► Gin Router
  │     │
  │     ├─► JSONRecovery (panic → 500 JSON envelope)
  │     ├─► RequestID (add X-Request-Id)
  │     ├─► Logger (structured logs)
  │     ├─► SecurityHeaders (optional; CSP/HSTS/XFO/...)
  │     └─► CORS (per config)
  │
  ├─► Route match: POST /v1/auth/login
  │     ├─► (Optional) RateLimit (in-memory or Redis, only for /v1/auth/login)
  │     │     └─► If limited → 429 + headers: Retry-After, X-RateLimit-*
  │     └─► ValidateJSON[LoginRequest] (MaxBodyBytes, binding/validation errors → 400)
  │
  └─► Handler: UserHandler.Login
        └─► UserUsecases.Login
              ├─► UserRepository.GetByEmail (Postgres)
              ├─► PasswordHasher.Compare (bcrypt)
              ├─► TokenIssuer.GenerateToken (JWT)
              └─► (Optional) RefreshTokenStore.Issue (Redis, if AUTH_REFRESH_ENABLED)
                  
            ←─ Response DTO (access_token, refresh_token?, user)
        ←─ HTTP 200 { data: ... } (envelope)

Errors
  - Domain/App errors → mapped via response.FromError → 400/401/404/409/500 with { error: { code, message } }
  - Panic → JSONRecovery → 500 with { error: { code: "server_error" }, meta: { request_id } }
```

### Protected route flow (/v1/auth/me with JWTAuth)
```txt
Client
  │
  ├─► Gin Router
  │     │
  │     ├─► JSONRecovery → RequestID → Logger → SecurityHeaders? → CORS
  │     └─► Route match: GET /v1/auth/me (protected group)
  │            └─► JWTAuth middleware
  │                  ├─ Check Authorization header present
  │                  ├─ Extract Bearer token
  │                  ├─ Validate via TokenValidator → JWTService.ValidateToken
  │                  │     ├─ Verify signature (HS256), iat, exp, nbf(+leeway)
  │                  │     ├─ Optional issuer/audience checks
  │                  │     └─ Return claims { sub=user_id, role }
  │                  ├─ Set context: user_id, user_role
  │                  └─ On error → 401 with WWW-Authenticate and envelope { error: { code: "unauthorized"|"invalid_token" } }
  │
  └─► Handler: UserHandler.GetMe
        └─► UserUsecases.GetMe(user_id)
              └─► UserRepository.GetByID (Postgres)
                   └─ Not found → domain ErrUserNotFound

Responses
  - 200: { data: UserResponse }
  - 401: { error: { code: "unauthorized", message: "missing or invalid token" } }
  - 404: { error: { code: "not_found", message: "user not found" } }
```

### Composition root wiring (cmd/api)
```txt
config.Load
  → initPostgresAndMigrate (Build DSN → Run migrations → Open *sql.DB)
  → initJWTService (infra) then build validator func for middleware
  → (optional) seedInitialUser
  → loadRBACPolicy (YAML)
  → buildUserComponents (repo, hasher, usecases, handlers)
  → NewRouter(userHandler, cfg, JWTAuth(validator)) + AddReadiness
  → http.Server + graceful shutdown (SIGINT/SIGTERM)
```

### Import dependency flow (allowed edges)
```txt
interfaces/http  → application (usecases, ports)
application      → domain (entities, VOs) & application/ports
infrastructure   → domain (and implement application/ports)
cmd/api          → all (composition root only)
```

### Naming & conventions (DDD / Clean Architecture)
- Phân định vai trò theo lớp:
  - Domain: Entities, Value Objects, Domain Services (chỉ business logic thuần), Domain Errors.
  - Application: Use Cases (điều phối), Ports (interfaces) cho phụ thuộc ra ngoài (Token, Email, SMS, ObjectStorage...). Không import hạ tầng.
  - Infrastructure: Adapters/Implementations cho các Ports (JWT, SMTP, Twilio, S3, Postgres...). Không biết HTTP/use case.
  - Composition Root (`cmd/api`): Wire mọi thứ (khởi tạo infra, inject vào ports/use cases, tạo router).

- Về đặt tên (naming):
  - Thư mục `internal/application/ports/` chứa các PORTS (interfaces/func adapters) mà Application cần.
  - Domain Service (nếu có) phải đặt ở `internal/domain/...` và không phụ thuộc application/infra.
  - Middleware/HTTP helpers ở `internal/interfaces/http/...` chỉ map/biến đổi, không chứa business logic.

- Ví dụ ports trong Application:
  - `TokenIssuer` (phát token) → implemented bởi `internal/infras/security`.
  - `EmailSender`, `SMSSender` → implemented bởi `internal/infras/notify/...`.
  - `ObjectStorage` → implemented bởi `internal/infras/storage/...`.
  - Inject các ports tại `cmd/api/bootstrap.go` để giữ Inversion of Control chặt chẽ.

#### Ports glossary (giải nghĩa tên gọi đề xuất)
- Issuer: thành phần “phát hành” (issue) một loại thông tin, ví dụ token (JWT). Ví dụ: `TokenIssuer.Generate(userID, role)`.
- Validator/Verifier: thành phần xác thực/kiểm chứng một giá trị (ví dụ: `Validate(token)`), thường dùng ở middleware.
- Sender: thành phần “gửi” một thông điệp (Email/SMS), ví dụ `EmailSender.Send(...)`, `SMSSender.Send(...)`.
- Storage: thành phần lưu trữ đối tượng/blob (ví dụ `ObjectStorage.Put/Get/Delete`).
- Provider/Gateway: adapter kết nối tới hệ thống bên ngoài (payment, oauth...), có thể dùng nếu phù hợp ngữ cảnh.

## Documentation
- Codebase review & pending fixes: `docs/review.md`
- Best-practice starter checklist: `docs/starter-kit-checklist.md`
  - Step-by-step roadmap: `docs/step-by-step.md`
  - RBAC policy file example: `configs/rbac.policy.yaml` (configure via `RBAC_POLICY_PATH`)
  - Authorization usage: include header `Authorization: Bearer <JWT>` for protected routes

## Rename project/module
If you fork this template and want to change the module/project name (Go module path), follow these steps.

1) Choose the new module path (example: `github.com/you/securehub`). Then update `go.mod`:
```bash
NEW=github.com/you/securehub
go mod edit -module "$NEW"
```

2) Update all internal imports from `appsechub/...` to the new module path:
```bash
OLD=appsechub
NEW=github.com/you/securehub
rg -l --type go "\"$OLD/" | xargs -r sed -i "s#\"$OLD/#\"$NEW/#g"
go mod tidy
go build ./...
```

3) Optional but recommended – align names across configs/docs:
- Docker image/name and references in `docker-compose*.yml` (e.g., `appsechub` → your new name)
- OpenAPI title in `internal/interfaces/http/apidocs/openapi.json` ("AppSecHub API" → your new title)
- README/CHANGELOG headings and references
- Default JWT metadata in `internal/config/config.go` (issuer/audience: `appsechub` → your values)
- Dev DB name/user in compose files (if you want to keep everything consistent)

4) If publishing the repo, ensure the module path matches the canonical VCS path (e.g., GitHub URL) to avoid `go get` issues.

## Security recommendations
- Password hashing with bcrypt (integrated); increase cost appropriately for production.
- Manage `JWT_SECRET` with a secret manager; use short TTL; consider refresh tokens/rotation.
- Enable CORS appropriately, add security headers (HSTS when HTTPS, X-Content-Type-Options, Referrer-Policy, X-Frame-Options).
- Content Security Policy (CSP): default applied by middleware: `default-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'`. Adjust per frontend needs.
- Structured logging with a `request-id`; never log secrets or sensitive data.

### Logging guidance
- Initialize once at entrypoint with `logger.Init(...)` and use `logger.L()` everywhere else.
- Do not create new loggers inside router/middlewares; it may desync level/format.

