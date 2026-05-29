#!/usr/bin/env bash
# Собирает perf_test/init.sql из db/migrations/*.up.sql
#
# Запускайте один раз в начале ДЗ (baseline «до оптимизаций»).
#
#   ./perf_test/generate_init_sql.sh -o perf_test/init.sql

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MIGRATIONS="${ROOT}/db/migrations"
OUT="${ROOT}/perf_test/init.sql"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -o|--output)
      OUT="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [-o path/to/init.sql]"
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

if [[ ! -d "$MIGRATIONS" ]]; then
  echo "Migrations dir not found: $MIGRATIONS" >&2
  exit 1
fi

shopt -s nullglob
files=("$MIGRATIONS"/*.up.sql)
shopt -u nullglob

if [[ ${#files[@]} -eq 0 ]]; then
  echo "No *.up.sql files in $MIGRATIONS" >&2
  exit 1
fi

COMMIT="unknown"
if git -C "$ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  COMMIT="$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo unknown)"
fi

mkdir -p "$(dirname "$OUT")"

{
  echo "-- perf_test/init.sql"
  echo "-- Снимок DDL до оптимизаций под нагрузочное тестирование."
  echo "-- Сгенерировано: $(date -u +"%Y-%m-%dT%H:%M:%SZ") (UTC)"
  echo "-- Git commit: ${COMMIT}"
  echo "-- Источник: db/migrations/*.up.sql (по порядку имени файла)"
  echo "--"
  echo "-- Пересборка: ./perf_test/generate_init_sql.sh"
  echo ""
  for f in "${files[@]}"; do
    echo "-- >>> $(basename "$f")"
    cat "$f"
    echo ""
  done
} >"$OUT"

echo "Wrote ${OUT} (${#files[@]} migration files, $(wc -l <"$OUT" | tr -d ' ') lines)"
