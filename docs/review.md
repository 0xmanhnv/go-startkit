## AppSecHub — Pending (High Priority)

- Observability
  - Expose `/metrics` (Prometheus)
  - Add pprof (dev-only)
  - Consider OpenTelemetry tracing for HTTP and DB

- Testing & CI
  - GitHub Actions: build, test, lint, govulncheck
  - Container image scanning (e.g., Trivy)
  - Tests:
    - Unit: domain/usecases (mock repo/hasher/jwt)
    - HTTP handlers (table-driven tests)
    - Integration: repositories with Postgres (testcontainers/compose)
  - Set reasonable coverage thresholds

- Secrets Management (Production)
  - Manage `JWT_SECRET`, `DB_PASSWORD` via a secret manager (AWS/GCP/Vault)
  - Rotate secrets periodically; inject at runtime (ENV/volumes)
  - Keep `/swagger` disabled in prod (already enforced by ENV)

## AppSecHub — Codebase Review (Initial)

### Current status (updated)
- Completed:
  - Manual wiring (DB → repo → hasher/JWT → usecases → handler → router)
  - Split `main` into dedicated bootstrap helpers: `initPostgresAndMigrate`, `initJWTService`, `buildUserComponents`, `loadRBACPolicy`, `buildRouter`
  - Usecases: CreateUser, Login + handler/router
  - JWT claims include role; JWTAuth middleware, RequireRoles/RequirePermissions (RBAC)
  - In-memory RBAC policy + YAML load (`RBAC_POLICY_PATH`) and warnings for unknown roles
  - Response envelope + basic domain error mapping; repo timeouts + map `sql.ErrNoRows`/`23505`
  - CORS/Trusted proxies from env; RequestID UUIDv4 + structured logging (slog)
  - Health: `/healthz`; Ready: `/readyz` DB ping (timeout)
  - Docker compose dev/prod; distroless runtime; hot-reload with Air
  - Router normalization: helpers `applyBaseMiddlewares`, `registerHealthRoutes`, `registerAPIV1Routes`, `registerAuthRoutes` (rate limit via config)

#### Recent updates (implemented)
- CORS: removed default CORS in `applyBaseMiddlewares`; applied only via `applyCORSFromConfig` (`HTTP_CORS_ALLOWED_ORIGINS`).
- Grouped auth routes in `registerAuthRoutes` (includes `/register`, `/login` and protected `/me`, `/change-password`); removed `registerAuthLogin`.
- Friendly JSON binding errors: empty body → "request body is empty"; malformed JSON → "malformed JSON at position N"; wrong type → "invalid type for field <field>".
- JWT: set `NotBefore=now` on token issue; during validation, if `now + leeway < nbf` return `token not yet valid` (reduces clock skew effects).
- Router API: consolidated to a single `NewRouter(userHandler, cfg, authMiddleware...)`.
- Graceful shutdown: `http.Server` (`ReadHeaderTimeout=5s`, `IdleTimeout=60s`) + `Shutdown(15s)` on `SIGINT/SIGTERM`; close `sqlDB` after stop.
- Logger: global default logger (`logger.Init` once at entrypoint, use `logger.L()` everywhere; `SetLevel` to adjust at runtime). Avoid per-file new loggers.
- HTTP error mapping: standardized for `Register`, `Login`, `GetMe`, `ChangePassword` (400/401/404/409/500) via `response` helpers.
- IoC for Auth: `JWTAuth` accepts `TokenValidator` instead of depending on infra directly; the validator is injected at the composition root.
- Application terminology: switched to "ports" (`TokenIssuer`, `EmailSender`, `SMSSender`, `ObjectStorage`).
- Bcrypt cost from env: `BCRYPT_COST` (4–31); default to library cost if unset/invalid.
- Config singleton: `config.Load()` parses once (sync.Once); inject `*config.Config` from composition root; avoid calling `Load()` in leaf code. Updated `buildUserComponents` to pass `cfg.Security.BcryptCost` into `NewBcryptHasher`.
- Security headers: default CSP `default-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'`; HSTS only on HTTPS (TLS or `X-Forwarded-Proto=https`).

- Remaining/Next recommendations:
  - Load RBAC from YAML in `main` when `RBAC_POLICY_PATH` is present (done)
  - Protect `/v1/admin` by chaining `JWTAuth(jwtSvc)` before `RequirePermissions(...)` (done)
  - Remove manual `X-Request-Id` setting; rely on `middleware.RequestID()` (done)
  - JWT hardening: iss/aud/nbf + leeway (done; refresh tokens feature available)
  - CI/CD: build/test/lint/govulncheck + Docker image build & image scan (GitHub Actions)
  - Quality: `golangci-lint` + Makefile targets (build/run/test/lint/migrate)
  - Testing: unit (domain/usecase), HTTP handlers, integration repo with Postgres (testcontainers/compose)
  - Observability: Prometheus metrics (`/metrics`), OpenTelemetry tracing (HTTP/DB), pprof (dev-only)
  - Security enhancements: refresh token rotation; increase bcrypt cost by env or consider argon2id
  - API Docs: OpenAPI/Swagger, `/swagger` (dev-only)
  - DevEx: convenient migration scripts (make migrate-up/down), seeding script/command
  - Added: seeding initial admin via `SEED_*` env if absent

### Post-migration notes (removed Wire)
- Wire removed completely; manual wiring in `cmd/api/main.go`.
- Review and remove any remaining Wire references in docs to avoid confusion.

