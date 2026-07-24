package applications

import "time"

type Application struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description"`
	AppType       string    `json:"app_type"`
	OwnerName     string    `json:"owner_name"`
	OwnerEmail    string    `json:"owner_email"`
	Status        string    `json:"status"`
	Source        string    `json:"source"`
	Enabled       bool      `json:"enabled"`
	ConfigVersion int       `json:"config_version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	RedirectURIs []string `json:"redirect_uris,omitempty"`
	WebOrigins   []string `json:"web_origins,omitempty"`
	Roles        []string `json:"roles,omitempty"`

	KeycloakClientUUID   string     `json:"keycloak_client_uuid"`
	KeycloakClientID     string     `json:"keycloak_client_id"`
	KeycloakClientSecret string     `json:"-"`
	ProvisionedAt        *time.Time `json:"provisioned_at"`
}
