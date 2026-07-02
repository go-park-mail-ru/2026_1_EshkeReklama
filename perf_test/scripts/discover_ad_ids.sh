#!/usr/bin/env bash
set -euo pipefail

PERF="$(cd "$(dirname "$0")/.." && pwd)"
ROOT="$(cd "$PERF/.." && pwd)"
ENV="${PERF}/.env"

[[ -f "$ENV" ]] || { echo "missing $ENV" >&2; exit 1; }
source "$ENV"
: "${POSTGRES_USER:?}" "${POSTGRES_DB:?}"

P="${LOADTEST_PREFIX:-LOADTEST_}"
ROW=$(docker compose -f "${ROOT}/docker-compose.yml" exec -T postgres \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tA \
  -c "SELECT min(id), max(id), count(*) FROM eshkere.ad WHERE title LIKE '${P}%';")

IFS='|' read -r MIN MAX CNT <<<"$ROW"
[[ -n "$MIN" && "$CNT" != "0" ]] || { echo "no LOADTEST ads" >&2; exit 1; }

grep -q '^READ_MIN_AD_ID=' "$ENV" && sed -i "s|^READ_MIN_AD_ID=.*|READ_MIN_AD_ID=$MIN|" "$ENV" || echo "READ_MIN_AD_ID=$MIN" >>"$ENV"
grep -q '^READ_MAX_AD_ID=' "$ENV" && sed -i "s|^READ_MAX_AD_ID=.*|READ_MAX_AD_ID=$MAX|" "$ENV" || echo "READ_MAX_AD_ID=$MAX" >>"$ENV"

echo "count=$CNT READ_MIN_AD_ID=$MIN READ_MAX_AD_ID=$MAX"