### Current architecture (overview)
- Layers: `internal/domain` (Entities/Value Objects), `internal/application` (UseCase/DTO/Ports), `internal/interfaces/http` (Gin router/handler/middleware), `internal/infras` (DB migrate, Security/JWT), `internal/config` (env loader).
- HTTP: Gin router at `internal/interfaces/http/route.go`, user handler at `internal/interfaces/http/handler/user_handler.go`.
- Domain: `internal/domain/user` with `User`, `Email`, `Role`, and domain errors.
- Security: JWT service `internal/infras/security/jwt_service.go` and auth middleware `internal/interfaces/http/middleware/auth.go`.
- DB: Migrator using `golang-migrate` at `internal/infras/db/migrator.go`.
- Config: `internal/config/config.go` + `internal/config/loader.go` (via `caarlos0/env`).

### Build status
- Build: OK on Go 1.24.

### Key issues to address
1) (already addressed in current code) — removed from list.
2) (already addressed in current code) — removed from list.

3) UseCase/DTO mismatches (historical notes)
   - `internal/application/usecase/userusecase/create_user.go`:
     - Missing `PasswordHasher` definition.
     - Used `dto.CreateUserRequest` which did not exist.
     - Called `user.NewUser(input.Email, hashed)` with wrong signature (domain requires `firstName, lastName, email, password, role`).
   - `internal/application/dto/user_dto.go` lacked `CreateUserRequest` at the time; contained `LoginRequest`, `LoginResponse`, `UserResponse`.

4) Handler depended on non-existent types/usecases (historical notes)
   - `internal/interfaces/http/handler/user_handler.go`:
     - Field `Usecase userusecase.UserUsecase` did not exist (no interface/struct defined).
     - Called `h.Usecase.Login.Execute(...)` but `Login` usecase was missing.

5) Incorrect auth service import (historical)
   - `internal/application/service/auth_service.go` imported `appsechub/internal/domain` (should be `internal/domain/user`).

6) Migrations & path (historical)
   - `migrations/` empty would cause `m.Up()` to fail; ensure valid files and correct path.
   - `main` passed `cfg.DB.MigrationsPath` (missing); added `MigrationsPath` to `Config` with default `migrations`.

7) (addressed) — unified router constructor and correct wiring of `*handler.UserHandler`.

8) Dockerfile: use stable `golang:1.24-alpine`; consider adding HTTP healthcheck later.

### Suggested fix roadmap (prioritized)
1) Config and entrypoint
   - Add `MigrationsPath` to `Config` (default `./migrations`).
   - Provide DSN builder for Postgres; ensure `cfg.HTTP.Port` usage in `main`.

2) Storage/Postgres repository
   - Add `internal/infras/storage/postgres/` with `NewPostgresConnection` and `user_repository.go` implementing `internal/domain/user.Repository`.

3) Migrations
   - Add `migrations/0001_init.up.sql` and `.down.sql` for `users` table (unique email, hashed password, role, timestamps).

4) UseCase/DTO
   - Add `dto.CreateUserRequest { FirstName, LastName, Email, Password, Role }`.
   - Define `PasswordHasher` and implement (bcrypt/argon2id).
   - `CreateUserUseCase.Execute`: validate, hash, save via repo, return `UserResponse`.
   - `LoginUseCase`: get by email, compare password, generate JWT.

5) Handler/Router
   - Define interface `UserUsecases` or inject per-usecase fields into `UserHandler`.
   - Add `POST /v1/auth/register` route.

6) Wire (if reintroduced)
   - Ensure proper imports/types, providers (db, repo, jwt, usecases, handlers), generate `wire_gen.go`.

7) Dockerfile
   - Use supported Go base image; expose health endpoints if needed.

### Initial security guidance
- Password hashing: bcrypt (cost ≥ 12) or argon2id; never store plaintext.
- JWT: manage keys via secret manager, set `aud/iss`, short TTL; consider refresh token rotation.
- Rate limiting: apply for login; consider temporary lockouts by IP/email.
- DB: least-privilege DB account; use SSL/TLS when connecting.
- Logs: do not log secrets/PII; structured logs; attach correlation/request ID.
- Headers: add security headers (CORS, HSTS, X-Content-Type-Options, etc.).
- Migrations: control schema changes; test rollbacks.

### Next steps (short)
- Manual wiring in `main` to ensure `/v1/auth/register` works end-to-end (DB, repo, hasher, usecases, handler, router).
- Implement `LoginUseCase` + open route `POST /v1/auth/login` (getByEmail → compare password → JWT).
- Add `.env.example`, `Makefile`, `docker-compose.yml` (Postgres), configure `golangci-lint`.
- Add middleware CORS/logger/request-id; unify error/response format.
- Add basic tests: unit (usecases), integration (repo + Postgres), HTTP handlers.

### Clean Architecture / DDD improvements
- Inversion of Control (decouple infra from application):
  - Avoid importing `internal/infras/*` in usecases.
  - Define ports in `internal/application/ports` (e.g., `TokenIssuer`, `EmailSender`, `SMSSender`, `ObjectStorage`) so usecases depend on interfaces. Implement in infra and inject in `main`.
  - JWT hardening: validate `iss/aud/nbf/exp` with leeway based on config.

- DTO and layer boundaries:
  - Keep HTTP DTOs under `interfaces/http` or use dedicated transport DTOs; let usecases accept internal input models/commands to reduce transport coupling.

