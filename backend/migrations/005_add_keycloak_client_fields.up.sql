ALTER TABLE applications
ADD COLUMN keycloak_client_uuid VARCHAR(150),
ADD COLUMN keycloak_client_id VARCHAR(150),
ADD COLUMN keycloak_client_secret TEXT,
ADD COLUMN provisioned_at TIMESTAMP;