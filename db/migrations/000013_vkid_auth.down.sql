ALTER TABLE auth.credentials
    DROP COLUMN IF EXISTS vk_user_id;

UPDATE auth.credentials
SET email = 'vk-' || id || '@vk.local'
WHERE email IS NULL;

UPDATE auth.credentials
SET phone = LPAD(id::text, 10, '0')
WHERE phone IS NULL;

UPDATE auth.credentials
SET password_hash = ''
WHERE password_hash IS NULL;

ALTER TABLE auth.credentials
    ALTER COLUMN email SET NOT NULL,
    ALTER COLUMN phone SET NOT NULL,
    ALTER COLUMN password_hash SET NOT NULL;