- Domain invariants & VOs:
  - Validate email/role inside `user.NewUser(...)` to enforce invariants at construction time.
  - Decide on a `Password` VO usage or remove it to avoid confusion.

- Domain → HTTP error mapping:
  - Maintain a mapping table (e.g., `ErrUserNotFound` → 404, `ErrEmailAlreadyExists` → 409, `ErrInvalidRole` → 400) and apply in handlers via the `response` package.
  - Keep messages safe; avoid leaking internals; include `request_id` in logs.

- Repository robustness:
  - Add `context.WithTimeout` per query.
  - Map `sql.ErrNoRows` → `domain.ErrUserNotFound`.
  - Consider light retries for transient network errors.
  - Map Postgres unique-violation (code `23505`) → `domain.ErrEmailAlreadyExists`.

- Real health/readiness:
  - `GET /readyz` should ping DB (with timeout) and return 500 when DB is down.

- Middleware & AuthZ:
  - Use `RequireRoles(...)` to protect sample admin routes.
  - Consider splitting role→permissions into a separate policy if you need more granular RBAC.
  - RequestID: use UUIDv4; structured logging with latency, status, request_id.
  - CORS: in prod, use origin whitelist from env.
  - Trusted proxies: apply `HTTP_TRUSTED_PROXIES` to router.

- Testing & observability:
  - Unit tests for domain/usecases (mock repo/hasher/jwt); integration tests for repo with Postgres (testcontainers).
  - Add metrics (Prometheus) and tracing (OpenTelemetry) for HTTP and DB.
  - Add CI (lint/test/build/vuln-scan) and reasonable coverage thresholds.

- Dockerfile base:
  - Use an official stable Go base image (e.g., `golang:1.24-alpine`).

## Fix Checklist (Actionable)

1) CORS — single configuration source (avoid double-middleware) — DONE
- File: `internal/interfaces/http/route.go`
  - In `applyBaseMiddlewares`, do not register a default `cors.New(...)`.
  - Keep `applyCORSFromConfig(r, cfg)` and configure via `HTTP_CORS_ALLOWED_ORIGINS`.
- Env (dev): `HTTP_CORS_ALLOWED_ORIGINS=*`
- Env (prod): whitelist origin (e.g., `HTTP_CORS_ALLOWED_ORIGINS=https://app.example.com`)

2) Security headers/HSTS — only enable when TLS is present
- File: `internal/interfaces/http/middleware/security_headers.go`
  - Option A (quick): set `HTTP_SECURITY_HEADERS=false` in dev; enable true in prod.
  - Option B (proper): set HSTS only when HTTPS is detected (`X-Forwarded-Proto=https`) or `cfg.Env=="prod"`.
- Env hint (dev): `HTTP_SECURITY_HEADERS=false`

3) Graceful shutdown + HTTP timeouts — DONE
- File: `cmd/api/main.go`
  - Use `http.Server` with timeouts; handle `SIGINT/SIGTERM` and `Shutdown(15s)`; close DB.

4) Go toolchain alignment — DONE with Go 1.24
- File: `go.mod`: `go 1.24`
- File: `build/Dockerfile`: base `golang:1.24-alpine`
- File: `Readme.md`: requires Go 1.24+

5) JWT `NotBefore` (nbf) logic — DONE
- File: `internal/infras/security/jwt_service.go`
  - `GenerateToken`: set `NotBefore=now`.
  - `ValidateToken`: if `now + leeway < nbf` → `token not yet valid`.
  - Do not use `jwt.WithNotBefore()` (not in v5); check via `claims.NotBefore` manually.

6) Rate limiting — document limitations
- File: `internal/interfaces/http/middleware/ratelimit.go`
  - In-memory per-process is fine for dev/single instance. For multi-instance, use Redis-based limiter.
  - Added runtime warning in prod when login rate limit is enabled to suggest Redis limiter.

7) Consolidate Router constructor — DONE
- File: `internal/interfaces/http/route.go`
  - Single `NewRouter(userHandler, cfg, authMiddleware...)`; removed variants without `cfg`.

8) Logger global default & runtime level — DONE
- Files: `pkg/logger/logger.go`, `cmd/api/main.go`
  - Added `logger.Init` (call once at entrypoint) and `logger.L()` everywhere; `SetLevel` to change at runtime.

9) HTTP error mapping — DONE
- File: `internal/interfaces/http/response/response.go`
  - Added helpers: `NotFound`, `Conflict`.
- File: `internal/interfaces/http/handler/user_handler.go`
  - `Register`: `ErrEmailAlreadyExists` → 409 Conflict.
  - `Login`: invalid credentials → 401 Unauthorized.
  - `GetMe`: `ErrUserNotFound` → 404 Not Found.
  - `ChangePassword`: `ErrUserNotFound` → 404 Not Found; `ErrInvalidPassword` → 400 Bad Request.
  - JSON bind errors return friendly messages.

10) API Docs (dev-only) — DONE
- Serve `GET /openapi.json` (embedded OpenAPI 3.0) and `GET /swagger` (ReDoc viewer) when `ENV=dev`.
- Do not expose in production.

