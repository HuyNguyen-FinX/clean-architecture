CREATE TABLE IF NOT EXISTS transfers (
    id TEXT PRIMARY KEY,
    from_account_id TEXT NOT NULL REFERENCES accounts(id),
    to_account_id TEXT NOT NULL REFERENCES accounts(id),
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency CHAR(3) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    CHECK (from_account_id <> to_account_id)
);

CREATE INDEX IF NOT EXISTS transfers_from_created_idx
    ON transfers (from_account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS transfers_to_created_idx
    ON transfers (to_account_id, created_at DESC);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key TEXT PRIMARY KEY,
    request_hash CHAR(64) NOT NULL,
    transfer_id TEXT UNIQUE REFERENCES transfers(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS outbox (
    id TEXT PRIMARY KEY,
    topic TEXT NOT NULL,
    message_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS outbox_pending_idx
    ON outbox (created_at)
    WHERE published_at IS NULL;
