# Contributing to Go Startkit

Thanks for your interest in contributing! We welcome issues, docs, and code.

## How to contribute
- Fork the repo and create a feature branch from `main`.
- Keep edits small and focused. Prefer multiple small PRs over one large PR.
- Add tests for new logic (unit or integration where applicable).
- Ensure `go test ./...` and `golangci-lint run` are green locally.
- Update `README.md`/`CHANGELOG.md` when changing behavior or flags.

## Development setup
- Go 1.24+
- Docker & Docker Compose v2
- Run dev stack with hot-reload:
  - `docker compose -f docker-compose.dev.yml up`

## Commit messages
- Use conventional commits style where possible: `feat: ...`, `fix: ...`, `docs: ...`, `refactor: ...`, `test: ...`, `ci: ...`, `chore: ...`.

## Pull Request checklist
- Code builds and tests pass: `go test -race -cover ./...`
- Lint passes: `golangci-lint run`
- New dependencies justified and minimal
- Added/updated tests and docs

## Coding guidelines
- Follow project architecture: `interfaces/http → application → domain`; infra implements ports.
- Keep public APIs explicit and types clear.
- Avoid deep nesting; use guard clauses.
- Do not log secrets/PII.

## License
By contributing, you agree your contributions will be licensed under the MIT License.

