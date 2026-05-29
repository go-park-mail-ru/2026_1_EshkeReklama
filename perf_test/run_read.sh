#!/usr/bin/env bash
# Итерация: нагрузочное тестирование READ (wrk).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${ROOT}/perf_test/.env"
ITERATION="${1:-1}"

# shellcheck disable=SC1090
source "$ENV_FILE"

: "${BASE_URL:?}"
: "${ADMIN_SESSION_ID:?set ADMIN_SESSION_ID (admin session)}"
: "${READ_MIN_AD_ID:?}"
: "${READ_MAX_AD_ID:?}"

WRK_THREADS="${WRK_THREADS:-8}"
WRK_CONNECTIONS="${WRK_CONNECTIONS:-200}"
WRK_DURATION_READ="${WRK_DURATION_READ:-60s}"

export READ_MIN_AD_ID READ_MAX_AD_ID

if ! command -v wrk >/dev/null; then
  echo "Установите wrk: brew install wrk" >&2
  exit 1
fi

RESULT_DIR="${ROOT}/perf_test/results"
mkdir -p "$RESULT_DIR"
OUT="${RESULT_DIR}/iteration_${ITERATION}_read_$(date +%Y%m%d_%H%M%S).txt"

echo "=== READ load test iteration ${ITERATION} ===" | tee "$OUT"
echo "GET ${BASE_URL}/api/admin/ads/{id}  ids=${READ_MIN_AD_ID}..${READ_MAX_AD_ID}" | tee -a "$OUT"
echo "" | tee -a "$OUT"

wrk -t"$WRK_THREADS" -c"$WRK_CONNECTIONS" -d"$WRK_DURATION_READ" \
  -s "${ROOT}/perf_test/wrk/read.lua" \
  -H "Cookie: session_id=${ADMIN_SESSION_ID}" \
  "${BASE_URL}" 2>&1 | tee -a "$OUT"

echo "" | tee -a "$OUT"
echo "Результат: $OUT" | tee -a "$OUT"
