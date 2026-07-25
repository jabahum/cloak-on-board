package delivery

import (
	"encoding/json"
	"time"
)

type Environment struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	PromotionOrder int       `json:"promotion_order"`
	Protected      bool      `json:"protected"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type RealmConnection struct {
	ID             string     `json:"id"`
	EnvironmentID  string     `json:"environment_id"`
	Environment    string     `json:"environment"`
	Name           string     `json:"name"`
	BaseURL        string     `json:"base_url"`
	Realm          string     `json:"realm"`
	AdminClientID  string     `json:"admin_client_id"`
	Enabled        bool       `json:"enabled"`
	SecretSet      bool       `json:"secret_set"`
	LastTestedAt   *time.Time `json:"last_tested_at"`
	LastTestStatus string     `json:"last_test_status"`
	LastTestError  string     `json:"last_test_error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type connectionCredential struct {
	RealmConnection
	Ciphertext []byte
	Nonce      []byte
	KeyVersion string
	Legacy     string
}

type SnapshotConfiguration struct {
	ClientID        string                `json:"client_id"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	AppType         string                `json:"app_type"`
	Enabled         bool                  `json:"enabled"`
	RedirectURIs    []string              `json:"redirect_uris"`
	WebOrigins      []string              `json:"web_origins"`
	Roles           []string              `json:"roles"`
	ProtocolMappers []MapperConfiguration `json:"protocol_mappers"`
	ClientScopes    []ScopeConfiguration  `json:"client_scopes"`
}

type MapperConfiguration struct {
	Name           string            `json:"name"`
	Protocol       string            `json:"protocol"`
	ProtocolMapper string            `json:"protocol_mapper"`
	Config         map[string]string `json:"config"`
}

type ScopeConfiguration struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Snapshot struct {
	ID                string                `json:"id"`
	ApplicationID     string                `json:"application_id"`
	Version           int                   `json:"version"`
	Configuration     SnapshotConfiguration `json:"configuration"`
	ConfigurationHash string                `json:"configuration_hash"`
	CreatedBySubject  string                `json:"created_by_subject"`
	CreatedByUsername string                `json:"created_by_username"`
	CreatedAt         time.Time             `json:"created_at"`
}

type Deployment struct {
	ID                 string          `json:"id"`
	ApplicationID      string          `json:"application_id"`
	ApplicationName    string          `json:"application_name"`
	EnvironmentID      string          `json:"environment_id"`
	Environment        string          `json:"environment"`
	PromotionOrder     int             `json:"promotion_order"`
	Protected          bool            `json:"protected"`
	RealmConnectionID  string          `json:"realm_connection_id"`
	Realm              string          `json:"realm"`
	SnapshotID         string          `json:"snapshot_id"`
	PreviousSnapshotID string          `json:"previous_snapshot_id,omitempty"`
	KeycloakClientUUID string          `json:"keycloak_client_uuid,omitempty"`
	KeycloakClientID   string          `json:"keycloak_client_id"`
	Overrides          json.RawMessage `json:"overrides"`
	Status             string          `json:"status"`
	DriftStatus        string          `json:"drift_status"`
	DeployedAt         *time.Time      `json:"deployed_at"`
	DeployedBySubject  string          `json:"deployed_by_subject,omitempty"`
	LastCheckedAt      *time.Time      `json:"last_checked_at"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type DriftRun struct {
	ID                 string         `json:"id"`
	DeploymentID       string         `json:"deployment_id"`
	Status             string         `json:"status"`
	DesiredHash        string         `json:"desired_hash,omitempty"`
	ActualHash         string         `json:"actual_hash,omitempty"`
	InitiatedBySubject string         `json:"initiated_by_subject,omitempty"`
	ErrorMessage       string         `json:"error_message,omitempty"`
	StartedAt          time.Time      `json:"started_at"`
	CompletedAt        *time.Time     `json:"completed_at"`
	Findings           []DriftFinding `json:"findings,omitempty"`
}

type DriftFinding struct {
	ID           string          `json:"id"`
	DriftRunID   string          `json:"drift_run_id"`
	Path         string          `json:"path"`
	ChangeType   string          `json:"change_type"`
	DesiredValue json.RawMessage `json:"desired_value,omitempty"`
	ActualValue  json.RawMessage `json:"actual_value,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type CreateEnvironmentRequest struct {
	Name           string `json:"name" binding:"required"`
	Slug           string `json:"slug" binding:"required"`
	PromotionOrder int    `json:"promotion_order" binding:"required"`
	Protected      bool   `json:"protected"`
	Enabled        *bool  `json:"enabled"`
}

type CreateConnectionRequest struct {
	EnvironmentID string `json:"environment_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	BaseURL       string `json:"base_url" binding:"required"`
	Realm         string `json:"realm" binding:"required"`
	AdminClientID string `json:"admin_client_id" binding:"required"`
	AdminSecret   string `json:"admin_client_secret" binding:"required"`
	Enabled       *bool  `json:"enabled"`
}

type UpdateConnectionRequest struct {
	EnvironmentID string `json:"environment_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	BaseURL       string `json:"base_url" binding:"required"`
	Realm         string `json:"realm" binding:"required"`
	AdminClientID string `json:"admin_client_id" binding:"required"`
	AdminSecret   string `json:"admin_client_secret"`
	Enabled       *bool  `json:"enabled"`
}

type CreateDeploymentRequest struct {
	EnvironmentID     string          `json:"environment_id" binding:"required"`
	RealmConnectionID string          `json:"realm_connection_id" binding:"required"`
	SnapshotID        string          `json:"snapshot_id"`
	Overrides         json.RawMessage `json:"overrides"`
}

type PromotionRequest struct {
	SourceDeploymentID       string          `json:"source_deployment_id" binding:"required"`
	DestinationEnvironmentID string          `json:"destination_environment_id" binding:"required"`
	RealmConnectionID        string          `json:"realm_connection_id" binding:"required"`
	Overrides                json.RawMessage `json:"overrides"`
}

type SecretDelivery struct {
	ID           string     `json:"id"`
	DeploymentID string     `json:"deployment_id"`
	ExpiresAt    time.Time  `json:"expires_at"`
	ConsumedAt   *time.Time `json:"consumed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
