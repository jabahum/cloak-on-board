CREATE TABLE onboarding_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(150) NOT NULL UNIQUE,
    description TEXT,
    app_type VARCHAR(50) NOT NULL,

    default_roles JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_redirect_uris JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_web_origins JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_mappers JSONB NOT NULL DEFAULT '[]'::jsonb,
    client_config JSONB NOT NULL DEFAULT '{}'::jsonb,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);