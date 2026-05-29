#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV="${ROOT}/perf_test/.env"

[[ -f "$ENV" ]] || { echo "missing $ENV" >&2; exit 1; }
source "$ENV"
: "${BASE_URL:?}" "${ADVERTISER_EMAIL:?}" "${ADVERTISER_PASSWORD:?}"

set_env() {
  grep -q "^$1=" "$ENV" && sed -i "s|^$1=.*|$1=$2|" "$ENV" || echo "$1=$2" >>"$ENV"
}

J=$(mktemp)
trap 'rm -f "$J"' EXIT

curl -sS -c "$J" -b "$J" "${BASE_URL}/api/ad_campaigns" -o /dev/null
CSRF=$(awk -F'\t' '$6=="csrf_token"{print $7}' "$J" | tail -1)
[[ -n "$CSRF" ]] || { echo "csrf_token not received" >&2; exit 1; }

curl -sS -c "$J" -b "$J" -X POST "${BASE_URL}/api/advertisers/login" \
  -H "Content-Type: application/json" -H "X-CSRF-Token: $CSRF" \
  -d "{\"identifier\":\"$ADVERTISER_EMAIL\",\"password\":\"$ADVERTISER_PASSWORD\"}" -o /dev/null

SESSION_ID=$(awk -F'\t' '$6=="session_id"{print $7}' "$J" | tail -1)
[[ -n "$SESSION_ID" ]] || { echo "login failed" >&2; exit 1; }

if [[ -z "${LOADTEST_CAMPAIGN:-}" ]]; then
  R=$(curl -sS -c "$J" -b "$J" -X POST "${BASE_URL}/api/ad_campaigns" \
    -H "Content-Type: application/json" -H "X-CSRF-Token: $CSRF" \
    -d '{"name":"LOADTEST_campaign","daily_budget":1000000,"cpm_price":1000,"main_action":"look"}')
  LOADTEST_CAMPAIGN=$(echo "$R" | sed -E 's/.*"id"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/')
  [[ "$LOADTEST_CAMPAIGN" =~ ^[0-9]+$ ]] || { echo "create campaign failed: $R" >&2; exit 1; }
fi

if [[ -z "${LOADTEST_GROUP:-}" ]]; then
  R=$(curl -sS -c "$J" -b "$J" -X POST "${BASE_URL}/api/ad_campaigns/${LOADTEST_CAMPAIGN}/ad_groups" \
    -H "Content-Type: application/json" -H "X-CSRF-Token: $CSRF" \
    -d '{"topic":"Технологии","region":"Москва","name":"LOADTEST_group","age_from":18,"age_to":65,"gender":"any"}')
  LOADTEST_GROUP=$(echo "$R" | sed -E 's/.*"id"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/')
  [[ "$LOADTEST_GROUP" =~ ^[0-9]+$ ]] || { echo "create group failed: $R" >&2; exit 1; }
fi

set_env CSRF_TOKEN "$CSRF"
set_env SESSION_ID "$SESSION_ID"
set_env LOADTEST_CAMPAIGN "$LOADTEST_CAMPAIGN"
set_env LOADTEST_GROUP "$LOADTEST_GROUP"

echo "OK campaign=$LOADTEST_CAMPAIGN group=$LOADTEST_GROUP"
