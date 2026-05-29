#!/usr/bin/env bash
# Подставить в .env READ_MIN_AD_ID / READ_MAX_AD_ID после create-теста.
# Требует psql и переменные PG* или DATABASE_URL.
set -euo pipefail

PREFIX="${LOADTEST_PREFIX:-LOADTEST_}"

if [[ -n "${DATABASE_URL:-}" ]]; then
  ROWS="$(psql "$DATABASE_URL" -tA -c \
    "SELECT min(id), max(id), count(*) FROM eshkere.ad WHERE title LIKE '${PREFIX}%';")"
else
  : "${PGHOST:=localhost}"
  : "${PGDATABASE:=db}"
  ROWS="$(psql -h "$PGHOST" -U "${PGUSER:-postgres}" -d "$PGDATABASE" -tA -c \
    "SELECT min(id), max(id), count(*) FROM eshkere.ad WHERE title LIKE '${PREFIX}%';")"
fi

IFS='|' read -r MIN_ID MAX_ID CNT <<<"$ROWS"
echo "count=${CNT} min_id=${MIN_ID} max_id=${MAX_ID}"
echo "Добавьте в perf_test/.env:"
echo "READ_MIN_AD_ID=${MIN_ID}"
echo "READ_MAX_AD_ID=${MAX_ID}"