11) Env templates — DONE
- Template at `docs/env.example.md` (copy to `.env` at project root when needed)
- Suggested vars:
  - `ENV=dev`
  - `HTTP_PORT=8080`
  - `HTTP_CORS_ALLOWED_ORIGINS=*`
  - `HTTP_SECURITY_HEADERS=false`
  - `DB_HOST=localhost` `DB_PORT=5432` `DB_USER=appsechub` `DB_PASSWORD=devpassword` `DB_NAME=appsechub` `DB_SSLMODE=disable`
  - `MIGRATIONS_PATH=migrations`
  - `JWT_SECRET=change-me-in-dev` `JWT_EXPIRE_SEC=3600` `JWT_ISSUER=appsechub` `JWT_AUDIENCE=appsechub-clients` `JWT_LEEWAY_SEC=30`
  - `SEED_ENABLE=true` and `SEED_*` for dev admin

12) Compose dev/prod — DONE
- `docker-compose.dev.yml`: includes `db`, `redis`, bind mount + Air; `HTTP_SECURITY_HEADERS=false` to avoid HSTS locally.
- `docker-compose.yml`: includes `db`, `redis`, builds image, injects secrets via env (`JWT_SECRET`, `DB_PASSWORD`).

13) Refresh tokens (feature flag) — DONE
- Redis-backed refresh store + endpoints `/v1/auth/refresh`, `/v1/auth/logout` (only when `AUTH_REFRESH_ENABLED=true`).
- TTL configured via `REFRESH_TTL_SEC` (default 7 days).

14) Group auth routes into one function — DONE
- File: `internal/interfaces/http/route.go`
  - Added `registerAuthRoutes` covering `/register`, `/login` (with optional rate limit), and protected `/me`, `/change-password`.
  - Updated `registerAPIV1Routes` to call `registerAuthRoutes`.

## AppSecHub — Pending (High Priority)

- Observability
  - Expose `/metrics` (Prometheus)
  - Add pprof (dev-only)
  - Consider OpenTelemetry tracing for HTTP and DB

- Testing & CI
  - GitHub Actions: build, test, lint, govulncheck
  - Container image scanning (e.g., Trivy)
  - Tests:
    - Unit: domain/usecases (mock repo/hasher/jwt)
    - HTTP handlers (table-driven tests)
    - Integration: repositories with Postgres (testcontainers/compose)
  - Set reasonable coverage thresholds

- Secrets Management (Production)
  - Manage `JWT_SECRET`, `DB_PASSWORD` via a secret manager (AWS/GCP/Vault)
  - Rotate secrets periodically; inject at runtime (ENV/volumes)
  - Keep `/swagger` disabled in prod (already enforced by ENV)

## AppSecHub — Codebase Review (Initial)

### Trạng thái hiện tại (cập nhật)
- Đã hoàn thành:
  - Wiring thủ công (DB → repo → hasher/JWT → usecases → handler → router)
  - Tách `main` thành các hàm bootstrap chuyên biệt: `initPostgresAndMigrate`, `initJWTService`, `buildUserComponents`, `loadRBACPolicy`, `buildRouter`
  - Usecase: CreateUser, Login + handler/router
  - JWT claims có role; middleware JWTAuth, RequireRoles/RequirePermissions (RBAC)
  - RBAC policy in-memory + nạp từ YAML (`RBAC_POLICY_PATH`) và cảnh báo role lạ
  - Response envelope + mapping lỗi domain cơ bản; repo timeouts + map `sql.ErrNoRows`/`23505`
  - CORS/Trusted proxies theo env; RequestID UUIDv4 + structured logging (slog)
  - Health: `/healthz`; Ready: `/readyz` ping DB (timeout)
  - Docker compose dev/prod; distroless runtime; hot-reload bằng Air
  - Router chuẩn hóa: helper nhỏ `applyBaseMiddlewares`, `registerHealthRoutes`, `registerAPIV1Routes`, `registerAuthLogin` (rate limit theo config)

#### Cập nhật mới (đã triển khai)
- CORS: loại bỏ CORS mặc định ở `applyBaseMiddlewares`; chỉ còn áp dụng qua `applyCORSFromConfig` (env `HTTP_CORS_ALLOWED_ORIGINS`).
- Gom nhóm route `auth` vào một hàm: `registerAuthRoutes` (đã bao gồm `/register`, `/login` và các route bảo vệ `/me`, `/change-password`); bỏ `registerAuthLogin`.
- Thông báo lỗi JSON thân thiện: body rỗng → `"request body is empty"`; JSON sai cú pháp → `"malformed JSON at position N"`; sai kiểu → `"invalid type for field <field>"`.
- JWT: đặt `NotBefore=now` khi phát token; khi validate nếu `now + leeway < nbf` thì trả lỗi `token not yet valid` (giảm ảnh hưởng clock skew).
 - Router API: hợp nhất còn một hàm `NewRouter(userHandler, cfg, authMiddleware...)`; loại bỏ biến thể không có `cfg`.
 - Graceful shutdown: dùng `http.Server` (`ReadHeaderTimeout=5s`, `IdleTimeout=60s`) + `Shutdown(15s)` khi nhận `SIGINT/SIGTERM`; đóng `sqlDB` sau khi dừng.
- Logger: thêm global default logger (`logger.Init` một lần ở entrypoint, dùng `logger.L()` ở mọi nơi; có `SetLevel` đổi mức log runtime). Đã thay toàn bộ `log.Printf`/`slog.Warn` còn sót bằng `logger.L().*`. Không tạo logger mới trong router; dùng `logger.L()`.
 - HTTP error mapping: chuẩn hóa mã lỗi/HTTP status cho `Register`, `Login`, `GetMe`, `ChangePassword` (400/401/404/409/500) qua `response` helpers.
