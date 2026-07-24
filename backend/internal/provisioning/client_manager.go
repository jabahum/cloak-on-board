package provisioning

import (
	"context"
	"fmt"
	"strings"

	"github.com/jabahum/keycloak-onboarder/backend/internal/applications"
	"github.com/jabahum/keycloak-onboarder/backend/internal/keycloak"
	"github.com/jabahum/keycloak-onboarder/backend/internal/settings"
)

type ClientManager struct {
	apps     *applications.Service
	settings *settings.Service
	jobs     *Service
}

func NewClientManager(apps *applications.Service, settings *settings.Service, jobs *Service) *ClientManager {
	return &ClientManager{apps: apps, settings: settings, jobs: jobs}
}

func (m *ClientManager) keycloak(ctx context.Context) (*keycloak.Client, error) {
	cfg, err := m.settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	return keycloak.NewClient(cfg.KeycloakBaseURL, cfg.KeycloakRealm, cfg.KeycloakAdminClientID, cfg.KeycloakAdminClientSecret), nil
}

func (m *ClientManager) ListKeycloakClients(ctx context.Context, search string) ([]keycloak.KeycloakClient, error) {
	kc, err := m.keycloak(ctx)
	if err != nil {
		return nil, err
	}
	clients, err := kc.ListClients(ctx, strings.TrimSpace(search))
	if err != nil {
		return nil, err
	}
	linked, err := m.apps.LinkedClientUUIDs(ctx)
	if err != nil {
		return nil, err
	}
	for i := range clients {
		clients[i].Imported = linked[clients[i].ID]
	}
	return clients, nil
}

func (m *ClientManager) ImportClient(ctx context.Context, req applications.ImportApplicationRequest) (applications.Application, error) {
	if strings.TrimSpace(req.KeycloakClientUUID) == "" {
		return applications.Application{}, fmt.Errorf("%w: Keycloak client UUID is required", applications.ErrValidation)
	}
	kc, err := m.keycloak(ctx)
	if err != nil {
		return applications.Application{}, err
	}
	client, err := kc.GetClient(ctx, req.KeycloakClientUUID)
	if err != nil {
		return applications.Application{}, err
	}
	roles, err := kc.ListClientRoles(ctx, client.ID)
	if err != nil {
		return applications.Application{}, err
	}
	if _, err := kc.GetClientScopeAssignments(ctx, client.ID); err != nil {
		return applications.Application{}, err
	}
	if _, err := kc.ListClientProtocolMappers(ctx, client.ID); err != nil {
		return applications.Application{}, err
	}
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}
	appType := strings.TrimSpace(req.AppType)
	if appType == "" {
		appType = inferApplicationType(client)
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = client.Name
	}
	if name == "" {
		name = client.ClientID
	}
	description := req.Description
	if description == "" {
		description = client.Description
	}
	return m.apps.Import(ctx, applications.CreateApplicationRequest{
		Name:         name,
		Slug:         client.ClientID,
		Description:  description,
		AppType:      appType,
		OwnerName:    req.OwnerName,
		OwnerEmail:   req.OwnerEmail,
		RedirectURIs: client.RedirectURIs,
		WebOrigins:   client.WebOrigins,
		Roles:        roleNames,
	}, client.ID, client.ClientID, client.Enabled)
}

func inferApplicationType(client keycloak.KeycloakClient) string {
	if client.ServiceAccountsEnabled {
		return "machine_to_machine"
	}
	if client.PublicClient {
		if client.Attributes["pkce.code.challenge.method"] != "" {
			return "frontend"
		}
		return "mobile"
	}
	return "backend"
}

