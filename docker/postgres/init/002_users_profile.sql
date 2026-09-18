-- Extend users.accounts for user-service profile fields.

ALTER TABLE users.accounts
    ADD COLUMN IF NOT EXISTS username   TEXT,
    ADD COLUMN IF NOT EXISTS kyc_status TEXT NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE UNIQUE INDEX IF NOT EXISTS users_accounts_username_idx
    ON users.accounts (username)
    WHERE username IS NOT NULL;
