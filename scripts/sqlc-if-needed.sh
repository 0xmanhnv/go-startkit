#!/usr/bin/env bash
set -euo pipefail
mkdir -p .tmp
HASH_FILE=.tmp/sqlc.hash

# Collect .sql files and sqlc.yaml (if present), compute stable hash
mapfile -d '' FILES < <( \
  { find . -type f -name '*.sql' -print0 2>/dev/null; \
    [[ -f sqlc.yaml ]] && printf 'sqlc.yaml\0'; } \
  | sort -z )

if ((${#FILES[@]}==0)); then
  echo "[sqlc] no SQL/sqlc.yaml found -> skip"
  exit 0
fi

NEW_HASH=$(
  printf '%s\0' "${FILES[@]}" \
  | xargs -0 sha256sum \
  | sha256sum \
  | awk '{print $1}'
)

if [[ ! -f "$HASH_FILE" ]] || [[ "$NEW_HASH" != "$(cat "$HASH_FILE")" ]]; then
  echo "[sqlc] changes detected -> generating..."
  sqlc generate
  echo "$NEW_HASH" > "$HASH_FILE"
else
  echo "[sqlc] no changes -> skip"
fi
