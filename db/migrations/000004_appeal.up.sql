BEGIN;

CREATE TYPE eshkere.appeal_status AS ENUM (
    'open', 'in_progress', 'closed'
);

CREATE TYPE eshkere.appeal_category AS ENUM (
    'bug', 'suggestion', 'complaint', 'question' -- баг, фича, продуктовая жалоба, вопрос по пользованию
);

CREATE TABLE eshkere.appeal (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    advertiser_id   INT                         REFERENCES eshkere.advertiser(id) ON DELETE SET NULL,
    status          eshkere.appeal_status       NOT NULL DEFAULT 'open',
    category        eshkere.appeal_category     NOT NULL,
    title           TEXT                        NOT NULL,
    description     TEXT                        NOT NULL,
    image_url       TEXT,
    name            TEXT                        NOT NULL,
    email           TEXT                        NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_appeal_title_length CHECK (LENGTH(title) <= 100)
);


CREATE TRIGGER set_updated_at_appeal
    BEFORE UPDATE
    ON eshkere.appeal
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMIT;