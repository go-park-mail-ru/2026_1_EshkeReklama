# Нагрузочное тестирование и оптимизация СУБД

## Основная сущность

**Объявление (`ad`)** — ядро продукта (рекламодатель → кампания → группа → объявление).

| Этап ДЗ | Endpoint | Метод |
|---------|----------|--------|
| Создание 100k | `/api/ad_campaigns/{campaign}/ad_groups/{group}/ads` | `POST` (multipart) |
| Чтение под нагрузкой | `/api/admin/ads/{ad_id}` | `GET` |

## Структура каталога

```
perf_test/
  README.md           — этот файл (отчёты по итерациям ниже)
  init.sql            — DDL до оптимизаций (generate_init_sql.sh)
  generate_init_sql.sh — собрать init.sql из миграций
  env.example         — шаблон переменных
  setup.sh            — логин, кампания, группа
  run_create.sh       — wrk: CREATE
  run_read.sh         — wrk: READ
  discover_ad_ids.sh  — min/max id после сида
  cleanup.sql         — удаление LOADTEST_*
  wrk/create.lua
  wrk/read.lua
  results/            — вывод wrk по итерациям
```

## Подготовка (один раз)

1. **Установка wrk:**

    MacOS:`brew install wrk`

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
   Залогиниться этим пользователем, положить `session_id` в `ADMIN_SESSION_ID` в `.env`.

4. Собрать **`init.sql`** один раз (baseline до оптимизаций):
   ```bash
   ./perf_test/generate_init_sql.sh
   ```
   После старта perf-работы не перезаписывайте без веской причины.

## Запуск тестов (автоматизация)

```bash
# 1) Подготовка сессии и LOADTEST кампании/группы
./perf_test/setup.sh

# 2) Нагрузка на CREATE (крутить, пока в БД не ~100k)
./perf_test/run_create.sh 1
# Проверка:
./perf_test/discover_ad_ids.sh   # или SELECT count(*) ...

# 3) Прописать READ_MIN_AD_ID / READ_MAX_AD_ID в .env

# 4) Нагрузка на READ
./perf_test/run_read.sh 1
```

Логи wrk сохраняются в `perf_test/results/iteration_*_*.txt`.

## Что писать в отчёт (шаблон итерации)

Скопируйте блок на каждую итерацию оптимизации.

### Итерация N

#### CREATE (заполнение ~100k)

- Дата, ВМ, версия коммита
- Параметры wrk: `-t`, `-c`, `-d`
- **Requests/sec**, latency (avg / stdev / max), errors
- Сколько объявлений в БД: `SELECT count(*) ... LIKE 'LOADTEST_%'`
- Файл: `perf_test/results/iteration_N_create_*.txt`

#### READ (после заполнения)

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

## Очистка

```bash
psql ... -f perf_test/cleanup.sql
```

## Примечания по данным

- Заголовки объявлений: `LOADTEST_1`, `LOADTEST_2`, … — валидные строки, удобно чистить.
- `short_desc`, `target_url` — фиксированные допустимые значения (см. `wrk/create.lua`).
- Для CREATE wrk нужны cookie `session_id` и заголовок `X-CSRF-Token` (выдаёт `setup.sh`).

## Цикл по заданию

1. `run_create.sh` → отчёт CREATE  
2. Убедиться в ~100k строк в БД  
3. `run_read.sh` → отчёт READ  
4. Анализ → оптимизация (по желанию)  
5. Повтор с п.3 (`run_read.sh 2`, …)
