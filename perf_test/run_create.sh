#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV="${ROOT}/perf_test/.env"
[[ -f "$ENV" ]] || { echo "missing $ENV" >&2; exit 1; }

# shellcheck disable=SC1090
source "$ENV"

: "${BASE_URL:?}" "${SESSION_ID:?}" "${CSRF_TOKEN:?}" "${LOADTEST_CAMPAIGN:?}" "${LOADTEST_GROUP:?}"

export LOADTEST_CAMPAIGN LOADTEST_GROUP LOADTEST_PREFIX="${LOADTEST_PREFIX:-LOADTEST_}"
export WRK_SESSION_ID="$SESSION_ID" WRK_CSRF_TOKEN="$CSRF_TOKEN"

OUT="${ROOT}/perf_test/results/iteration_${1:-1}_create_$(date +%Y%m%d_%H%M%S).txt"
mkdir -p "${ROOT}/perf_test/results"

wrk -t"${WRK_THREADS:-8}" -c"${WRK_CONNECTIONS:-100}" -d"${WRK_DURATION_CREATE:-30m}" \
  -s "${ROOT}/perf_test/wrk/create.lua" "$BASE_URL" | tee "$OUT"
