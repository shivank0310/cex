-- Auth credentials (password hashes) and durable identity for login.
-- User profiles remain in users.accounts (user-service).

CREATE TABLE IF NOT EXISTS auth.credentials (
    user_id         TEXT PRIMARY KEY REFERENCES users.accounts (id) ON DELETE CASCADE,
    email           TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'TRADER',
    status          TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_credentials_email ON auth.credentials (email);
