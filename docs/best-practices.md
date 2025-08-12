## Best Practical (bảo mật & vận hành) — cập nhật dần

### Email validation
- Dùng `strict_email` (dựa trên `net/mail`): cấm display name, trim, so khớp địa chỉ.
- Áp dụng cho DTO nhạy cảm: `CreateUserRequest.Email`, `LoginRequest.Email`.
- Bổ sung test với edge-cases (khoảng trắng, tên miền con, ký tự đặc biệt hợp lệ).

### Bcrypt (mật khẩu)
- Interface `PasswordHasher.Hash` phải trả lỗi; không nuốt lỗi từ thư viện.
- `BCRYPT_COST` qua env; khuyến nghị: dev 10–12, prod ≥ 12 (benchmark theo hạ tầng).
- Cân nhắc Argon2id khi phù hợp (memory-hard); benchmark trước khi chuyển.

### Refresh token (Redis)
- Map `redis.Nil` → `ErrInvalidRefreshToken`; `Revoke` idempotent.
- Bật rotation-by-use (mỗi refresh sinh token mới, thu hồi token cũ).
- (Tùy chọn) Lưu refresh token dưới dạng hash để giảm rủi ro rò rỉ.
- Thêm test cho nhánh lỗi Redis và các hành vi invalid.

### JWT (HS256 → RS256/EdDSA, kid, rotation)
- Cho phép cấu hình thuật toán qua `JWT_ALG` (`HS256` mặc định; hỗ trợ `RS256`, `EdDSA`). Các biến liên quan:
  - `JWT_KID`: id của khóa ký hiện tại, gắn vào header.
  - `JWT_PRIVATE_KEY_PATH` hoặc `JWT_PRIVATE_KEY_PEM`: private key (RS/EdDSA).
  - `JWT_PUBLIC_KEYS_DIR`: thư mục chứa public keys (tên file = `kid`).
- Gắn `kid` vào header khi ký; verify dựa trên `kid` và tập public keys.
- Nguồn khóa:
  - Ưu tiên đường dẫn file/secret manager (Vault/KMS); tránh nhúng PEM dài trong env.
  - Quyền file chặt (`0400`), chủ sở hữu process.
  - Hỗ trợ hot-reload (SIGHUP) nếu cần.
- Rotation thực tế theo pha:
  1) Bắt đầu phát hành bằng khóa/thuật toán mới (`kid` mới), vẫn verify khóa cũ (đa public keys).
  2) Sau thời gian ân hạn, ngừng verify khóa cũ.
- Chính sách verify:
  - Whitelist thuật toán (không chấp nhận khác với `JWT_ALG`).
  - Yêu cầu `kid` khi có >1 key; tránh fallback trừ khi chỉ có đúng 1 key cấu hình.
  - Validate đủ `iss`/`aud`/`iat`/`nbf`/`exp`; set `typ: JWT` trong header (khuyến nghị).
  - (Tùy chọn) Thêm `jti` để hỗ trợ denylist ngắn hạn khi cần thu hồi access token.
- Phân phối public keys:
  - Ưu tiên JWKS endpoint; hoặc thư mục PEM (map `kid` = tên file không phần mở rộng).
- Test cần có: parse PEM, `unknown kid`, nhiều keys song song, `nbf/leeway`, hết hạn, backward-compat khi rotate.

### Rate limiting & auth
- Login: rate limit per-IP và per-account (email); header `Retry-After`, `X-RateLimit-*` rõ ràng.
- Redis limiter: bật fail-closed cho endpoint nhạy cảm bằng env `HTTP_LOGIN_RATELIMIT_FAIL_CLOSED=true`.

### CORS & headers bảo mật
- Ở prod, khóa origin bằng whitelist (`HTTP_CORS_ALLOWED_ORIGINS`); tránh `*`.
- HSTS chỉ khi HTTPS (TLS thực hoặc `X-Forwarded-Proto=https`).

### Observability
- Thêm `/metrics` (Prometheus): HTTP request count/latency, DB pool stats, rate-limit hits.
- Tracing (OpenTelemetry) cho HTTP + DB; pprof (dev-only).

### CI/CD & kiểm thử
- CI: build/test/lint/govulncheck; Docker image build + image scanning (Trivy).
- Lint mở rộng: gosec, gocritic, misspell, depguard, errorlint.
- Test: unit (domain/usecases), HTTP handlers (table-driven), integration (Postgres/Redis) bằng testcontainers/compose.

### Secret management
- `JWT_SECRET`/khóa riêng, `DB_PASSWORD` quản lý qua secret manager; xoay vòng định kỳ.
- Inject qua env/runtime (không commit), audit truy cập secrets.


