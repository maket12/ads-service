CREATE TABLE IF NOT EXISTS account_roles (
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    role text NOT NULL CHECK ( role IN ( 'user', 'admin' ) ),
    PRIMARY KEY (account_id)
);

CREATE INDEX idx_account_roles_role ON account_roles(role);