BEGIN;

-- 1. Наполнение базовых справочников
INSERT INTO eshkere.topic (name) VALUES ('Авто'), ('IT и Технологии'), ('Красота и здоровье'), ('Образование');

INSERT INTO eshkere.region (name) VALUES ('Москва'), ('Санкт-Петербург'), ('Новосибирск'), ('Казань');

-- 2. Рекламодатели (Advertisers)
-- Баланс хранится в копейках (BIGINT)
INSERT INTO eshkere.advertiser (name, balance)
VALUES
    ('ООО Ромашка', 5000000),
    ('ИП ТехноМир', 150050);

-- 2.1. Учетные данные рекламодателей в auth-service
INSERT INTO auth.credentials (id, email, phone, password_hash)
VALUES
    (1, 'info@romashka.ru', '9001112233', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    (2, 'ads@techno.io', '9998887766', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy');

-- 3. Рекламные кампании (Ad Campaigns)
-- Бюджет хранится в копейках (BIGINT)
INSERT INTO eshkere.ad_campaign (advertiser_id, name, daily_budget, status)
VALUES
    (1, 'Распродажа тюльпанов', 100000, 'working'),
    (1, 'Осенняя акция', 50000, 'turned_off'),
    (2, 'Продвижение курсов Go', 200000, 'working');

-- 4. Группы объявлений с таргетингом (Ad Groups)
-- Предполагаем: 1 - any, 2 - Авто, 3 - IT, 4 - Красота / 1 - any, 2 - Мск
INSERT INTO eshkere.ad_group (ad_campaign_id, topic_id, region_id, name, age_from, age_to, gender)
VALUES
    (1, 4, 2, 'Женщины Москва 25-45', 25, 45, 'woman'),
    (3, 3, 1, 'Разработчики РФ 18-99', 18, 99, 'any');

-- 5. Сами объявления (Ads)
INSERT INTO eshkere.ad (ad_group_id, status, title, short_desc, image_url, target_url)
VALUES
    (1, 'working', 'Скидка на букеты 30%', 'Только до конца недели тюльпаны по спеццене!', 'https://s3.amazonaws.com/my-ads/banners/flower_sale_600x600.png', 'https://flowers.ru/sale'),
    (1, 'moderation', 'Розы из Голландии', 'Свежий завоз, бесплатная доставка.', 'https://s3.amazonaws.com/my-ads/banners/roses_new.jpg', 'https://flowers.ru/roses'),
    (2, 'working', 'Стань Go-разработчиком', 'Интенсивный курс с трудоустройством.', 'https://s3.amazonaws.com/my-ads/banners/go_course_main.png', 'https://techno.io/go-course');

-- 6. Партнеры и их площадки (Partners & Sites)
INSERT INTO eshkere.partner (name, email, phone_number, password_hash, password_salt, balance)
VALUES ('Сергей Владелец Блога', 'sergey@blog.ru', '9115554422', 'hash789', 'salt789', 0);

INSERT INTO eshkere.partner_site (partner_id, topic_id, region_id, age_from, age_to, gender, url)
VALUES (1, 3, 1, 18, 60, 'any', 'https://it-news-portal.ru');

-- 7. Действия (Логи показов и кликов)
-- Используем существующие ID (ad_id=1, site_id=1, region_id=2)
INSERT INTO eshkere.ad_action (ad_id, partner_site_id, region_id, action, age, gender)
VALUES
    (1, 1, 2, 'look', 30, 'woman'),
    (1, 1, 2, 'look', 28, 'woman'),
    (1, 1, 2, 'click', 28, 'woman'),
    (3, 1, 1, 'look', 22, 'man');

COMMIT;
