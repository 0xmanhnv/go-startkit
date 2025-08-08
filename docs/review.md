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
- Thêm `golangci-lint` config; GH Actions: build/test/lint/govulncheck; image scan (trivy) tùy nhu cầu.
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

