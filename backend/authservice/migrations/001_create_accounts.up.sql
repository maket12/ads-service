CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS accounts (
    id uuid PRIMARY KEY,
    email citext NOT NULL UNIQUE,
    password_hash text NOT NULL,
    status text NOT NULL CHECK
        ( status IN ('active', 'blocked', 'deleted') ) DEFAULT 'active',
    email_verified boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    last_login_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_accounts_status ON accounts(status);
CREATE INDEX IF NOT EXISTS idx_accounts_email_verf ON accounts(email_verified);
