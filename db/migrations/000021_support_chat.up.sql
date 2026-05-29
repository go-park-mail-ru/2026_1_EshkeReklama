CREATE TYPE eshkere.support_message_author AS ENUM ('advertiser', 'moderator', 'system');

CREATE TABLE eshkere.support_thread (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    campaign_id INT NOT NULL UNIQUE REFERENCES eshkere.ad_campaign(id) ON DELETE CASCADE,
    advertiser_id INT NOT NULL REFERENCES eshkere.advertiser(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE TABLE eshkere.support_message (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    thread_id INT NOT NULL REFERENCES eshkere.support_thread(id) ON DELETE CASCADE,
    author_type eshkere.support_message_author NOT NULL,
    author_id INT REFERENCES eshkere.advertiser(id) ON DELETE SET NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_support_message_thread_id
    ON eshkere.support_message(thread_id, created_at);
