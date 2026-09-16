-- CEX PostgreSQL bootstrap schema (future persistence layer).
-- Services currently use in-memory stores; this schema is ready for migration.

CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS users;
CREATE SCHEMA IF NOT EXISTS orders;
CREATE SCHEMA IF NOT EXISTS ledger;
CREATE SCHEMA IF NOT EXISTS wallet;
CREATE SCHEMA IF NOT EXISTS settlement;

-- Users
CREATE TABLE IF NOT EXISTS users.accounts (
    id          TEXT PRIMARY KEY,
    email       TEXT UNIQUE NOT NULL,
    status      TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ledger accounts (mirrors ledger-service model)
CREATE TABLE IF NOT EXISTS ledger.accounts (
    user_id     TEXT NOT NULL,
    asset       TEXT NOT NULL,
    available   BIGINT NOT NULL DEFAULT 0,
    locked      BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, asset)
);

CREATE TABLE IF NOT EXISTS ledger.journals (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL,
    reference   TEXT NOT NULL UNIQUE,
    symbol      TEXT,
    posted_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ledger.legs (
    id          TEXT PRIMARY KEY,
    journal_id  TEXT NOT NULL REFERENCES ledger.journals(id),
    user_id     TEXT NOT NULL,
    asset       TEXT NOT NULL,
    type        TEXT NOT NULL,
    amount      BIGINT NOT NULL,
    reference   TEXT
);

-- Wallet
CREATE TABLE IF NOT EXISTS wallet.wallets (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    asset       TEXT NOT NULL,
    chain       TEXT NOT NULL,
    address     TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, asset)
);

CREATE TABLE IF NOT EXISTS wallet.deposits (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL,
    asset           TEXT NOT NULL,
    amount          BIGINT NOT NULL,
    tx_hash         TEXT NOT NULL UNIQUE,
    to_address      TEXT NOT NULL,
    confirmations   INT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wallet.withdrawals (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    asset       TEXT NOT NULL,
    amount      BIGINT NOT NULL,
    to_address  TEXT NOT NULL,
    tx_hash     TEXT,
    status      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Orders (future order-service persistence)
CREATE TABLE IF NOT EXISTS orders.orders (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    symbol      TEXT NOT NULL,
    side        TEXT NOT NULL,
    type        TEXT NOT NULL,
    status      TEXT NOT NULL,
    price       BIGINT NOT NULL,
    quantity    BIGINT NOT NULL,
    remaining   BIGINT NOT NULL,
    filled      BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Settlement records
CREATE TABLE IF NOT EXISTS settlement.settlements (
    id          TEXT PRIMARY KEY,
    trade_id    TEXT NOT NULL UNIQUE,
    journal_id  TEXT,
    symbol      TEXT NOT NULL,
    buyer_id    TEXT NOT NULL,
    seller_id   TEXT NOT NULL,
    status      TEXT NOT NULL,
    settled_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
