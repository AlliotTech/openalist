#!/usr/bin/env bash

set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
web_dir="${OPENALIST_WEB_DIR:-$(dirname "$project_dir")/openalist-web}"
data_dir="${OPENALIST_DATA_DIR:-$project_dir/data-dev}"
backend_pid=""

cleanup() {
  if [[ -n "$backend_pid" ]] && kill -0 "$backend_pid" 2>/dev/null; then
    kill "$backend_pid" 2>/dev/null || true
    wait "$backend_pid" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required to start the backend." >&2
  exit 1
fi

if ! command -v pnpm >/dev/null 2>&1; then
  echo "pnpm is required to start the frontend." >&2
  exit 1
fi

if [[ ! -f "$web_dir/package.json" ]]; then
  echo "Frontend repository not found at: $web_dir" >&2
  echo "Set OPENALIST_WEB_DIR to its location and try again." >&2
  exit 1
fi

if [[ ! -d "$web_dir/node_modules" ]]; then
  echo "Frontend dependencies are missing. Run:" >&2
  echo "  cd $web_dir && pnpm install --frozen-lockfile" >&2
  exit 1
fi

echo "Starting backend at http://localhost:5244"
echo "Starting frontend development server (Vite will print its URL)"

(
  cd "$project_dir"
  exec go run . server --dev --data "$data_dir"
) &
backend_pid=$!

cd "$web_dir"
pnpm dev
