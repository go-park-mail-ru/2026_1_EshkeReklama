#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV="${ROOT}/perf_test/.env"
[[ -f "$ENV" ]] || { echo "missing $ENV" >&2; exit 1; }

# shellcheck disable=SC1090
source "$ENV"

: "${BASE_URL:?}" "${ADMIN_SESSION_ID:?}" "${READ_MIN_AD_ID:?}" "${READ_MAX_AD_ID:?}"

export READ_MIN_AD_ID READ_MAX_AD_ID

OUT="${ROOT}/perf_test/results/iteration_${1:-1}_read_$(date +%Y%m%d_%H%M%S).txt"
mkdir -p "${ROOT}/perf_test/results"

wrk -t"${WRK_THREADS:-8}" -c"${WRK_CONNECTIONS:-200}" -d"${WRK_DURATION_READ:-60s}" \
  -s "${ROOT}/perf_test/wrk/read.lua" \
  -H "Cookie: session_id=${ADMIN_SESSION_ID}" "$BASE_URL" | tee "$OUT"
