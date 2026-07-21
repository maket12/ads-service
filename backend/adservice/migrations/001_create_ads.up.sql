CREATE TABLE IF NOT EXISTS ads (
    id uuid PRIMARY KEY,
    seller_id uuid NOT NULL, -- i.e. account_id
    title varchar(255) NOT NULL,
    description text,
    price bigint NOT NULL DEFAULT 0, -- in cents
    status text NOT NULL CHECK
        ( status IN
          ('published', 'on_moderation', 'rejected', 'deleted')
        ) DEFAULT 'on_moderation',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ads_seller_id ON ads(seller_id);
CREATE INDEX IF NOT EXISTS idx_ads_status ON ads(status);
