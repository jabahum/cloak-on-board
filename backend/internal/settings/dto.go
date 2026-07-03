package settings

type SaveSettingsRequest struct {
	KeycloakBaseURL           string `json:"keycloak_base_url" binding:"required"`
	KeycloakRealm             string `json:"keycloak_realm" binding:"required"`
	KeycloakAdminClientID     string `json:"keycloak_admin_client_id" binding:"required"`
	KeycloakAdminClientSecret string `json:"keycloak_admin_client_secret"`
}
