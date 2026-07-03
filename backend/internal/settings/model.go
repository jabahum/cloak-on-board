package settings

import "time"

type Settings struct {
	ID                        string    `json:"id"`
	KeycloakBaseURL           string    `json:"keycloak_base_url"`
	KeycloakRealm             string    `json:"keycloak_realm"`
	KeycloakAdminClientID     string    `json:"keycloak_admin_client_id"`
	KeycloakAdminClientSecret string    `json:"-"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}
