#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "Usage: $0 <new_module_path> [old_module_path]" >&2
  echo "Examples:" >&2
  echo "  $0 github.com/you/appsechub" >&2
  echo "  $0 github.com/you/appsechub appsechub" >&2
}

if [[ ! -f go.mod ]]; then
  echo "go.mod not found in current directory" >&2
  exit 1
fi

if [[ $# -eq 1 ]]; then
  NEW_MODULE="$1"
  # infer OLD from go.mod
  OLD_MODULE=$(awk '/^module /{print $2}' go.mod)
elif [[ $# -eq 2 ]]; then
  NEW_MODULE="$1"
  OLD_MODULE="$2"
else
  usage
  exit 1
fi

if [[ -z "${NEW_MODULE}" ]]; then
  echo "new module path must not be empty" >&2
  exit 1
fi

echo "Renaming module: ${OLD_MODULE} -> ${NEW_MODULE}"

# Update go.mod module path
go mod edit -module "${NEW_MODULE}"

# Replace imports in .go files (avoid dependency on ripgrep)
match_pattern="\"${OLD_MODULE}/"
mapfile -t files < <(grep -RIl --include='*.go' -- "${match_pattern}" . || true)
if [[ ${#files[@]} -gt 0 ]]; then
  for f in "${files[@]}"; do
    sed -i.bak "s#\"${OLD_MODULE}/#\"${NEW_MODULE}/#g" "$f"
    rm -f "$f.bak"
  done
fi

# Tidy and verify build
go mod tidy
go build ./...

echo "Done. Consider also updating image names in docker-compose*.yml and OpenAPI title if desired."