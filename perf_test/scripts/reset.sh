#!/usr/bin/env bash
set -euo pipefail

PERF="$(cd "$(dirname "$0")/.." && pwd)"
ROOT="$(cd "$PERF/.." && pwd)"
ENV="${PERF}/.env"

[[ -f "$ENV" ]] || { echo "missing $ENV" >&2; exit 1; }
source "$ENV"
: "${POSTGRES_USER:?}" "${POSTGRES_DB:?}"

docker compose -f "${ROOT}/docker-compose.yml" exec -T postgres \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f - <"${PERF}/sql/cleanup.sql"

for key in CSRF_TOKEN SESSION_ID LOADTEST_CAMPAIGN LOADTEST_GROUP READ_MIN_AD_ID READ_MAX_AD_ID; do
  grep -q "^${key}=" "$ENV" && sed -i "s|^${key}=.*|${key}=|" "$ENV" || true
done

exec "${PERF}/scripts/setup.sh"
