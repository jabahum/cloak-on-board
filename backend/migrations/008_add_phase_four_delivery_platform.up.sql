CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(80) NOT NULL UNIQUE,
    promotion_order INTEGER NOT NULL UNIQUE,
    protected BOOLEAN NOT NULL DEFAULT FALSE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE realm_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE RESTRICT,
    name VARCHAR(120) NOT NULL,
    base_url TEXT NOT NULL,
    realm VARCHAR(100) NOT NULL,
    admin_client_id VARCHAR(150) NOT NULL,
    admin_secret_ciphertext BYTEA,
    admin_secret_nonce BYTEA,
    encryption_key_version VARCHAR(50),
    legacy_admin_secret TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_tested_at TIMESTAMP,
    last_test_status VARCHAR(30),
    last_test_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(environment_id, name),
    UNIQUE(base_url, realm)
);

CREATE TABLE application_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    configuration JSONB NOT NULL,
    configuration_hash VARCHAR(64) NOT NULL,
    created_by_subject VARCHAR(255) NOT NULL,
    created_by_username VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(application_id, version),
    UNIQUE(application_id, configuration_hash)
);

CREATE FUNCTION prevent_application_snapshot_update() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'application snapshots are immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER application_snapshots_immutable
BEFORE UPDATE ON application_snapshots
FOR EACH ROW EXECUTE FUNCTION prevent_application_snapshot_update();

CREATE TABLE application_deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE RESTRICT,
    realm_connection_id UUID NOT NULL REFERENCES realm_connections(id) ON DELETE RESTRICT,
    snapshot_id UUID REFERENCES application_snapshots(id) ON DELETE RESTRICT,
    previous_snapshot_id UUID REFERENCES application_snapshots(id) ON DELETE RESTRICT,
    keycloak_client_uuid VARCHAR(150),
    keycloak_client_id VARCHAR(150) NOT NULL,
    overrides JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(30) NOT NULL DEFAULT 'not_deployed',
    drift_status VARCHAR(30) NOT NULL DEFAULT 'unknown',
    deployed_at TIMESTAMP,
    deployed_by_subject VARCHAR(255),
    last_checked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(application_id, environment_id),
    UNIQUE(realm_connection_id, keycloak_client_id),
    UNIQUE(realm_connection_id, keycloak_client_uuid)
);

