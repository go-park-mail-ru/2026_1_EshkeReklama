#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${ROOT}/perf_test/init.sql"

{
  echo "-- $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  for f in "${ROOT}"/db/migrations/*.up.sql; do
    echo "-- $(basename "$f")"
    cat "$f"
    echo
  done
} >"$OUT"

echo "wrote $OUT"