- IoC cho Auth: middleware `JWTAuth` nhận `TokenValidator func` thay vì phụ thuộc trực tiếp vào infra JWT; validator được bọc/tiêm ở composition root.
- Đổi Application “service” → “ports”: `TokenIssuer`, `EmailSender`, `SMSSender`, `ObjectStorage`. Cập nhật import & build OK.
- Bcrypt cost theo env: `BCRYPT_COST` (4–31), mặc định dùng `bcrypt.DefaultCost` nếu không set/không hợp lệ.
 - Config singleton: `config.Load()` parse 1 lần (sync.Once); inject `*config.Config` từ composition root, không gọi `Load()` trong leaf code. Đã chỉnh `buildUserComponents` để nhận `cfg` và truyền `cfg.Security.BcryptCost` vào `NewBcryptHasher`.
- Security headers: thêm CSP mặc định `default-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'`; HSTS chỉ bật khi HTTPS (TLS hoặc `X-Forwarded-Proto=https`).

- Còn lại/Khuyến nghị tiếp:
  - ~~Nạp RBAC từ YAML ngay trong `main` khi có `RBAC_POLICY_PATH`~~ (đã nạp)
  - ~~Bảo vệ nhóm `/v1/admin`: chain `JWTAuth(jwtSvc)` trước `RequirePermissions(...)`~~ (đã thêm)
  - ~~Loại bỏ middleware tự đặt `X-Request-Id` trong `NewRouter`; chỉ dùng `middleware.RequestID()`~~ (đã bỏ)
  - ~~JWT hardening: iss/aud/nbf + leeway~~ (đã thêm; còn refresh tokens nếu cần)
  - CI/CD: build/test/lint/govulncheck + Docker image build & image scan (GitHub Actions)
  - Chất lượng: `golangci-lint` + `Makefile` targets (build/run/test/lint/migrate)
  - Testing: unit (domain/usecase), HTTP handlers, integration repo với Postgres (testcontainers/compose)
  - Observability: Prometheus metrics (kèm `/metrics`), OpenTelemetry tracing (HTTP/DB), pprof (dev-only)
  - ~~Security headers, rate limit (login)~~ (đã thêm)
  - Security nâng cao: refresh token/rotation; tăng cost bcrypt theo env hoặc chuyển argon2id
  - API Docs: OpenAPI/Swagger, route `/swagger` (dev-only)
  - DevEx: scripts migrate tiện (make migrate-up/down), seed script/command
  - ĐÃ: thêm cơ chế seed user khởi tạo qua biến env `SEED_*` (tạo admin nếu chưa tồn tại)

### Ghi chú hậu chuyển đổi (bỏ Wire)
- Đã bỏ hoàn toàn Wire; wiring thủ công trong `cmd/api/main.go`.
- Cần rà soát và loại bỏ tham chiếu Wire còn lại trong tài liệu để tránh gây hiểu nhầm.

### Kiến trúc hiện tại (tổng quan)
- **Phân lớp**: `internal/domain` (Entity/Value Object), `internal/application` (UseCase/DTO/Service), `internal/interfaces/http` (Gin router/handler/middleware), `internal/infras` (DB migrate, Security/JWT), `internal/config` (load env).
- **HTTP**: Gin router tại `internal/interfaces/http/route.go`, handler người dùng tại `internal/interfaces/http/handler/user_handler.go`.
- **Domain**: `internal/domain/user` với `User`, `Email`, `Role`, và bộ lỗi domain.
- **Security**: JWT service tại `internal/infras/security/jwt_service.go` và middleware xác thực tại `internal/interfaces/http/middleware/auth.go`.
- **DB**: Migrator dùng `golang-migrate` tại `internal/infras/db/migrator.go`.
- **Config**: `internal/config/config.go` + `internal/config/loader.go` (nạp env bằng `caarlos0/env`).

### Trạng thái build hiện tại
- Build: OK trên Go 1.24.

### Vấn đề chính cần khắc phục
1) (đã xử lý ở code hiện tại) — mục này được loại bỏ khỏi danh sách.
2) (đã xử lý ở code hiện tại) — mục này được loại bỏ khỏi danh sách.

3) **UseCase/DTO chưa đồng bộ**
   - `internal/application/usecase/userusecase/create_user.go`:
     - Thiếu định nghĩa `PasswordHasher`.
     - Dùng `dto.CreateUserRequest` nhưng DTO này chưa tồn tại.
     - Gọi `user.NewUser(input.Email, hashed)` sai chữ ký (domain yêu cầu `firstName, lastName, email, password, role`).
   - `internal/application/dto/user_dto.go` chưa có `CreateUserRequest`; có `LoginRequest`, `LoginResponse`, `UserResponse`.

4) **Handler phụ thuộc type/usecase không tồn tại**
   - `internal/interfaces/http/handler/user_handler.go`:
     - Field `Usecase userusecase.UserUsecase` không tồn tại (chưa định nghĩa interface/struct `UserUsecase`).
     - Dùng `h.Usecase.Login.Execute(...)` nhưng `Login` usecase chưa có.

5) **Auth service import sai**
   - `internal/application/service/auth_service.go` import `appsechub/internal/domain` (không tồn tại). Domain nằm ở `internal/domain/user`.

