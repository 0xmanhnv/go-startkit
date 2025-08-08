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
- Build đang lỗi. Lỗi đầu tiên khi chạy `go build ./...`:
  - package `appsechub/internal/domain` không tồn tại (import sai trong `internal/application/service/auth_service.go`).
- Nhiều lỗi wiring/thiếu implement khác sẽ xuất hiện tiếp theo (liệt kê bên dưới).

### Vấn đề chính cần khắc phục
1) **Cấu hình/Entrypoint không khớp**
   - `cmd/api/main.go` dùng các field không tồn tại trong `Config`:
     - `cfg.Port`, `cfg.DB.DSN`, `cfg.DB.MigrationsPath` không có trong `internal/config/config.go`.
     - `Config` hiện có `HTTP.Port`, `DB.{Host,Port,User,Password,Name,SSLMode}` và `JWT.{Secret,ExpireSec}`.
   - Hành động: build `DSN` từ `DBConfig`, thêm `MigrationsPath` (vd `./migrations`) hoặc hardcode mặc định, và dùng `cfg.HTTP.Port` khi `router.Run`.

2) **DI/Wire sai import và type** (`cmd/api/wire.go`)
   - Import sai: `internal/application/usecases/userusecase` (thực tế là `usecase` không có "s").
   - Import package không tồn tại: `internal/infras/storage/postgres`.
   - Trả về kiểu không tồn tại: `*handler.Handler`. Trong code chỉ có `*handler.UserHandler`.
   - Thiếu providers: `NewPostgresConnection`, `NewUserRepository`, `NewUserUsecase` chưa được định nghĩa.

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

7) **Router/DI không khớp**
   - `internal/interfaces/http/route.go` cần `*handler.UserHandler` trong `NewRouter`, nhưng `InitHandler` hiện trả `*handler.Handler` (không tồn tại).

8) **Dockerfile chưa thực tế** (`build/Dockerfile`)
   - Dùng `golang:1.24.5-alpine` (phiên bản chưa tồn tại thời điểm hiện tại). Nên dùng `1.22` hoặc phiên bản Go mục tiêu thực tế.
   - `HEALTHCHECK` gọi `"/app/appsechub" "health"` nhưng binary không có lệnh này.

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
  - Tạo interface `TokenService` trong `internal/application/service` (ví dụ: `GenerateToken(userID, role string) (string, error)`), để `LoginUseCase` chỉ phụ thuộc interface. Hạ tầng (`security.JWTService`) implement và inject ở `main`.
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

