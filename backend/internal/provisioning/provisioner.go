package provisioning

import (
	"context"
	"fmt"

	"github.com/jabahum/keycloak-onboarder/backend/internal/applications"
	"github.com/jabahum/keycloak-onboarder/backend/internal/keycloak"
	"github.com/jabahum/keycloak-onboarder/backend/internal/settings"
)

type Provisioner struct {
	appService      *applications.Service
	settingsService *settings.Service
	jobService      *Service
}

func NewProvisioner(
	appService *applications.Service,
	settingsService *settings.Service,
	jobService *Service,
) *Provisioner {
	return &Provisioner{
		appService:      appService,
		settingsService: settingsService,
		jobService:      jobService,
	}
}

func (p *Provisioner) ProvisionApplication(ctx context.Context, applicationID string) (Job, error) {
	app, err := p.appService.GetByID(ctx, applicationID)
	if err != nil {
		return Job{}, err
	}

	job, err := p.jobService.CreateJob(ctx, CreateJobRequest{
		ApplicationID: applicationID,
		Action:        "provision_application",
	})
	if err != nil {
		return Job{}, err
	}

	if err := p.jobService.MarkJobRunning(ctx, job.ID); err != nil {
		return Job{}, err
	}
	if err := p.appService.UpdateStatus(ctx, applicationID, "provisioning"); err != nil {
		return Job{}, err
	}

	if err := p.runProvisioning(ctx, job.ID, app); err != nil {
		_ = p.jobService.MarkJobFailed(ctx, job.ID, err.Error())
		_ = p.appService.UpdateStatus(ctx, applicationID, "failed")
		return p.jobService.GetJobByID(ctx, job.ID)
	}

	if err := p.jobService.MarkJobSucceeded(ctx, job.ID); err != nil {
		return Job{}, err
	}

	return p.jobService.GetJobByID(ctx, job.ID)
}

func (p *Provisioner) runProvisioning(ctx context.Context, jobID string, app applications.Application) error {
	cfg, err := p.settingsService.Get(ctx)
	if err != nil {
		return err
	}

	kc := keycloak.NewClient(
		cfg.KeycloakBaseURL,
		cfg.KeycloakRealm,
		cfg.KeycloakAdminClientID,
		cfg.KeycloakAdminClientSecret,
	)

	client, err := p.createClientStep(ctx, jobID, kc, app)
	if err != nil {
		return err
	}

	if err := p.createRolesStep(ctx, jobID, kc, client.ID, app.Roles); err != nil {
		return err
	}

	if err := p.createMappersStep(ctx, jobID, kc, client.ID); err != nil {
		return err
	}

	clientSecret := ""

	if app.AppType == "backend" || app.AppType == "machine_to_machine" {
		secret, err := p.getClientSecretStep(ctx, jobID, kc, client.ID)
		if err != nil {
			return err
		}

		clientSecret = secret
	}

	if err := p.appService.MarkProvisioned(
		ctx,
		app.ID,
		client.ID,
		client.ClientID,
		clientSecret,
	); err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) createClientStep(
	ctx context.Context,
	jobID string,
	kc *keycloak.Client,
	app applications.Application,
) (keycloak.KeycloakClient, error) {
	step, err := p.jobService.CreateStep(ctx, CreateStepRequest{
		JobID:    jobID,
		StepName: "create_keycloak_client",
	})
	if err != nil {
		return keycloak.KeycloakClient{}, err
	}

	_ = p.jobService.MarkStepRunning(ctx, step.ID)

	req := keycloak.CreateClientRequest{
		ClientID:                  app.Slug,
		Name:                      app.Name,
		Description:               app.Description,
		Protocol:                  "openid-connect",
		PublicClient:              app.AppType == "frontend" || app.AppType == "mobile",
		BearerOnly:                false,
		StandardFlowEnabled:       app.AppType == "frontend" || app.AppType == "mobile",
		DirectAccessGrantsEnabled: false,
		ServiceAccountsEnabled:    app.AppType == "machine_to_machine",
		RedirectURIs:              app.RedirectURIs,
		WebOrigins:                app.WebOrigins,
		Enabled:                   true,
	}

	if app.AppType == "frontend" || app.AppType == "mobile" {
		req.Attributes = map[string]string{
			"pkce.code.challenge.method": "S256",
		}
	}

	if app.AppType == "backend" {
		req.PublicClient = false
		req.StandardFlowEnabled = false
		req.ServiceAccountsEnabled = false
	}

	client, err := kc.CreateOrGetClient(ctx, req)
	if err != nil {
		_ = p.jobService.MarkStepFailed(ctx, step.ID, err.Error())
		return keycloak.KeycloakClient{}, err
	}

	_ = p.jobService.MarkStepSucceeded(ctx, step.ID)

	return client, nil
}

func (p *Provisioner) createRolesStep(
	ctx context.Context,
	jobID string,
	kc *keycloak.Client,
	clientUUID string,
	roles []string,
) error {
	step, err := p.jobService.CreateStep(ctx, CreateStepRequest{
		JobID:    jobID,
		StepName: "create_client_roles",
	})
	if err != nil {
		return err
	}

	_ = p.jobService.MarkStepRunning(ctx, step.ID)

	if err := kc.CreateClientRoles(ctx, clientUUID, roles); err != nil {
		_ = p.jobService.MarkStepFailed(ctx, step.ID, err.Error())
		return err
	}

	_ = p.jobService.MarkStepSucceeded(ctx, step.ID)

	return nil
}

func (p *Provisioner) createMappersStep(
	ctx context.Context,
	jobID string,
	kc *keycloak.Client,
	clientUUID string,
) error {
	step, err := p.jobService.CreateStep(ctx, CreateStepRequest{
		JobID:    jobID,
		StepName: "create_protocol_mappers",
	})
	if err != nil {
		return err
	}

	_ = p.jobService.MarkStepRunning(ctx, step.ID)

	if err := kc.CreateDefaultProtocolMappers(ctx, clientUUID); err != nil {
		_ = p.jobService.MarkStepFailed(ctx, step.ID, err.Error())
		return fmt.Errorf("create protocol mappers: %w", err)
	}

	_ = p.jobService.MarkStepSucceeded(ctx, step.ID)

	return nil
}

func (p *Provisioner) getClientSecretStep(
	ctx context.Context,
	jobID string,
	kc *keycloak.Client,
	clientUUID string,
) (string, error) {
	step, err := p.jobService.CreateStep(ctx, CreateStepRequest{
		JobID:    jobID,
		StepName: "get_client_secret",
	})
	if err != nil {
		return "", err
	}

	_ = p.jobService.MarkStepRunning(ctx, step.ID)

	secret, err := kc.GetClientSecret(ctx, clientUUID)
	if err != nil {
		_ = p.jobService.MarkStepFailed(ctx, step.ID, err.Error())
		return "", err
	}

	_ = p.jobService.MarkStepSucceeded(ctx, step.ID)

	return secret, nil
}