6) **Migrations trống và đường dẫn**
   - Thư mục `migrations/` chỉ có `.gitkeep`. `m.Up()` sẽ lỗi nếu không có file hợp lệ hoặc đường dẫn sai.
   - `main` đang truyền `cfg.DB.MigrationsPath` (không có). Cần thêm vào `Config` hoặc mặc định `./migrations`.

7) (đã xử lý) — đã hợp nhất constructor router và wiring đúng `*handler.UserHandler`.

8) Dockerfile: đang dùng `golang:1.24-alpine` ổn định; healthcheck HTTP có thể cân nhắc thêm sau.

### Đề xuất lộ trình sửa (ưu tiên)
1) **Sửa cấu hình và entrypoint**
   - Thêm `MigrationsPath` vào `Config` hoặc mặc định `./migrations`.
   - Thêm hàm build DSN từ `DBConfig` (Postgres). Sửa `main` dùng `cfg.HTTP.Port`.

2) **Thêm storage Postgres và repository**
   - Tạo `internal/infras/storage/postgres/` với:
     - `NewPostgresConnection(cfg *config.Config) (*sql.DB, error)`.
     - `user_repository.go` implement `internal/domain/user.Repository`.

3) **Hoàn thiện migrations**
   - Thêm `migrations/0001_init.up.sql` và `.down.sql` cho bảng `users` (email unique, hashed password, role, timestamps).

4) **Hoàn thiện UseCase/DTO**
   - Thêm `dto.CreateUserRequest { FirstName, LastName, Email, Password, Role }`.
   - Định nghĩa `PasswordHasher` (hoặc dùng service riêng) và implement (bcrypt/argon2id).
   - Sửa `CreateUserUseCase.Execute` để validate, hash, lưu repo, trả `UserResponse`.
   - Thêm `LoginUseCase` (get by email, compare password, generate JWT).

5) **Sửa Handler/Router**
   - Định nghĩa interface `UserUsecase` hoặc truyền riêng từng usecase vào `UserHandler`.
   - Bổ sung route `POST /v1/auth/register` cho `Register`.

6) **Sửa Wire**
   - Sửa import đúng `usecase` và types, thêm providers (db, repo, jwt, usecases, handlers), generate `wire_gen.go`.

7) **Cập nhật Dockerfile**
   - Dùng base image Go hợp lệ (vd `golang:1.22-alpine`).
   - Bỏ hoặc thay `HEALTHCHECK` bằng endpoint HTTP.

### Gợi ý bảo mật ban đầu
- **Hash mật khẩu**: dùng `bcrypt` (cost ≥ 12) hoặc `argon2id`; không lưu plaintext.
- **JWT**: key từ secret manager, set `aud/iss`, thời hạn ngắn; cân nhắc refresh token/rotation.
- **Rate limiting**: áp dụng cho login; lockout tạm thời theo IP/email.
- **DB**: tài khoản DB with least privilege; dùng SSL/TLS khi kết nối.
- **Logs**: không log secrets/PPI; structured logging; gắn correlation ID.
- **Headers**: thêm security headers (CORS, HSTS, X-Content-Type-Options, ...).
- **Migrations**: kiểm soát schema changes; test rollback.

### Việc nên làm tiếp theo (ngắn gọn)
- Manual wiring trong `main` để endpoint `/v1/auth/register` hoạt động thực tế (khởi tạo DB, repo, hasher, usecase, handler, router).
- Implement `LoginUseCase` + mở route `POST /v1/auth/login` (getByEmail → compare password → JWT).
- Thêm `.env.example`, `Makefile`, `docker-compose.yml` (Postgres), cấu hình `golangci-lint`.
- Bổ sung middleware CORS/logger/request-id; thống nhất format lỗi/response.
- Thêm tests cơ bản: unit (usecases), integration (repo + Postgres), HTTP handlers.

### Đề xuất cải thiện Clean Architecture / DDD
- Inversion of Control (tách hạ tầng khỏi application):
  - Tránh `usecase` import trực tiếp `internal/infras/*`.
  - Tạo ports trong `internal/application/ports` (ví dụ: `TokenIssuer`, `EmailSender`, `SMSSender`, `ObjectStorage`), để usecase chỉ phụ thuộc interface. Hạ tầng implement và inject ở `main`.
  - JWT hardening: thêm kiểm tra `iss/aud/nbf/exp` (có leeway) trong `ValidateToken` qua cấu hình.

- DTO và biên giới layer:
  - Giữ DTO HTTP ở `interfaces/http` hoặc dùng DTO riêng cho transport; `usecase` có thể nhận “input model/command” nội bộ để giảm phụ thuộc transport.

- Domain invariants và VO:
  - Đưa validate email/role vào `user.NewUser(...)` để đảm bảo bất biến ngay khi khởi tạo (kể cả tên không rỗng).
  - Quyết định với `Password` VO: hoặc dùng thật trong `User` (và cập nhật repo scan/insert), hoặc loại bỏ VO này để tránh rối.

- Mapping lỗi domain → HTTP:
  - Xây dựng bảng ánh xạ (ví dụ `ErrUserNotFound` → 404, `ErrEmailAlreadyExists` → 409, `ErrInvalidRole` → 400) và áp dụng tại handler qua package `response`.
  - Giữ message an toàn, tránh lộ chi tiết nội bộ; log nội bộ kèm `request_id`.

  ~- Cần ánh xạ trong handler~
  - ĐÃ: ánh xạ cơ bản trong `UserHandler.Register` và `UserHandler.Login`; `ErrEmailAlreadyExists` map từ Postgres `23505`; `ErrInvalidEmail` dùng trong `NewEmail`.

