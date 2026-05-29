# Нагрузочное тестирование и оптимизация СУБД

## Основная сущность

**Объявление (`ad`)** — основная сущность продукта (рекламодатель → кампания → группа → объявление).

| Этап ДЗ | Endpoint | Метод |
|---------|----------|--------|
| Создание | `/api/ad_campaigns/{campaign}/ad_groups/{group}/ads` | `POST` (multipart) |
| Чтение под нагрузкой | `/api/admin/ads/{ad_id}` | `GET` |

## Структура каталога

```
perf_test/
  README.md
  env.example
  scripts/
    setup.sh
    reset.sh
    run_create.sh
    run_read.sh
    discover_ad_ids.sh
    generate_init_sql.sh
  sql/
    cleanup.sql
    init.sql
  wrk/
    create.lua
    read.lua
  results/
```

## Подготовка (один раз)

1. **Установка wrk:**

    Ubuntu:`sudo apt install wrk`

2. **Конфиг**:
   ```bash
   cp perf_test/env.example perf_test/.env
   # отредактировать BASE_URL, логин, пароль
   
   source perf_test/.env
   ```

3. **Проверить/выдать роль Admin для read-теста**:
   ```sql
   UPDATE eshkere.advertiser SET role = 'admin' WHERE id = <ваш_id>;
   ```
   
4. Собрать **`init.sql`** один раз до оптимизаций:
   ```bash
   ./perf_test/scripts/generate_init_sql.sh
   ```
   После старта perf-работы не перезаписывайте без веской причины.

## Запуск тестов (автоматизация)

```bash
# 1) Подготовка сессии и LOADTEST кампании/группы
./perf_test/scripts/setup.sh

# 2) Нагрузка на CREATE (крутить, пока в БД не ~100k)
./perf_test/scripts/run_create.sh <N>  # N -- номер оптимизации

# 3) Обязательная проверка и сохранение крайних id для последующего удаления:
./perf_test/scripts/discover_ad_ids.sh

# 4) Нагрузка на READ
./perf_test/scripts/run_read.sh <N>  # N -- номер оптимизации
```

Логи wrk сохраняются в `perf_test/results/iteration_*_*.txt`.

## Шаблон отчёта

### Итерация N

#### CREATE

- Дата, ВМ, версия коммита
- Параметры wrk: `-t`, `-c`, `-d`
- **Requests/sec**, latency (avg / stdev / max), errors
- Сколько объявлений в БД: `SELECT count(*) ... LIKE 'LOADTEST_%'`
- Файл: `perf_test/results/iteration_N_create_*.txt`

#### READ

- Диапазон id: `READ_MIN_AD_ID` … `READ_MAX_AD_ID`
- **Requests/sec**, latency, errors
- Файл: `perf_test/results/iteration_N_read_*.txt`

#### Анализ

- Узкое место (CPU / PostgreSQL / сеть / auth / S3 …)
- Как подтвердили (логи, `EXPLAIN ANALYZE`, метрики Grafana, pprof)

#### Оптимизация (если делали)

- Что изменили (индекс, запрос, пул соединений …)
- Почему ожидали эффект

#### Сравнение с итерацией N-1

| Метрика | N-1 | N | Δ |
|---------|-----|---|-----|
| READ RPS | | | |
| READ p99 | | | |

---

### Итерация 0 (базовая, до оптимизаций)

_Заполните после первого прогона._

## Очистка и сброс

```bash
./perf_test/scripts/reset.sh
```

## Примечания по данным

- Заголовки объявлений: `LOADTEST_1`, `LOADTEST_2`, … — валидные строки, удобно чистить.
- `short_desc`, `target_url` — фиксированные допустимые значения (см. `wrk/create.lua`).
- Для CREATE wrk нужны cookie `session_id` и заголовок `X-CSRF-Token` (выдаёт `setup.sh`).