func (m *ClientManager) UpdateApplication(ctx context.Context, id string, req applications.UpdateApplicationRequest) (applications.Application, error) {
	current, err := m.apps.GetByID(ctx, id)
	if err != nil {
		return applications.Application{}, err
	}
	req, err = m.apps.ValidateUpdate(ctx, id, req)
	if err != nil {
		return applications.Application{}, err
	}
	if current.KeycloakClientUUID == "" {
		return m.apps.Update(ctx, id, req)
	}

	job, err := m.jobs.CreateJob(ctx, CreateJobRequest{ApplicationID: id, Action: "update_application"})
	if err != nil {
		return applications.Application{}, err
	}
	_ = m.jobs.MarkJobRunning(ctx, job.ID)

	kc, err := m.keycloak(ctx)
	if err != nil {
		m.failJob(ctx, job.ID, id, err)
		return applications.Application{}, err
	}
	client, err := kc.GetClient(ctx, current.KeycloakClientUUID)
	if err != nil {
		m.failJob(ctx, job.ID, id, err)
		return applications.Application{}, err
	}
	enabled := current.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	client.ClientID = strings.TrimSpace(req.Slug)
	client.Name = strings.TrimSpace(req.Name)
	client.Description = req.Description
	client.RedirectURIs = req.RedirectURIs
	client.WebOrigins = req.WebOrigins
	client.Enabled = enabled
	applyClientType(&client, req.AppType)

	if err := m.runStep(ctx, job.ID, "update_keycloak_client", func() error {
		return kc.UpdateClient(ctx, current.KeycloakClientUUID, client)
	}); err != nil {
		m.failJob(ctx, job.ID, id, err)
		return applications.Application{}, err
	}
	if err := m.runStep(ctx, job.ID, "sync_client_roles", func() error {
		return kc.SyncClientRoles(ctx, current.KeycloakClientUUID, req.Roles)
	}); err != nil {
		m.failJob(ctx, job.ID, id, err)
		return applications.Application{}, err
	}
	updated, err := m.apps.Update(ctx, id, req)
	if err != nil {
		m.failJob(ctx, job.ID, id, err)
		return applications.Application{}, err
	}
	_ = m.apps.UpdateStatus(ctx, id, "provisioned")
	_ = m.jobs.MarkJobSucceeded(ctx, job.ID)
	updated.Status = "provisioned"
	return updated, nil
}

func applyClientType(client *keycloak.KeycloakClient, appType string) {
	client.BearerOnly = false
	client.DirectAccessGrantsEnabled = false
	client.PublicClient = appType == "frontend" || appType == "mobile"
	client.StandardFlowEnabled = client.PublicClient
	client.ServiceAccountsEnabled = appType == "machine_to_machine"
	if client.PublicClient {
		if client.Attributes == nil {
			client.Attributes = map[string]string{}
		}
		client.Attributes["pkce.code.challenge.method"] = "S256"
	} else if client.Attributes != nil {
		delete(client.Attributes, "pkce.code.challenge.method")
	}
}

func (m *ClientManager) DeleteApplication(ctx context.Context, id string, deleteKeycloak bool) error {
	app, err := m.apps.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if deleteKeycloak && app.KeycloakClientUUID != "" {
		job, err := m.jobs.CreateJob(ctx, CreateJobRequest{ApplicationID: id, Action: "delete_keycloak_client"})
		if err != nil {
			return err
		}
		_ = m.jobs.MarkJobRunning(ctx, job.ID)
		kc, err := m.keycloak(ctx)
		if err != nil {
			m.failJob(ctx, job.ID, id, err)
			return err
		}
		if err := m.runStep(ctx, job.ID, "delete_keycloak_client", func() error {
			return kc.DeleteClient(ctx, app.KeycloakClientUUID)
		}); err != nil {
			m.failJob(ctx, job.ID, id, err)
			return err
		}
		_ = m.jobs.MarkJobSucceeded(ctx, job.ID)
	}
	return m.apps.Delete(ctx, id)
}

func (m *ClientManager) GetClientScopes(ctx context.Context, id string) (keycloak.ClientScopeAssignments, error) {
	app, kc, err := m.linkedClient(ctx, id)
	if err != nil {
		return keycloak.ClientScopeAssignments{}, err
	}
	return kc.GetClientScopeAssignments(ctx, app.KeycloakClientUUID)
}

func (m *ClientManager) AssignClientScope(ctx context.Context, id, scopeID, scopeType string) error {
	app, kc, err := m.linkedClient(ctx, id)
	if err != nil {
		return err
	}
	switch scopeType {
	case "default":
		_ = kc.RemoveOptionalClientScope(ctx, app.KeycloakClientUUID, scopeID)
		return kc.AssignDefaultClientScope(ctx, app.KeycloakClientUUID, scopeID)
	case "optional":
		_ = kc.RemoveDefaultClientScope(ctx, app.KeycloakClientUUID, scopeID)
		return kc.AssignOptionalClientScope(ctx, app.KeycloakClientUUID, scopeID)
	default:
		return fmt.Errorf("%w: scope type must be default or optional", applications.ErrValidation)
	}
}

