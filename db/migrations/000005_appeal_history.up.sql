BEGIN;

CREATE TYPE eshkere.appeal_message_author AS ENUM (
    'admin', 'user'
);

CREATE TABLE eshkere.appeal_message (
    id          INT                             PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    appeal_id   INT                             NOT NULL REFERENCES eshkere.appeal(id) ON DELETE CASCADE,
    author      eshkere.appeal_message_author   NOT NULL,
    text        TEXT                            NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE        NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_appeal_message_appeal_id_created_at
    ON eshkere.appeal_message(appeal_id, created_at);

CREATE TABLE eshkere.appeal_status_history (
    id          INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    appeal_id   INT                         NOT NULL REFERENCES eshkere.appeal(id) ON DELETE CASCADE,
    status      eshkere.appeal_status       NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_appeal_status_history_appeal_id_created_at
    ON eshkere.appeal_status_history(appeal_id, created_at);

COMMIT;