CREATE TABLE drift_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    deployment_id UUID NOT NULL REFERENCES application_deployments(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL DEFAULT 'checking',
    desired_hash VARCHAR(64),
    actual_hash VARCHAR(64),
    initiated_by_subject VARCHAR(255),
    error_message TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE TABLE drift_findings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    drift_run_id UUID NOT NULL REFERENCES drift_runs(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    change_type VARCHAR(20) NOT NULL,
    desired_value JSONB,
    actual_value JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE secret_deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    deployment_id UUID NOT NULL REFERENCES application_deployments(id) ON DELETE CASCADE,
    recipient_subject VARCHAR(255) NOT NULL,
    secret_ciphertext BYTEA,
    secret_nonce BYTEA,
    encryption_key_version VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    consumed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE provisioning_jobs
    ADD COLUMN environment_id UUID REFERENCES environments(id) ON DELETE SET NULL,
    ADD COLUMN realm_connection_id UUID REFERENCES realm_connections(id) ON DELETE SET NULL,
    ADD COLUMN deployment_id UUID REFERENCES application_deployments(id) ON DELETE SET NULL,
    ADD COLUMN snapshot_id UUID REFERENCES application_snapshots(id) ON DELETE SET NULL;

ALTER TABLE approval_requests
    ADD COLUMN environment_id UUID REFERENCES environments(id) ON DELETE SET NULL,
    ADD COLUMN realm_connection_id UUID REFERENCES realm_connections(id) ON DELETE SET NULL,
    ADD COLUMN deployment_id UUID REFERENCES application_deployments(id) ON DELETE SET NULL,
    ADD COLUMN snapshot_id UUID REFERENCES application_snapshots(id) ON DELETE SET NULL;

CREATE INDEX application_deployments_application_idx ON application_deployments(application_id);
CREATE INDEX application_deployments_environment_idx ON application_deployments(environment_id);
CREATE INDEX drift_runs_deployment_idx ON drift_runs(deployment_id, started_at DESC);
CREATE INDEX drift_findings_run_idx ON drift_findings(drift_run_id);
CREATE INDEX secret_deliveries_recipient_idx ON secret_deliveries(recipient_subject, expires_at);

INSERT INTO environments(name, slug, promotion_order, protected) VALUES
    ('Development', 'development', 10, FALSE),
    ('Staging', 'staging', 20, FALSE),
    ('Production', 'production', 30, TRUE);

INSERT INTO realm_connections(
    environment_id, name, base_url, realm, admin_client_id, legacy_admin_secret
)
SELECT e.id, 'Migrated default realm', s.keycloak_base_url, s.keycloak_realm,
       s.keycloak_admin_client_id, s.keycloak_admin_client_secret
FROM settings s
JOIN environments e ON e.slug = 'development'
ORDER BY s.updated_at DESC
LIMIT 1;

INSERT INTO application_snapshots(
    application_id, version, configuration, configuration_hash,
    created_by_subject, created_by_username
)
SELECT a.id, a.config_version,
       jsonb_build_object(
           'client_id', COALESCE(NULLIF(a.keycloak_client_id, ''), a.slug),
           'name', a.name,
           'description', COALESCE(a.description, ''),
           'app_type', a.app_type,
           'enabled', a.enabled,
           'redirect_uris', COALESCE((SELECT jsonb_agg(r.redirect_uri ORDER BY r.redirect_uri)
              FROM application_redirect_uris r WHERE r.application_id=a.id), '[]'::jsonb),
           'web_origins', COALESCE((SELECT jsonb_agg(w.web_origin ORDER BY w.web_origin)
              FROM application_web_origins w WHERE w.application_id=a.id), '[]'::jsonb),
           'roles', COALESCE((SELECT jsonb_agg(ro.role_name ORDER BY ro.role_name)
              FROM application_roles ro WHERE ro.application_id=a.id), '[]'::jsonb)
       ),
       encode(digest(jsonb_build_object(
           'client_id', COALESCE(NULLIF(a.keycloak_client_id, ''), a.slug),
           'name', a.name, 'description', COALESCE(a.description, ''),
           'app_type', a.app_type, 'enabled', a.enabled,
           'redirect_uris', COALESCE((SELECT jsonb_agg(r.redirect_uri ORDER BY r.redirect_uri)
              FROM application_redirect_uris r WHERE r.application_id=a.id), '[]'::jsonb),
           'web_origins', COALESCE((SELECT jsonb_agg(w.web_origin ORDER BY w.web_origin)
              FROM application_web_origins w WHERE w.application_id=a.id), '[]'::jsonb),
           'roles', COALESCE((SELECT jsonb_agg(ro.role_name ORDER BY ro.role_name)
              FROM application_roles ro WHERE ro.application_id=a.id), '[]'::jsonb)
       )::text, 'sha256'), 'hex'),
       'migration', 'Phase 4 migration'
FROM applications a;

INSERT INTO application_deployments(
    application_id, environment_id, realm_connection_id, snapshot_id,
    keycloak_client_uuid, keycloak_client_id, status, drift_status, deployed_at
)
SELECT a.id, e.id, rc.id, s.id, a.keycloak_client_uuid,
       COALESCE(NULLIF(a.keycloak_client_id, ''), a.slug),
       CASE WHEN a.keycloak_client_uuid IS NULL THEN 'not_deployed' ELSE 'deployed' END,
       'unknown', a.provisioned_at
FROM applications a
JOIN application_snapshots s ON s.application_id=a.id AND s.version=a.config_version
JOIN environments e ON e.slug='development'
JOIN realm_connections rc ON rc.environment_id=e.id
ON CONFLICT DO NOTHING;
