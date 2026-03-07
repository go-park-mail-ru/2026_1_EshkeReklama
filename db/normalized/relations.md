## 1 НФ: Все значения атрибутов атомарны
В нашей схеме нет отношений, в которых в каком-либо кортеже одному атрибуту соответствуют несколько значений

## 2 НФ: 1 НФ и нет частичных функциональных зависимостей
В большинстве таблиц используется суррогатный первичный ключ id, поэтому частичная зависимость невозможна
В таблице daily_stat 2НФ выполняется, так как нет атрибутов, которые бы зависели от части ключа

## 3 НФ: 2 НФ и нет транзитивных зависимостей
Мы вынесли из таблиц `ad_campaign`, `ad_group`, `ad`, `partner_site` атрибуты `topic`, `status` и `region`, чтобы избавиться от транзитивных зависимостей

## НФ Бойса-Кодда: 3 НФ и все детерминанты являются потенциальными ключами
В таблицах `advertiser` и `partner` существуют функциональные зависимости от атрибутов `email` и `phone_number`, а в таблице `partner_site` зависимость от атрибута `url`. Так как данные поля являются уникальными, они классифицируются как потенциальные ключи, что полностью удовлетворяет требованиям НФБК

### `advertiser`:

{id} -> name, email, phone_number, password_hash, password_salt, balance, created_at, updated_at

{email} -> id, name, phone_number, password_hash, password_salt, balance, created_at, updated_at

{phone_number} -> id, name, email, password_hash, password_salt, balance, created_at, updated_at

### `ad_campaign`:
{id} -> advertiser_id, status_id, name, daily_budget, created_at, updated_at

### `ad_group`:
{id} -> campaign_id, topic_id, region_id, name, age_from, age_to, gender, created_at, updated_at

### `ad`:
{id} -> group_id, status_id, title, short_desc, target_url, created_at, updated_at

### `partner`:
{id} -> name, email, phone_number, password_hash, password_salt, balance, created_at, updated_at

{email} -> id, name, phone_number, password_hash, password_salt, balance, created_at, updated_at

{phone_number} -> id, name, email, password_hash, password_salt, balance, created_at, updated_at

### `partner_site`:
{id} -> partner_id, topic_id, region_id, age_from, age_to, gender, url, created_at, updated_at

{url} -> id, partner_id, topic_id, region_id, age_from, age_to, gender, created_at, updated_at

### `ad_action`:
{id} -> ad_id, partner_site_id, action, age, region, gender, created_at

### `daily_stat`:
{ad_id, partner_site_id, stat_date} -> displays_count, clicks_count, target_actions_count, advertiser_spend, partner_reward

### `topic`:
{id} -> name

### `status`:
{id} -> name

### `region`:
{id} -> name
