BEGIN;

-- Переименовываем значение enum 'noob' -> 'basic'
ALTER TYPE eshkere.tariff_type RENAME VALUE 'noob' TO 'basic';

-- Добавляем поле для хранения даты окончания подписки
-- NULL = бессрочно (basic и cheater), для pro = дата окончания
ALTER TABLE eshkere.advertiser
    ADD COLUMN IF NOT EXISTS tariff_expires_at TIMESTAMPTZ;

COMMIT;
