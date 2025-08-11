#!/usr/bin/env bash

set -euo pipefail

COMPOSE_FILE="docker-compose.test.yml"
DB_HOST_DEFAULT="localhost"
DB_PORT_DEFAULT="55432"
DB_USER_DEFAULT="gostartkit"
DB_PASSWORD_DEFAULT="devpassword"
DB_NAME_DEFAULT="gostartkit"
REDIS_ADDR_DEFAULT="localhost:56379"

KEEP_CONTAINERS=false

usage() {
  echo "Usage: $0 [--keep] [--] [go test args...]" >&2
  echo "Examples:" >&2
  echo "  $0" >&2
  echo "  $0 -- -run TestPostgres_UserRepository_CRUD -v" >&2
  echo "  $0 --keep" >&2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --keep)
      KEEP_CONTAINERS=true
      shift
      ;;
    -h|--help)
      usage; exit 0 ;;
    --)
      shift; break ;;
    *)
      break ;;
  esac
done

echo "[it] Starting test services via $COMPOSE_FILE..."
if ! docker compose -f "$COMPOSE_FILE" up -d --wait 2>/dev/null; then
  echo "[it] 'docker compose --wait' not supported, falling back to up -d + sleep"
  docker compose -f "$COMPOSE_FILE" up -d
  sleep 5
fi

export DB_HOST="${DB_HOST:-$DB_HOST_DEFAULT}"
export DB_PORT="${DB_PORT:-$DB_PORT_DEFAULT}"
export DB_USER="${DB_USER:-$DB_USER_DEFAULT}"
export DB_PASSWORD="${DB_PASSWORD:-$DB_PASSWORD_DEFAULT}"
export DB_NAME="${DB_NAME:-$DB_NAME_DEFAULT}"
export REDIS_ADDR="${REDIS_ADDR:-$REDIS_ADDR_DEFAULT}"

echo "[it] Running integration tests with env: DB=$DB_HOST:$DB_PORT, REDIS=$REDIS_ADDR"
go test -tags=integration ./internal/tests/integration "$@"

EXIT_CODE=$?

if [[ "$KEEP_CONTAINERS" != true ]]; then
  echo "[it] Stopping test services..."
  docker compose -f "$COMPOSE_FILE" down -v
else
  echo "[it] Keeping test services running as requested (use 'docker compose -f $COMPOSE_FILE down -v' to stop)."
fi

exit $EXIT_CODE


