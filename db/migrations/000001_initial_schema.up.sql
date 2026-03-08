BEGIN;

CREATE SCHEMA IF NOT EXISTS eshekere;

CREATE TYPE eshekere.action_type AS ENUM (
    'look', 'click'
    );

CREATE TYPE eshekere.gender_type AS ENUM (
    'man', 'woman', 'any'
    );

CREATE TYPE eshekere.status_type AS ENUM (
    'turned_off', 'moderation', 'working', 'rejected', 'not_enough_money'
    );

CREATE TABLE eshekere.topic (
    id      INT     PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name    TEXT    NOT NULL UNIQUE
);

CREATE TABLE eshekere.region (
    id      INT     PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name    TEXT    NOT NULL UNIQUE
);

INSERT INTO eshekere.topic (name) VALUES ('any');
INSERT INTO eshekere.region (name) VALUES ('any');

CREATE TABLE eshekere.advertiser (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name            TEXT                        NOT NULL,
    email           TEXT                        NOT NULL UNIQUE,
    phone_number    TEXT                        NOT NULL UNIQUE,
    password_hash   TEXT                        NOT NULL,
    password_salt   TEXT                        NOT NULL,
    balance         DECIMAL(15, 2)              NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_advertiser_name_length CHECK (LENGTH(name) <= 255),
    CONSTRAINT check_advertiser_phone_number_length CHECK (LENGTH(phone_number) = 10),
    CONSTRAINT check_advertiser_balance_positive CHECK (balance >= 0)
);

CREATE TABLE eshekere.ad_campaign (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    advertiser_id   INT                         REFERENCES eshekere.advertiser(id) ON DELETE CASCADE,
    status          eshekere.status_type                 NOT NULL DEFAULT 'turned_off',
    name            TEXT                        NOT NULL,
    daily_budget    DECIMAL(15, 2)              NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_ad_campaign_name_length CHECK (LENGTH(name) <= 255),
    CONSTRAINT check_ad_campaign_daily_budget_positive CHECK (daily_budget >= 0)
);

CREATE TABLE eshekere.ad_group (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    ad_campaign_id  INT                         REFERENCES eshekere.ad_campaign(id) ON DELETE CASCADE,
    topic_id        INT                         DEFAULT 1 REFERENCES eshekere.topic(id) ON DELETE SET DEFAULT,
    region_id       INT                         DEFAULT 1 REFERENCES eshekere.region(id) ON DELETE SET DEFAULT,
    name            TEXT                        NOT NULL,
    age_from        INT                         NOT NULL,
    age_to          INT                         NOT NULL,
    gender          eshekere.gender_type                 NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_ad_group_name_length CHECK (LENGTH(name) <= 255),
    CONSTRAINT check_ad_group_age CHECK (age_from <= age_to AND age_from BETWEEN 0 AND 150 AND age_to BETWEEN 0 AND 150)
);

CREATE TABLE eshekere.ad (
    id          INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    ad_group_id INT                         REFERENCES eshekere.ad_group(id) ON DELETE CASCADE,
    status      eshekere.status_type                 NOT NULL DEFAULT 'turned_off',
    title       TEXT                        NOT NULL,
    short_desc  TEXT                        NOT NULL,
    target_url  TEXT                        NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_ad_title_length CHECK (LENGTH(title) <= 60),
    CONSTRAINT check_ad_short_desc_length CHECK (LENGTH(short_desc) <= 255)
);

CREATE TABLE eshekere.partner (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name            TEXT                        NOT NULL,
    email           TEXT                        NOT NULL UNIQUE,
    phone_number    TEXT                        NOT NULL UNIQUE,
    password_hash   TEXT                        NOT NULL,
    password_salt   TEXT                        NOT NULL,
    balance         DECIMAL(15, 2)              NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_partner_name_length CHECK (LENGTH(name) <= 255),
    CONSTRAINT check_partner_phone_number_length CHECK (LENGTH(phone_number) = 10),
    CONSTRAINT check_partner_balance_positive CHECK (balance >= 0)
);

CREATE TABLE eshekere.partner_site (
    id          INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    partner_id  INT                         REFERENCES eshekere.partner(id) ON DELETE CASCADE,
    topic_id    INT                         DEFAULT 1 REFERENCES eshekere.topic(id) ON DELETE SET DEFAULT,
    region_id   INT                         DEFAULT 1 REFERENCES eshekere.region(id) ON DELETE SET DEFAULT,
    age_from    INT                         NOT NULL,
    age_to      INT                         NOT NULL,
    gender      eshekere.gender_type                 NOT NULL DEFAULT 'any',
    url         TEXT                        NOT NULL UNIQUE,
    created_at  TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_partner_site_age CHECK (age_from <= age_to AND age_from BETWEEN 0 AND 150 AND age_to BETWEEN 0 AND 150)
);

CREATE TABLE eshekere.ad_action (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    ad_id           INT                         REFERENCES eshekere.ad(id) ON DELETE CASCADE,
    partner_site_id INT                         REFERENCES eshekere.partner_site(id) ON DELETE CASCADE,
    region_id       INT                         REFERENCES eshekere.region(id) ON DELETE CASCADE,
    action          eshekere.action_type                 NOT NULL DEFAULT 'look',
    age             INT                         NOT NULL DEFAULT 25,
    gender          eshekere.gender_type                 NOT NULL DEFAULT 'any',
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    CONSTRAINT check_ad_action_age CHECK (age BETWEEN 0 AND 150)
);

-- Создание функции обновления атрибута обновления
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Настройка триггеров обновления атрибута обновления
CREATE TRIGGER set_updated_at_advertiser
    BEFORE UPDATE
    ON eshekere.advertiser
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_ad_campaign
    BEFORE UPDATE
    ON eshekere.ad_campaign
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_ad_group
    BEFORE UPDATE
    ON eshekere.ad_group
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_ad
    BEFORE UPDATE
    ON eshekere.ad
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_partner
    BEFORE UPDATE
    ON eshekere.partner
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_partner_site
    BEFORE UPDATE
    ON eshekere.partner_site
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


COMMIT;