package applications

type CreateApplicationRequest struct {
	Name         string   `json:"name" binding:"required"`
	Slug         string   `json:"slug" binding:"required"`
	Description  string   `json:"description"`
	AppType      string   `json:"app_type" binding:"required"`
	OwnerName    string   `json:"owner_name"`
	OwnerEmail   string   `json:"owner_email"`
	RedirectURIs []string `json:"redirect_uris"`
	WebOrigins   []string `json:"web_origins"`
	Roles        []string `json:"roles"`
}

type UpdateApplicationRequest struct {
	Name         string   `json:"name" binding:"required"`
	Slug         string   `json:"slug" binding:"required"`
	Description  string   `json:"description"`
	AppType      string   `json:"app_type" binding:"required"`
	OwnerName    string   `json:"owner_name"`
	OwnerEmail   string   `json:"owner_email"`
	RedirectURIs []string `json:"redirect_uris"`
	WebOrigins   []string `json:"web_origins"`
	Roles        []string `json:"roles"`
	Enabled      *bool    `json:"enabled"`
}

type ImportApplicationRequest struct {
	KeycloakClientUUID string `json:"keycloak_client_uuid" binding:"required"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	AppType            string `json:"app_type"`
	OwnerName          string `json:"owner_name"`
	OwnerEmail         string `json:"owner_email"`
}

type AssignClientScopeRequest struct {
	Type string `json:"type" binding:"required"`
}
