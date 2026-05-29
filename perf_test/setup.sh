#!/usr/bin/env bash
# Одноразовая подготовка: CSRF, логин, кампания и группа для LOADTEST.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${ROOT}/perf_test/.env"
EXAMPLE="${ROOT}/perf_test/env.example"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Создайте perf_test/.env из env.example" >&2
  cp "$EXAMPLE" "$ENV_FILE"
  echo "Отредактируйте $ENV_FILE и запустите снова." >&2
  exit 1
fi

# shellcheck disable=SC1090
source "$ENV_FILE"

: "${BASE_URL:?set BASE_URL in .env}"
: "${ADVERTISER_EMAIL:?}"
: "${ADVERTISER_PASSWORD:?}"

cookie_value() {
  awk -F'\t' -v name="$2" '$6 == name { print $7 }' "$1" | tail -1
}

COOKIE_JAR="$(mktemp)"
trap 'rm -f "$COOKIE_JAR"' EXIT

curl -sS -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
  "${BASE_URL}/api/ad_campaigns" -o /dev/null -w "%{http_code}" >"${COOKIE_JAR}.code" || true
HTTP_CODE="$(cat "${COOKIE_JAR}.code")"
rm -f "${COOKIE_JAR}.code"

CSRF="$(cookie_value "$COOKIE_JAR" csrf_token)"
if [[ -z "$CSRF" ]]; then
  echo "csrf_token not received (GET ${BASE_URL}/api/ad_campaigns → HTTP ${HTTP_CODE:-?})" >&2
  echo "Проверьте BASE_URL в perf_test/.env и что API отвечает:" >&2
  echo "  curl -v \"\${BASE_URL}/healthz\"" >&2
  if [[ -s "$COOKIE_JAR" ]]; then
    echo "Cookie jar:" >&2
    grep -v '^#' "$COOKIE_JAR" >&2 || true
  fi
  exit 1
fi

LOGIN_RESP="$(curl -sS -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
  -X POST "${BASE_URL}/api/advertisers/login" \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: ${CSRF}" \
  -d "{\"identifier\":\"${ADVERTISER_EMAIL}\",\"password\":\"${ADVERTISER_PASSWORD}\"}")"

SESSION_ID="$(cookie_value "$COOKIE_JAR" session_id)"
if [[ -z "$SESSION_ID" ]]; then
  echo "login failed: ${LOGIN_RESP}" >&2
  exit 1
fi

if [[ -z "${CAMPAIGN_ID:-}" ]]; then
  CAMP_RESP="$(curl -sS -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
    -X POST "${BASE_URL}/api/ad_campaigns" \
    -H "Content-Type: application/json" \
    -H "X-CSRF-Token: ${CSRF}" \
    -d '{"name":"LOADTEST_campaign","daily_budget":1000000,"cpm_price":1000,"main_action":"look"}')"
  CAMPAIGN_ID="$(python3 -c "import json,sys; print(json.load(sys.stdin)['id'])" <<<"$CAMP_RESP")"
fi

if [[ -z "${GROUP_ID:-}" ]]; then
  GROUP_RESP="$(curl -sS -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
    -X POST "${BASE_URL}/api/ad_campaigns/${CAMPAIGN_ID}/ad_groups" \
    -H "Content-Type: application/json" \
    -H "X-CSRF-Token: ${CSRF}" \
    -d '{"topic":"IT и Технологии","region":"Москва","name":"LOADTEST_group","age_from":18,"age_to":65,"gender":"any"}')"
  GROUP_ID="$(python3 -c "import json,sys; print(json.load(sys.stdin)['id'])" <<<"$GROUP_RESP")"
fi

# Обновляем .env (macOS/Linux compatible)
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

update_env CSRF_TOKEN "$CSRF"
update_env SESSION_ID "$SESSION_ID"
update_env CAMPAIGN_ID "$CAMPAIGN_ID"
update_env GROUP_ID "$GROUP_ID"
update_env LOADTEST_CAMPAIGN "$CAMPAIGN_ID"
update_env LOADTEST_GROUP "$GROUP_ID"

echo "OK: session_id=... campaign=${CAMPAIGN_ID} group=${GROUP_ID}"
echo "Для read-теста: выдайте role=admin и укажите ADMIN_SESSION_ID в .env"
echo "  UPDATE eshkere.advertiser SET role = 'admin' WHERE id = <your_id>;"
