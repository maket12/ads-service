CREATE TABLE IF NOT EXISTS profiles (
    account_id uuid UNIQUE NOT NULL,
    first_name varchar(20),
    last_name varchar(20),
    phone varchar(15),
    avatar_url varchar(512),
    bio varchar(512),
    updated_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_profiles_account_id ON profiles(account_id);