ALTER TABLE applications
DROP COLUMN IF EXISTS provisioned_at,
DROP COLUMN IF EXISTS keycloak_client_secret,
DROP COLUMN IF EXISTS keycloak_client_id,
DROP COLUMN IF EXISTS keycloak_client_uuid;