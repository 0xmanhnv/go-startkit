## Step-by-step implementation plan

This roadmap helps you evolve the starter kit gradually. Follow stages in order; each stage should keep the app buildable and runnable.

### Stage 0 – Prerequisites
- Copy `.env.example` → `.env` (adjust secrets).
- Dev run: `docker compose -f docker-compose.dev.yml up`
- Prod (reference): `docker compose up -d --build`
 - (Optional) Enable seeding initial user by setting `SEED_ENABLE=true` and `SEED_USER_*` variables.

### Stage 1 – Manual wiring (enable register endpoint end-to-end)
- ~~Wire dependencies in `cmd/api/main.go` (DB → repo → hasher → usecase → handler → router) instead of passing `nil`.~~
- ~~Refactor `main` thành bootstrap functions (`initPostgresAndMigrate`, `initJWTService`, `buildUserComponents`, `loadRBACPolicy`, `buildRouter`).~~
- Minimal example (adapt names as needed):
```go
// in main.go
dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", ...)
sqlDB, err := postgres.NewPostgresConnection(dsn)
if err != nil { log.Fatal(err) }
userRepo := postgres.NewUserRepository(sqlDB)
hasher := security.NewBcryptHasher()
createUC := &userusecase.CreateUserUseCase{repo: userRepo, hasher: hasher}
userHandler := handler.NewUserHandler(createUC)
router := httpiface.NewRouter(userHandler)
```

- ### Stage 2 – Login use case + route
- ~~Implement `LoginUseCase`:~~
  - Input: `dto.LoginRequest{Email, Password}`; Output: `dto.LoginResponse{AccessToken, User}`.
  - Steps: `repo.GetByEmail` → `hasher.Compare` → `jwt.GenerateToken(user.ID, user.Role)`.
- ~~Handler: add `Login` method and wire route `POST /v1/auth/login`.~~

 - ### Stage 3 – Authorization (RBAC)
 - ~~Add middleware `RequireRoles(roles...)` that reads role from JWT claims/context.~~
 - ~~Include role in token claims (custom claim) when generating JWT.~~
 - Protect sample routes with `RequireRoles("admin")`.
 - [ ] Chain `JWTAuth(jwtSvc)` trước `RequirePermissions(...)` ở nhóm `/v1/admin`.
 - [ ] Nạp RBAC từ YAML trong `main` khi có `RBAC_POLICY_PATH`.

### Stage 4 – Response & error standardization
- ~~Introduce a response helper: `{ data, error: {code, message}, meta }`.~~
- ~~Map domain errors → proper HTTP status codes (e.g., `ErrUserNotFound`→404, `ErrEmailAlreadyExists`→409, `ErrInvalidRole`→400).~~

 ### Stage 5 – HTTP middleware baseline
 - ~~Add middleware: RequestID (UUIDv4), logger (structured)~~, CORS (whitelist from env), gzip (optional), rate limit (for login), secure headers. (CORS theo env đã có)
 - [ ] Loại bỏ middleware set `X-Request-Id` thủ công trong `NewRouter` (đã có `middleware.RequestID()`).
- ~~Chuẩn hóa router thành các helper (`applyBaseMiddlewares`, `registerHealthRoutes`, `registerAPIV1Routes`, `registerAuthLogin`).~~
 - Keep `SetTrustedProxies(nil)` by default; configure via env (`HTTP_TRUSTED_PROXIES`) for real proxies. (đã có)

### Stage 6 – Health/Ready endpoints
- ~~`GET /healthz`: lightweight (always 200 if process is up).~~
- ~~`GET /readyz`: checks DB connectivity (`db.PingContext` with timeout). Return 500 when DB down.~~

### Stage 7 – Tooling & CI
- Add `Makefile` targets: build/run/test/lint/migrate.
- Add `golangci-lint` config and `pre-commit` hooks.
- CI (GitHub Actions): build + test + lint + govulncheck + Docker image build + image scan.
 - [ ] Thêm skeleton tests (domain/usecase/handlers) và chạy trong CI.

### Stage 8 – Observability
- Metrics (Prometheus): HTTP latency/count, DB latency, error counts.
- Tracing (OpenTelemetry): HTTP + DB spans.
- pprof (dev only), behind admin-only route if needed.
 - [ ] Thêm endpoint `/metrics` và cấu hình Prometheus scrape.

### Stage 9 – Testing
- Unit tests: domain and usecases (mock repo/hasher/jwt).
- Integration tests: repository using `testcontainers-go` or compose.
- HTTP handler tests: Gin + `httptest`, table-driven tests.

 ### Stage 10 – Security enhancements
 - Switch to argon2id (optional) or increase bcrypt cost in prod.
 - Secrets management (Vault/KMS); rotate secrets.
 - Add gosec, govulncheck; image scan (trivy/grype); secret scan (gitleaks); generate SBOM (syft).
 - [ ] JWT hardening: kiểm tra `iss/aud/nbf` + leeway cho `exp`.
 - [ ] Refresh token/rotation flow.

### Stage 11 – API documentation
- Generate OpenAPI/Swagger (swaggo/kin-openapi), expose `/swagger` in dev only.

