package templates

import "encoding/json"

func raw(value string) json.RawMessage {
	return json.RawMessage(value)
}

func DefaultTemplates() []Template {
	return []Template{
		{
			Name:        "React SPA",
			Description: "Public browser application using authorization code flow with PKCE.",
			AppType:     "frontend",
			DefaultRoles: raw(`[
				"access",
				"admin",
				"manager",
				"viewer"
			]`),
			DefaultRedirectURIs: raw(`[
				"http://localhost:3000/*"
			]`),
			DefaultWebOrigins: raw(`[
				"http://localhost:3000"
			]`),
			DefaultScopes: raw(`[
				"openid",
				"profile",
				"email",
				"roles"
			]`),
			DefaultMappers: raw(`[
				"user_id",
				"username",
				"full_name",
				"email",
				"client_roles",
				"realm_roles"
			]`),
			ClientConfig: raw(`{
				"client_type": "public",
				"standard_flow_enabled": true,
				"direct_access_grants_enabled": false,
				"service_accounts_enabled": false,
				"pkce_required": true
			}`),
		},
		{
			Name:        "Backend API",
			Description: "Confidential backend API client used to validate bearer tokens.",
			AppType:     "backend",
			DefaultRoles: raw(`[
				"access",
				"admin",
				"manager",
				"viewer"
			]`),
			DefaultRedirectURIs: raw(`[]`),
			DefaultWebOrigins:   raw(`[]`),
			DefaultScopes: raw(`[
				"openid",
				"profile",
				"email",
				"roles"
			]`),
			DefaultMappers: raw(`[
				"user_id",
				"username",
				"client_roles",
				"realm_roles"
			]`),
			ClientConfig: raw(`{
				"client_type": "confidential",
				"standard_flow_enabled": false,
				"direct_access_grants_enabled": false,
				"service_accounts_enabled": false,
				"pkce_required": false
			}`),
		},
		{
			Name:        "Flutter Mobile",
			Description: "Public mobile application using authorization code flow with PKCE.",
			AppType:     "mobile",
			DefaultRoles: raw(`[
				"access",
				"admin",
				"viewer"
			]`),
			DefaultRedirectURIs: raw(`[
				"com.example.app:/callback"
			]`),
			DefaultWebOrigins: raw(`[]`),
			DefaultScopes: raw(`[
				"openid",
				"profile",
				"email",
				"roles"
			]`),
			DefaultMappers: raw(`[
				"user_id",
				"username",
				"full_name",
				"email",
				"client_roles",
				"realm_roles"
			]`),
			ClientConfig: raw(`{
				"client_type": "public",
				"standard_flow_enabled": true,
				"direct_access_grants_enabled": false,
				"service_accounts_enabled": false,
				"pkce_required": true
			}`),
		},
		{
			Name:        "Machine-to-Machine",
			Description: "Confidential service client using client credentials flow.",
			AppType:     "machine_to_machine",
			DefaultRoles: raw(`[
				"access",
				"service"
			]`),
			DefaultRedirectURIs: raw(`[]`),
			DefaultWebOrigins:   raw(`[]`),
			DefaultScopes: raw(`[
				"openid",
				"profile",
				"email",
				"roles"
			]`),
			DefaultMappers: raw(`[
				"client_roles",
				"realm_roles"
			]`),
			ClientConfig: raw(`{
				"client_type": "confidential",
				"standard_flow_enabled": false,
				"direct_access_grants_enabled": false,
				"service_accounts_enabled": true,
				"pkce_required": false
			}`),
		},
	}
}
