BEGIN;

INSERT INTO eshkere.topic (name) VALUES
    ('Технологии'),
    ('Бизнес'),
    ('Красота и здоровье'),
    ('Авто'),
    ('Недвижимость'),
    ('Еда и рестораны'),
    ('Путешествия'),
    ('Спорт'),
    ('Мода'),
    ('Образование')
    ON CONFLICT (name) DO NOTHING;

COMMIT;