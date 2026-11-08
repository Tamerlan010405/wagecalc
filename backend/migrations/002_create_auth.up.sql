CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    display_name text,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,

    CONSTRAINT users_email_not_empty CHECK (btrim(email) <> '')
);

CREATE UNIQUE INDEX users_email_lower_unique
    ON users (lower(email));

CREATE TABLE auth_identities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider text NOT NULL,
    provider_user_id text NOT NULL,
    password_hash text,
    created_at timestamptz NOT NULL,

    CONSTRAINT auth_identities_provider_valid
        CHECK (provider IN ('password', 'max')),
    CONSTRAINT auth_identities_provider_user_id_not_empty
        CHECK (btrim(provider_user_id) <> ''),
    CONSTRAINT auth_identities_password_hash_required
        CHECK (
            (provider = 'password' AND password_hash IS NOT NULL AND btrim(password_hash) <> '')
            OR
            (provider <> 'password')
        )
);

CREATE UNIQUE INDEX auth_identities_provider_user_unique
    ON auth_identities (provider, provider_user_id);

CREATE TABLE sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,

    CONSTRAINT sessions_token_hash_not_empty CHECK (btrim(token_hash) <> '')
);

CREATE UNIQUE INDEX sessions_token_hash_unique
    ON sessions (token_hash);

CREATE INDEX sessions_user_id_idx
    ON sessions (user_id);

CREATE INDEX sessions_expires_at_idx
    ON sessions (expires_at);

CREATE TABLE organization_members (
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role text NOT NULL,
    organization_name_key text NOT NULL,
    created_at timestamptz NOT NULL,

    PRIMARY KEY (organization_id, user_id),

    CONSTRAINT organization_members_role_valid
        CHECK (role IN ('OWNER', 'ADMIN', 'EDITOR', 'VIEWER')),
    CONSTRAINT organization_members_name_key_not_empty
        CHECK (btrim(organization_name_key) <> '')
);

CREATE UNIQUE INDEX organization_members_user_name_unique
    ON organization_members (user_id, organization_name_key);

CREATE INDEX organization_members_user_id_idx
    ON organization_members (user_id);

ALTER TABLE organizations
    ADD COLUMN normalized_name text;

UPDATE organizations
SET normalized_name = lower(regexp_replace(btrim(name), '\s+', ' ', 'g'));

ALTER TABLE organizations
    ALTER COLUMN normalized_name SET NOT NULL;

ALTER TABLE organizations
    ADD CONSTRAINT organizations_normalized_name_not_empty
        CHECK (btrim(normalized_name) <> '');

DROP INDEX IF EXISTS organizations_name_lower_unique;

ALTER TABLE calculations
    ADD COLUMN created_by_user_id uuid REFERENCES users(id);