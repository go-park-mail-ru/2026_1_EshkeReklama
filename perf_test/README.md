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
    ```bash
    sudo apt install wrk
    ```

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

## Запуск тестов

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

| Параметр                  | Значение |
|---------------------------|----------|
| Дата                      | |
| URL ВМ                    | |
| Параметры wrk             | `-t8 -c200 -d30m` |
| Requests/sec              | |
| Latency avg / stdev / max | |
| Всего запросов            | |
| Socket errors             | connect , read , write , timeout |
| Объявлений в БД           | |
| Файл                      | `perf_test/results/iteration_N_create_*.txt` |

#### READ

| Параметр | Значение |
|----------|----------|
| Диапазон id | … |
| Параметры wrk | `-t8 -c200 -d1m` |
| Requests/sec | |
| Latency avg / stdev / max | |
| Всего запросов | |
| Socket errors | connect , read , write , timeout |
| Файл | `perf_test/results/iteration_N_read_*.txt` |

#### Анализ

- **CREATE:** узкое место — …; подтверждение — …
- **READ:** узкое место — …; подтверждение — …

**Вывод для итерации:**

#### Оптимизация (если делали)

- Что изменили (индекс, запрос, пул соединений …)
- Почему ожидали эффект

#### Сравнение с итерацией N-1

| Метрика | N-1 | N | Δ |
|---------|-----|---|-----|
| CREATE RPS | | | |
| CREATE latency avg | | | |
| READ RPS | | | |
| READ latency avg | | | |

---

### Итерация 1 (базовая, до оптимизаций)

#### CREATE

| Параметр                  | Значение |
|---------------------------|----------|
| Дата                      | 2026-05-30 |
| URL ВМ                    | `212.233.96.112:8000` |
| Параметры wrk             | `-t8 -c200 -d30m` |
| Requests/sec              | **60.86** |
| Latency avg / stdev / max | 178.84 ms / 99.15 ms / 1.99 s |
| Всего запросов            | 109 560 за 30 min |
| Socket errors             | connect 0, read 957, write 0, **timeout 33 958** |
| Объявлений в БД           | 110 534 |
| Файл                      | `perf_test/results/iteration_1_create_20260530_000303.txt` |

#### READ

| Параметр | Значение |
|----------|----------|
| Диапазон id | 3866 … 114399 |
| Параметры wrk | `-t8 -c200 -d1m` |
| Requests/sec | **132.11** |
| Latency avg / stdev / max | 57.69 ms / 124.57 ms / 1.97 s |
| Всего запросов | 7 939 за 1 min |
| Socket errors | connect 0, read 0, write 0, **timeout 1 527** |
| Файл | `perf_test/results/iteration_1_read_20260530_004125.txt` |

#### Анализ

**CREATE (~61 RPS, avg 179 ms, ~24% timeout)**

- **Узкое место:** насыщение приложения и БД при 200 одновременных соединений, а не один «медленный» SQL.
- На каждый запрос: проверка сессии (Redis), CSRF, разбор multipart, затем в PostgreSQL цепочка `ownedGroup` — `SELECT ad_group` + `SELECT ad_campaign` — и `INSERT INTO ad`. Картинка в нагрузке не передаётся (S3 не участвует).
- **33 958 timeout** при ~110k завершённых запросов — wrk ждёт ответ дольше лимита; очередь растёт быстрее, чем сервер успевает обрабатывать. **957 read errors** — симптом перегруза, а не «плохого» запроса.
- Пул соединений к PostgreSQL в коде не настраивается (`sql.Open` без `SetMaxOpenConns`), при 200 wrk-коннектах возможна конкуренция за соединения к БД.
- **Как подтвердить дальше:** `EXPLAIN ANALYZE` для insert/select group/campaign; метрики PostgreSQL (active connections, wait events); pprof CPU приложения под CREATE-нагрузкой.

**READ (~132 RPS, avg 58 ms, ~16% timeout)**

- **Узкое место:** накладные расходы на каждый запрос + конкуренция при высокой параллельности; сам `SELECT … FROM ad WHERE id = $1` по PK при ~110k строк должен быть быстрым.
- Цепочка: Redis (сессия) → `GetAdvertiserByID` (проверка роли admin в middleware **на каждый запрос**) → `GetByID` объявления по первичному ключу → JSON.
- READ быстрее CREATE (~2× RPS, ~3× меньше latency): меньше SQL на запрос, нет multipart и проверки владения группой/кампанией.
- **1 527 timeout** из 7 939 успевших — та же картина очереди при `-c200`.
- **Как подтвердить дальше:** `EXPLAIN ANALYZE SELECT … FROM eshkere.ad WHERE id = …`; сравнить RPS при `-c50` vs `-c200`; замерить долю времени в auth/IsAdmin vs SQL.

**Вывод для итерации 1:** приоритет — снять очередь (лимит wrk-соединений / пул PG / кэш роли admin), затем точечно оптимизировать лишние запросы на CREATE (`ownedGroup`).

#### Оптимизация

_Не проводилась — 0 итерация._

#### Сравнение с итерацией N-1


| Метрика | N-1 | N | Δ |
|---------|-----|---|-----|
| CREATE RPS | — | 60.86 | — |
| CREATE latency avg | — | 178.84 ms | — |
| READ RPS | — | 132.11 | — |
| READ latency avg | — | 57.69 ms | — |


## Очистка и сброс

```bash
./perf_test/scripts/reset.sh
```

## Примечания по данным

- Заголовки объявлений: `LOADTEST_1`, `LOADTEST_2`, … — валидные строки, удобно чистить.
- `short_desc`, `target_url` — фиксированные допустимые значения (см. `wrk/create.lua`).
- Для CREATE wrk нужны cookie `session_id` и заголовок `X-CSRF-Token` (выдаёт `setup.sh`).
