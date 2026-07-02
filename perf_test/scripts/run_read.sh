#!/usr/bin/env bash
set -euo pipefail

PERF="$(cd "$(dirname "$0")/.." && pwd)"
ENV="${PERF}/.env"

[[ -f "$ENV" ]] || { echo "missing $ENV" >&2; exit 1; }
source "$ENV"

: "${BASE_URL:?}" "${SESSION_ID:?}" "${READ_MIN_AD_ID:?}" "${READ_MAX_AD_ID:?}"

export READ_MIN_AD_ID READ_MAX_AD_ID

OUT="${PERF}/results/iteration_${1:-1}_read_$(date +%Y%m%d_%H%M%S).txt"
mkdir -p "${PERF}/results"

wrk -t"${WRK_THREADS:-8}" -c"${WRK_CONNECTIONS:-200}" -d"${WRK_DURATION_READ:-60s}" \
  -s "${PERF}/wrk/read.lua" \
  -H "Cookie: session_id=${SESSION_ID}" "$BASE_URL" | tee "$OUT"