- Repository robustness:
  - ~~Thêm `context.WithTimeout` cho từng truy vấn.~~
  - ~~Map `sql.ErrNoRows` → `domain.ErrUserNotFound`.~~
  - Cân nhắc retry nhẹ (idempotent) với lỗi network tạm thời. (chưa)
  - ~~Map lỗi unique-violation Postgres (code `23505`) → `domain.ErrEmailAlreadyExists`.~~

- Health/Ready thực sự:
  - ~~`GET /readyz` ping DB (với timeout) để phản ánh readiness đúng; trả 500 khi DB down.~~

- Middleware & AuthZ:
  - Dùng `RequireRoles(...)` để bảo vệ route mẫu (ví dụ nhóm `/v1/admin`).
  - Cân nhắc tách policy role→permissions nếu cần chi tiết hơn RBAC đơn giản.
  - ~~RequestID: dùng UUIDv4 thay vì timestamp; thêm middleware logger structured (zap/slog) kèm `request_id`, status, latency.~~
  - ~~CORS: ở prod dùng whitelist qua env thay vì `*`.~~
  - ~~Trusted proxies: đọc `HTTP_TRUSTED_PROXIES` và áp dụng vào `SetTrustedProxies`.~~

- Testing & quan sát:
  - Unit test cho domain/usecase (mock repo/hasher/jwt), integration test repo với Postgres (testcontainers).
  - Thêm metrics (Prometheus) và tracing (OpenTelemetry) ở HTTP và DB.
  - Thêm CI (lint/test/build/vuln-scan) và threshold coverage hợp lý.

- Dockerfile base:
  - Dùng base Go chính thức sẵn có (ví dụ `golang:1.22-alpine`) thay vì phiên bản chưa ổn định.

### Lộ trình hành động ngắn
1) Tạo `internal/application/service/token_service.go` và cập nhật `LoginUseCase` dùng interface này; inject `security.JWTService` ở `main`.
2) Dồn validation vào `user.NewUser(...)`; chuẩn hóa mapping lỗi domain → HTTP trong handler (thêm bảng map error→status).
3) Repo: map `sql.ErrNoRows` và Postgres `23505`; giữ timeout; thêm ping DB thật trong `/readyz`.
4) Middleware: chuyển RequestID sang UUIDv4; thêm logger structured; CORS whitelist từ env; `HTTP_TRUSTED_PROXIES` vào router.
5) Viết test cơ bản cho `CreateUserUseCase` và `LoginUseCase` (mock repo/hasher/jwt) và 1–2 test handler; thêm CI đơn giản.

Tài liệu này sẽ được cập nhật sau mỗi lần khắc phục một cụm lỗi lớn để tiện theo dõi tiến độ.


## Fix Checklist (Actionable)

~~1) CORS – một nguồn cấu hình duy nhất (tránh double-middleware)~~ (ĐÃ THỰC HIỆN)
- File: `internal/interfaces/http/route.go`
  - Trong `applyBaseMiddlewares`, gỡ khối `r.Use(cors.New(...))` mặc định `AllowOrigins: "*"`.
  - Giữ nguyên `applyCORSFromConfig(r, cfg)` và cấu hình qua env `HTTP_CORS_ALLOWED_ORIGINS`.
- Env (dev): `HTTP_CORS_ALLOWED_ORIGINS=*`
- Env (prod): whitelist origin (vd: `HTTP_CORS_ALLOWED_ORIGINS=https://app.example.com`)
  

2) Security headers/HSTS – chỉ bật khi có TLS
- File: `internal/interfaces/http/middleware/security_headers.go`
  - Tùy chọn A (nhanh): đặt `HTTP_SECURITY_HEADERS=false` cho dev (compose dev); bật true ở prod.
  - Tùy chọn B (chuẩn): chỉ set HSTS khi nhận diện HTTPS (vd kiểm `X-Forwarded-Proto=https`) hoặc `cfg.Env=="prod"`.
- Env gợi ý (dev): `HTTP_SECURITY_HEADERS=false`

~~3) Graceful shutdown + HTTP timeouts~~ (ĐÃ THỰC HIỆN)
- File: `cmd/api/main.go`
  - Dùng `http.Server` với timeouts; xử lý `SIGINT/SIGTERM` và `Shutdown(15s)`; đóng DB.

~~4) Go toolchain đồng bộ~~ (ĐÃ THỰC HIỆN với Go 1.24)
- File: `go.mod`: `go 1.24`
- File: `build/Dockerfile`: base `golang:1.24-alpine`
- File: `Readme.md`: yêu cầu Go 1.24+

~~5) JWT `NotBefore` (nbf) logic rõ ràng~~ (ĐÃ THỰC HIỆN)
- File: `internal/infras/security/jwt_service.go`
  - `GenerateToken`: set `NotBefore=now`.
  - `ValidateToken`: nếu `now + leeway < nbf` → `token not yet valid`.
  - Không dùng `jwt.WithNotBefore()` (không có trong v5); kiểm tra thủ công với `claims.NotBefore`.

6) Rate limiting – ghi chú giới hạn
- File: `internal/interfaces/http/middleware/ratelimit.go`
  - Hiện tại in-memory per-process (ổn cho dev/single instance). Với multi-instance, cân nhắc Redis-based limiter (triển khai sau khi scale).
  - ĐÃ THÊM cảnh báo runtime ở prod khi bật login rate limit: log nhắc dùng distributed limiter (Redis) cho multi-instance.

