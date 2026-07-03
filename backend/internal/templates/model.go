package templates

import (
	"encoding/json"
	"time"
)

type Template struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	Description         string          `json:"description"`
	AppType             string          `json:"app_type"`
	DefaultRoles        json.RawMessage `json:"default_roles"`
	DefaultRedirectURIs json.RawMessage `json:"default_redirect_uris"`
	DefaultWebOrigins   json.RawMessage `json:"default_web_origins"`
	DefaultScopes       json.RawMessage `json:"default_scopes"`
	DefaultMappers      json.RawMessage `json:"default_mappers"`
	ClientConfig        json.RawMessage `json:"client_config"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}
