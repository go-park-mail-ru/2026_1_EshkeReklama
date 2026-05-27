BEGIN;

INSERT INTO eshkere.region (name) VALUES
    ('Москва'),
    ('Санкт-Петербург'),
    ('Казань'),
    ('Екатеринбург'),
    ('Новосибирск'),
    ('Краснодар'),
    ('Нижний Новгород'),
    ('Самара'),
    ('Ростов-на-Дону')
    ON CONFLICT (name) DO NOTHING;

COMMIT;