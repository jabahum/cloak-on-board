CREATE TABLE settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    keycloak_base_url TEXT NOT NULL,
    keycloak_realm VARCHAR(100) NOT NULL,
    keycloak_admin_client_id VARCHAR(150) NOT NULL,
    keycloak_admin_client_secret TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);