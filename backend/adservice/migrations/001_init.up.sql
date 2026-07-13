DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ad_status') THEN
            CREATE TYPE ad_status AS ENUM (
                'published', 'on_moderation', 'rejected', 'deleted'
            );
        END IF;
    END
$$;

CREATE TABLE IF NOT EXISTS ads (
    id uuid PRIMARY KEY,
    seller_id uuid NOT NULL, -- i.e. account_id
    title varchar(255) NOT NULL,
    description text,
    price BIGINT NOT NULL DEFAULT 0, -- in cents
    status ad_status NOT NULL DEFAULT 'on_moderation',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
