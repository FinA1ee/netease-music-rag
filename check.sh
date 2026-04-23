#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "[1/2] Running backend checks..."
(
  cd "$ROOT_DIR/backend"
  go test ./...
)

echo "[2/2] Running frontend build..."
(
  cd "$ROOT_DIR/frontend"
  npm run build
)

echo "All checks passed."