func (m *ClientManager) RemoveClientScope(ctx context.Context, id, scopeID, scopeType string) error {
	app, kc, err := m.linkedClient(ctx, id)
	if err != nil {
		return err
	}
	switch scopeType {
	case "default":
		return kc.RemoveDefaultClientScope(ctx, app.KeycloakClientUUID, scopeID)
	case "optional":
		return kc.RemoveOptionalClientScope(ctx, app.KeycloakClientUUID, scopeID)
	default:
		return fmt.Errorf("%w: scope type must be default or optional", applications.ErrValidation)
	}
}

func (m *ClientManager) ListProtocolMappers(ctx context.Context, id string) ([]keycloak.ProtocolMapper, error) {
	app, kc, err := m.linkedClient(ctx, id)
	if err != nil {
		return nil, err
	}
	return kc.ListClientProtocolMappers(ctx, app.KeycloakClientUUID)
}

func (m *ClientManager) CreateProtocolMapper(ctx context.Context, id string, mapper keycloak.ProtocolMapperRequest) error {
	if err := validateMapper(mapper); err != nil {
		return err
	}
	app, kc, err := m.linkedClient(ctx, id)
	if err != nil {
		return err
	}
	return kc.CreateClientProtocolMapper(ctx, app.KeycloakClientUUID, mapper)
}

func (m *ClientManager) UpdateProtocolMapper(ctx context.Context, id, mapperID string, mapper keycloak.ProtocolMapperRequest) error {
	if strings.TrimSpace(mapperID) == "" {
		return fmt.Errorf("%w: mapper id is required", applications.ErrValidation)
	}
	if err := validateMapper(mapper); err != nil {
		return err
	}
	app, kc, err := m.linkedClient(ctx, id)
	if err != nil {
		return err
	}
	return kc.UpdateClientProtocolMapper(ctx, app.KeycloakClientUUID, mapperID, mapper)
}

func (m *ClientManager) DeleteProtocolMapper(ctx context.Context, id, mapperID string) error {
	app, kc, err := m.linkedClient(ctx, id)
	if err != nil {
		return err
	}
	return kc.DeleteClientProtocolMapper(ctx, app.KeycloakClientUUID, mapperID)
}

func validateMapper(mapper keycloak.ProtocolMapperRequest) error {
	if strings.TrimSpace(mapper.Name) == "" || mapper.Protocol != "openid-connect" {
		return fmt.Errorf("%w: mapper name and openid-connect protocol are required", applications.ErrValidation)
	}
	switch mapper.ProtocolMapper {
	case "oidc-usermodel-attribute-mapper":
		if strings.TrimSpace(mapper.Config["user.attribute"]) == "" ||
			strings.TrimSpace(mapper.Config["claim.name"]) == "" ||
			strings.TrimSpace(mapper.Config["jsonType.label"]) == "" {
			return fmt.Errorf("%w: user attribute, claim name, and JSON type are required", applications.ErrValidation)
		}
	case "oidc-usermodel-client-role-mapper", "oidc-usermodel-realm-role-mapper":
		if strings.TrimSpace(mapper.Config["claim.name"]) == "" {
			return fmt.Errorf("%w: claim name is required", applications.ErrValidation)
		}
	default:
		return fmt.Errorf("%w: unsupported protocol mapper type", applications.ErrValidation)
	}
	return nil
}

func (m *ClientManager) linkedClient(ctx context.Context, id string) (applications.Application, *keycloak.Client, error) {
	app, err := m.apps.GetByID(ctx, id)
	if err != nil {
		return applications.Application{}, nil, err
	}
	if app.KeycloakClientUUID == "" {
		return applications.Application{}, nil, applications.ErrNotProvisioned
	}
	kc, err := m.keycloak(ctx)
	return app, kc, err
}

func (m *ClientManager) runStep(ctx context.Context, jobID, name string, operation func() error) error {
	step, err := m.jobs.CreateStep(ctx, CreateStepRequest{JobID: jobID, StepName: name})
	if err != nil {
		return err
	}
	_ = m.jobs.MarkStepRunning(ctx, step.ID)
	if err := operation(); err != nil {
		_ = m.jobs.MarkStepFailed(ctx, step.ID, err.Error())
		return err
	}
	return m.jobs.MarkStepSucceeded(ctx, step.ID)
}

func (m *ClientManager) failJob(ctx context.Context, jobID, appID string, err error) {
	_ = m.jobs.MarkJobFailed(ctx, jobID, err.Error())
	_ = m.apps.UpdateStatus(ctx, appID, "failed")
}
