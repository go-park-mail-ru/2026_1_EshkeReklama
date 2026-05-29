#!/usr/bin/env bash
# Подставить в .env READ_MIN_AD_ID / READ_MAX_AD_ID после create-теста.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${ROOT}/perf_test/.env"
if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  source "$ENV_FILE"
fi

PREFIX="${LOADTEST_PREFIX:-LOADTEST_}"
PGDATABASE="${PGDATABASE:-db}"
PGUSER="${PGUSER:-postgres}"

SQL="SELECT min(id), max(id), count(*) FROM eshkere.ad WHERE title LIKE '${PREFIX}%';"

run_query() {
  if [[ -n "${DATABASE_URL:-}" ]] && command -v psql >/dev/null 2>&1; then
    psql "$DATABASE_URL" -tA -c "$SQL"
    return
  fi
  if command -v psql >/dev/null 2>&1; then
    psql -h "${PGHOST:-127.0.0.1}" -U "$PGUSER" -d "$PGDATABASE" -tA -c "$SQL"
    return
  fi
  if docker compose -f "${ROOT}/docker-compose.yml" exec -T postgres \
    psql -U "$PGUSER" -d "$PGDATABASE" -tA -c "$SQL" 2>/dev/null; then
    return
  fi
  echo "psql not found. Run manually:" >&2
  echo "  docker compose exec postgres psql -U $PGUSER -d $PGDATABASE -c \"${SQL}\"" >&2
  exit 1
}

ROWS="$(run_query)"

IFS='|' read -r MIN_ID MAX_ID CNT <<<"$ROWS"
echo "count=${CNT} min_id=${MIN_ID} max_id=${MAX_ID}"

if [[ -z "$MIN_ID" || -z "$MAX_ID" || "$CNT" == "0" ]]; then
  echo "Нет объявлений LOADTEST_* — сначала ./perf_test/run_create.sh" >&2
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Нет ${ENV_FILE}" >&2
  exit 1
fi

update_env() {
  local key="$1" val="$2"
  if grep -q "^${key}=" "$ENV_FILE"; then
    if [[ "$(uname)" == "Darwin" ]]; then
      sed -i '' "s|^${key}=.*|${key}=${val}|" "$ENV_FILE"
    else
      sed -i "s|^${key}=.*|${key}=${val}|" "$ENV_FILE"
    fi
  else
    echo "${key}=${val}" >>"$ENV_FILE"
  fi
}

update_env READ_MIN_AD_ID "$MIN_ID"
update_env READ_MAX_AD_ID "$MAX_ID"

echo "Записано в ${ENV_FILE}:"
echo "  READ_MIN_AD_ID=${MIN_ID}"
echo "  READ_MAX_AD_ID=${MAX_ID}"
