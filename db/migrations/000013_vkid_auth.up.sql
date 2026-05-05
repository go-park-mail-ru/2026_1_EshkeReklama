ALTER TABLE auth.credentials
    ALTER COLUMN email DROP NOT NULL,
    ALTER COLUMN phone DROP NOT NULL,
    ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE auth.credentials
    ADD COLUMN IF NOT EXISTS vk_user_id BIGINT UNIQUE;
