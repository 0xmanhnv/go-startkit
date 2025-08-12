# AppSecHub — Next Updates (Actionable)

- Domain & schema
  - Add tables/migrations: `projects`, `assets`, `findings`, `evidence`, `integrations`, `pipelines`, `sbom_components`, `pentest_runs`.
  - Create sqlc queries for core entities; generate code and repositories.

- Ingestion & normalization
  - Implement parsers: SARIF (SAST) and CycloneDX (SBOM/SCA).
  - Endpoints: `POST /v1/findings/sarif`, `POST /v1/sbom` (validate, size-limit, idempotent, dedup by fingerprint).
  - Persist evidence (later: move large artefacts to S3-compatible storage).

- AuthZ/tenancy
  - Extend RBAC to org/project scopes; add audit logs for critical changes.
  - Plan SSO (OIDC); JWT RS256/EdDSA with key rotation (kid).

- Observability & operations
  - Expose `/metrics` (Prometheus): HTTP latency, DB pool stats, rate-limit counters.
  - Add OpenTelemetry tracing (HTTP + pgx) and pprof (dev-only).
  - Readiness `/readyz`: include Redis ping when used (refresh/limiter).

- Security hardening
  - Rate-limit per-account (email) for `/v1/auth/login` (combine with IP).
  - Lock CORS origins in prod; keep HSTS only on HTTPS.
  - Increase bcrypt cost per env or consider argon2id.
  - Manage secrets via secret manager (JWT keys, DB/Redis tokens).

- DB/indexing & performance
  - Index findings by `(project_id, severity, rule, component, status, updated_at)`.
  - Expose and tune pgxpool via env (max/min conns, lifetime/idle, health check).
  - Add light retries for idempotent reads on transient errors (pgx).

- CI/CD & quality
  - GitHub Actions: build/test/lint/govulncheck; container image scan (Trivy).
  - Lint: add `gosec`, `errorlint`, `depguard`; pin tool versions.
  - Tests: table-driven handler tests (authn/z, CORS, mapper); integration (testcontainers) and parser fixtures.

- API/UI slices (incremental)
  - Projects CRUD; Findings list/filter/export; SBOM list; Integrations CRUD.
  - Webhooks (Slack/Jira); dashboards (severity, trends, MTTR).

- Immediate sprint outline
  - Migrations + sqlc for Projects/Findings/SBOM.
  - Endpoints: `POST /v1/projects`, `GET /v1/findings` with filters; SBOM ingest (CycloneDX), SARIF ingest.
  - Add `/metrics` and tracing middleware; wire Grafana/Prom stack (optional compose).
  - Implement per-account login rate-limit.


