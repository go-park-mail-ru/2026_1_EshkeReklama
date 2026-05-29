ALTER TABLE eshkere.payment_transaction
    ADD COLUMN IF NOT EXISTS payment_type TEXT NOT NULL DEFAULT 'balance';
