ALTER TABLE calculations
    DROP COLUMN IF EXISTS created_by_user_id;

ALTER TABLE organizations
    DROP CONSTRAINT IF EXISTS organizations_normalized_name_not_empty;

ALTER TABLE organizations
    DROP COLUMN IF EXISTS normalized_name;

CREATE UNIQUE INDEX IF NOT EXISTS organizations_name_lower_unique
    ON organizations (lower(name));

DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS auth_identities;
DROP TABLE IF EXISTS users;