~~14) Hợp nhất constructor Router~~ (ĐÃ THỰC HIỆN)
- File: `internal/interfaces/http/route.go`
  - Còn duy nhất `NewRouter(userHandler, cfg, authMiddleware...)`; bỏ `NewRouterWithConfig` và biến thể không có `cfg`.

~~15) Logger global default & runtime level~~ (ĐÃ THỰC HIỆN)
- File: `pkg/logger/logger.go`, `cmd/api/main.go`
  - Thêm `logger.Init` (gọi 1 lần ở entrypoint) và `logger.L()` dùng ở mọi nơi; `SetLevel` đổi mức log runtime.

~~7) Mapping lỗi HTTP~~ (ĐÃ THỰC HIỆN)
- File: `internal/interfaces/http/response/response.go`
  - Thêm helpers: `NotFound`, `Conflict`.
- File: `internal/interfaces/http/handler/user_handler.go`
  - `Register`: `ErrEmailAlreadyExists` → 409 Conflict.
  - `Login`: invalid credentials → 401 Unauthorized.
  - `GetMe`: `ErrUserNotFound` → 404 Not Found.
  - `ChangePassword`: `ErrUserNotFound` → 404 Not Found; `ErrInvalidPassword` → 400 Bad Request.
  - JSON bind errors: thông điệp thân thiện (đã thực hiện trước).

8) Observability
- Thêm endpoint `/metrics` (Prometheus) qua middleware/handler riêng (dev/prod đều dùng).
- Cân nhắc OpenTelemetry tracing cho HTTP và DB (gin middleware + pgx/driver hook nếu cần), dev bật sampling cao; prod tùy chỉnh.

9) Testing & CI
- ~~Thêm `golangci-lint` config~~ (ĐÃ THỰC HIỆN: `.golangci.yml`)
- ~~Makefile targets~~ (ĐÃ THỰC HIỆN: `build/run/test/fmt/vet/lint/tools`, `dev-up/down`, `prod-up/down`)
- (CÒN LẠI) GH Actions: build/test/lint/govulncheck; image scan (trivy) tùy nhu cầu.
- Viết tests:
  - Unit: domain/usecases (mock repo/hasher/jwt).
  - HTTP handlers: Gin + `httptest` (table-driven).
  - Integration: repo Postgres bằng testcontainers-go hoặc compose.
- Makefile targets: `build`, `run`, `test`, `lint`, `migrate-up`, `migrate-down`.

10) API Docs (dev-only)
~~Tích hợp Swagger/OpenAPI (dev-only)~~ (ĐÃ THỰC HIỆN)
- Serve `GET /openapi.json` (embedded OpenAPI 3.0) và `GET /swagger` (ReDoc viewer) khi `ENV=dev`.
- Không expose ở production.

~~11) Env templates~~ (ĐÃ THỰC HIỆN)
- Đã thêm template tại `docs/env.example.md` (copy nội dung ra `.env` ở project root khi cần)
- Gợi ý biến:
  - `ENV=dev`
  - `HTTP_PORT=8080`
  - `HTTP_CORS_ALLOWED_ORIGINS=*`
  - `HTTP_SECURITY_HEADERS=false`
  - `DB_HOST=localhost` `DB_PORT=5432` `DB_USER=appsechub` `DB_PASSWORD=devpassword` `DB_NAME=appsechub` `DB_SSLMODE=disable`
  - `MIGRATIONS_PATH=migrations`
  - `JWT_SECRET=change-me-in-dev` `JWT_EXPIRE_SEC=3600` `JWT_ISSUER=appsechub` `JWT_AUDIENCE=appsechub-clients` `JWT_LEEWAY_SEC=30`
  - `SEED_ENABLE=true` và bộ `SEED_*` cho admin dev

~~12) Compose dev/prod~~ (ĐÃ THỰC HIỆN)
- `docker-compose.dev.yml`: có `db`, `redis`, bind mount + Air; `HTTP_SECURITY_HEADERS=false` để tránh HSTS local.
- `docker-compose.yml`: có `db`, `redis`, build image, inject secrets qua env (`JWT_SECRET`, `DB_PASSWORD`).

~~13) Refresh tokens (feature flag)~~ (ĐÃ THỰC HIỆN)
- Redis-backed refresh store + endpoints `/v1/auth/refresh`, `/v1/auth/logout` (chỉ bật khi `AUTH_REFRESH_ENABLED=true`).
- TTL cấu hình qua `REFRESH_TTL_SEC` (mặc định 7 ngày).

~~13) Gom nhóm route `auth` vào một hàm~~ (ĐÃ THỰC HIỆN)
- File: `internal/interfaces/http/route.go`
  - Thêm `registerAuthRoutes` gộp `/register`, `/login` (kèm rate limit khi có `cfg`), và các route bảo vệ `/me`, `/change-password`.
  - Cập nhật `registerAPIV1Routes` gọi `registerAuthRoutes`; bỏ hàm riêng `registerAuthLogin`.

Thứ tự khuyến nghị thực hiện: (1) CORS, (2) Security headers/HSTS, (3) Graceful shutdown, (4) Go toolchain, (5) JWT nbf, (7) Mapping lỗi, (11) Env template; sau đó nâng cấp Observability/Tests/CI.

