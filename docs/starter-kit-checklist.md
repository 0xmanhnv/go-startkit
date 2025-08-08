## Go Starter Kit — Best Practical Checklist

### 1) Config & Environment
- Add `MigrationsPath` (đã có) và DSN builder từ `DBConfig` (đã áp dụng trong `main`).
- Thêm `./.env.example` với các biến: `HTTP_PORT`, `DB_HOST/PORT/USER/PASSWORD/NAME/SSLMODE`, `JWT_SECRET`, `MIGRATIONS_PATH`.
- Dùng `caarlos0/env` (đã dùng) và chặn thiếu biến bắt buộc (secret, DB) khi ở `ENV=prod`.

### 2) Database & Migrations
- Đã thêm `migrations/0001_init.{up,down}.sql` tạo bảng `users`.
- Viết helper dựng DSN (đã tạm trong `main`) hoặc hàm `config.PostgresDSN()`.
- Thêm `scripts/migrate.sh` (up/down/force/version) để thao tác tiện.

### 3) Repository Layer
- ~~Đã thêm `internal/infras/storage/postgres` với `connection.go` và `user_repository.go` implement `user.Repository`.~~
- ~~Cần bổ sung context timeouts (vd `ctx, cancel := context.WithTimeout(ctx, 3*time.Second)`).~~

### 4) Usecases & Services
- `CreateUserUseCase` đã chuẩn hóa input/validation/hash/save/response.
- ~~Thêm `LoginUseCase`~~:
  - ~~GetByEmail → compare password (bcrypt) → generate JWT → trả `dto.LoginResponse`.~~
  - ~~Định nghĩa interface `TokenIssuer` (hoặc dùng `security.JWTService` trực tiếp).~~
- Chuẩn hóa interface `UserUsecase` (gom nhóm usecase) hoặc inject từng usecase vào handler.
 - ~~PasswordHasher interface~~ dùng ở usecases; implement bằng bcrypt ở infra.
 - ~~TokenIssuer (application interface)~~ được implement bởi JWT infra.
 - [ ] Thêm use case tham khảo: `GetMe`, `ChangePassword`, `UpdateProfile`, `ListUsers` (pagination, filtering).
 - [ ] Bao phủ context/timeouts cho usecases (đã có ở repo); không log trong usecases; chỉ trả lỗi domain/application.
 - [ ] Transaction boundary cho các usecase cần atomic (thiết kế `UnitOfWork`/`TxManager` abstraction ở application; implement ở infra).
 - [ ] `Clock`/`TimeProvider` interface để test deterministic timestamps.
 - [ ] Mock interfaces (repo/hasher/token/clock) để phục vụ unit test.

### 5) HTTP Layer
- ~~Router hiện expose `POST /v1/auth/register`; mở `POST /v1/auth/login` khi có `LoginUseCase`.~~
- ~~Middleware: JWT (đã có), thêm CORS, request ID, recovery, logging.~~
- ~~Chuẩn hóa response envelope và error mapping (4xx/5xx) + validation errors.~~
  - ~~Thêm security headers (HSTS, X-Content-Type-Options, Referrer-Policy, X-Frame-Options nếu cần)~~
  - ~~Rate limit cho `/v1/auth/login`~~

### 6) Dependency Injection (Wire) hoặc Manual Wiring
 - (Đã bỏ Wire) Manual wiring trong `main` (khuyến nghị):
```go
// main.go (wiring tối thiểu)
dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", ...)
dbConn, err := postgres.NewPostgresConnection(dsn)
if err != nil { log.Fatal(err) }
repo := postgres.NewUserRepository(dbConn)
hasher := security.NewBcryptHasher(cfg.Security.BcryptCost)
createUC := &userusecase.CreateUserUseCase{repo: repo, hasher: hasher}
userHandler := handler.NewUserHandler(createUC)
router := httpiface.NewRouter(userHandler, cfg)
```

### 7) Security Baseline
- Password hashing: bcrypt (đã thêm), cost ≥ 12 cho prod.
- JWT: secret từ secret manager; set `iss/aud`; TTL ngắn; cân nhắc refresh tokens.
  - ~~Thêm kiểm tra `iss/aud/nbf` + leeway trong `ValidateToken`~~
- Rate limit login, tạm khóa theo IP/user sau nhiều lần sai; log nghi ngờ.
- CORS: whitelist origin; secure headers (HSTS, X-Content-Type-Options, Referrer-Policy...).
- DB: user least-privilege; bật TLS nếu có; tránh superuser.
- Logging: không log secrets/PPI; thêm request-id; scrub payloads nhạy cảm.

### 8) Observability
- Structured logging (zap/logrus/slog); middleware access log.
- Metrics (Prometheus): HTTP requests, DB latency, errors; endpoint `/metrics`.
- Tracing (OTel): HTTP + DB spans.
 - ~~Structured access logging (slog) đã có; còn lại: metrics/tracing.~~

### 9) Testing
- Unit tests: domain, usecases (mock repo/hasher/jwt).
- Integration tests: repo với `testcontainers` hoặc `docker-compose`.
- HTTP handler tests: Gin + httptest; table-driven.
 - [ ] Thiết lập bộ skeleton test (domain/usecase/handlers) và chạy trong CI.
 - [ ] Integration test tối thiểu cho `user_repository` với Postgres thực.

### 10) Tooling & Quality
- `golangci-lint` config; chạy trong CI.
- `make` hoặc `task` targets: `build`, `run`, `test`, `lint`, `migrate-up/down`.
- Pre-commit hooks: fmt, vet, lint.
 - [ ] Thêm GitHub Actions: build/test/lint/govulncheck, build & scan Docker image.
 - [ ] Makefile targets: build/run/test/lint/migrate-up/migrate-down

### 11) Docker & Dev Experience
- Dockerfile: dùng `golang:1.22-alpine` (hoặc version bạn target); bỏ healthcheck custom hoặc dùng HTTP.
- `docker-compose.yml`: Postgres + app + admin (pgAdmin) cho dev.
- `.air.toml`/`air` (hot reload) tùy nhu cầu.

### 12) API Docs
- OpenAPI/Swagger với `swaggo` hoặc `kin-openapi`; route `/swagger` (dev only).
 - [ ] Thêm generator và wire route `/swagger` (dev-only).

---

## Việc nên làm tiếp theo (ngắn gọn)
- ~~Manual wiring trong `main` theo snippet để chạy `/v1/auth/register` (DB → repo → hasher → usecase → handler → router).~~
- ~~Implement `LoginUseCase` + mở route `POST /v1/auth/login` (GetByEmail → Compare → JWT).~~
- Thêm `.env.example`, `Makefile`, `docker-compose.yml` (Postgres), cấu hình `golangci-lint`. (compose đã có; Makefile/lint còn thiếu)
- Cải tiến middleware: ~~CORS (đọc từ env)~~, RequestID (đã có dạng đơn giản), logger structured (chưa), thống nhất error/response (đã có envelope, đã map lỗi domain cơ bản).
- Viết tests cơ bản: unit (usecase), integration (repo), HTTP (handlers). (chưa)
- Cập nhật Dockerfile (Go version hợp lệ), cân nhắc thêm Swagger/OpenAPI cho dev. (Dockerfile đã chuẩn distroless; Swagger chưa)

