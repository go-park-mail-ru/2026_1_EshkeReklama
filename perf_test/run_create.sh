#!/usr/bin/env bash
# Итерация: нагрузочное тестирование CREATE (wrk) + накопление ~100k объявлений.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${ROOT}/perf_test/.env"
ITERATION="${1:-1}"

# shellcheck disable=SC1090
source "$ENV_FILE"

: "${BASE_URL:?}"
: "${SESSION_ID:?}"
: "${CSRF_TOKEN:?}"
: "${LOADTEST_CAMPAIGN:?}"
: "${LOADTEST_GROUP:?}"

WRK_THREADS="${WRK_THREADS:-8}"
WRK_CONNECTIONS="${WRK_CONNECTIONS:-100}"
WRK_DURATION_CREATE="${WRK_DURATION_CREATE:-30m}"
LOADTEST_PREFIX="${LOADTEST_PREFIX:-LOADTEST_}"

export LOADTEST_CAMPAIGN LOADTEST_GROUP LOADTEST_PREFIX
export WRK_SESSION_ID="$SESSION_ID"
export WRK_CSRF_TOKEN="$CSRF_TOKEN"

if ! command -v wrk >/dev/null; then
  echo "Установите wrk: brew install wrk" >&2
  exit 1
fi

RESULT_DIR="${ROOT}/perf_test/results"
mkdir -p "$RESULT_DIR"
OUT="${RESULT_DIR}/iteration_${ITERATION}_create_$(date +%Y%m%d_%H%M%S).txt"

echo "=== CREATE load test iteration ${ITERATION} ===" | tee "$OUT"
echo "POST ${BASE_URL}/api/ad_campaigns/${LOADTEST_CAMPAIGN}/ad_groups/${LOADTEST_GROUP}/ads" | tee -a "$OUT"
echo "wrk -t${WRK_THREADS} -c${WRK_CONNECTIONS} -d${WRK_DURATION_CREATE}" | tee -a "$OUT"
echo "" | tee -a "$OUT"

wrk -t"$WRK_THREADS" -c"$WRK_CONNECTIONS" -d"$WRK_DURATION_CREATE" \
  -s "${ROOT}/perf_test/wrk/create.lua" \
  "${BASE_URL}" 2>&1 | tee -a "$OUT"

echo "" | tee -a "$OUT"
echo "Проверьте количество в БД:" | tee -a "$OUT"
echo "  SELECT count(*) FROM eshkere.ad WHERE title LIKE '${LOADTEST_PREFIX}%';" | tee -a "$OUT"
echo "Результат: $OUT" | tee -a "$OUT"
