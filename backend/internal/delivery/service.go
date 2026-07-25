package delivery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jabahum/keycloak-onboarder/backend/internal/applications"
	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
	"github.com/jabahum/keycloak-onboarder/backend/internal/credentials"
	"github.com/jabahum/keycloak-onboarder/backend/internal/keycloak"
	"github.com/jabahum/keycloak-onboarder/backend/internal/notifications"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db            *pgxpool.Pool
	repo          *Repository
	apps          *applications.Service
	keyring       *credentials.Keyring
	notifications *notifications.Service
	deliveryTTL   time.Duration
}

func NewService(db *pgxpool.Pool, apps *applications.Service, ring *credentials.Keyring, n *notifications.Service, ttl time.Duration) *Service {
	return &Service{db: db, repo: NewRepository(db), apps: apps, keyring: ring, notifications: n, deliveryTTL: ttl}
}

func (s *Service) ListEnvironments(ctx context.Context) ([]Environment, error) {
	return s.repo.ListEnvironments(ctx)
}

func (s *Service) CreateEnvironment(ctx context.Context, input CreateEnvironmentRequest) (Environment, error) {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Slug) == "" || input.PromotionOrder <= 0 {
		return Environment{}, fmt.Errorf("%w: name, slug, and a positive promotion order are required", ErrValidation)
	}
	return s.repo.CreateEnvironment(ctx, input)
}

func (s *Service) UpdateEnvironment(ctx context.Context, id string, input CreateEnvironmentRequest) (Environment, error) {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Slug) == "" || input.PromotionOrder <= 0 {
		return Environment{}, fmt.Errorf("%w: name, slug, and a positive promotion order are required", ErrValidation)
	}
	return s.repo.UpdateEnvironment(ctx, id, input)
}

func (s *Service) DeleteEnvironment(ctx context.Context, id string) error {
	return s.repo.DeleteEnvironment(ctx, id)
}

func (s *Service) ListConnections(ctx context.Context) ([]RealmConnection, error) {
	return s.repo.ListConnections(ctx)
}

func (s *Service) CreateConnection(ctx context.Context, input CreateConnectionRequest) (RealmConnection, error) {
	if s.keyring == nil {
		return RealmConnection{}, credentials.ErrNoKeys
	}
	if !strings.HasPrefix(input.BaseURL, "http://") && !strings.HasPrefix(input.BaseURL, "https://") {
		return RealmConnection{}, fmt.Errorf("%w: base_url must be an HTTP or HTTPS URL", ErrValidation)
	}
	// Generate the ID first so the connection ID can be authenticated as AES-GCM AAD.
	var id string
	err := s.db.QueryRow(ctx, `SELECT uuid_generate_v4()`).Scan(&id)
	if err != nil {
		return RealmConnection{}, err
	}
	ciphertext, nonce, version, err := s.keyring.Encrypt([]byte(input.AdminSecret), id)
	if err != nil {
		return RealmConnection{}, err
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	err = s.db.QueryRow(ctx, `INSERT INTO realm_connections(id,environment_id,name,base_url,realm,
		admin_client_id,admin_secret_ciphertext,admin_secret_nonce,encryption_key_version,enabled)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`, id, input.EnvironmentID, input.Name,
		strings.TrimRight(input.BaseURL, "/"), input.Realm, input.AdminClientID, ciphertext, nonce, version, enabled).Scan(&id)
	if err != nil {
		return RealmConnection{}, err
	}
	credential, err := s.repo.GetConnectionCredential(ctx, id)
	return credential.RealmConnection, err
}

func (s *Service) DisableConnection(ctx context.Context, id string) error {
	return s.repo.DisableConnection(ctx, id)
}

func (s *Service) UpdateConnection(ctx context.Context, id string, input UpdateConnectionRequest) (RealmConnection, error) {
	if !strings.HasPrefix(input.BaseURL, "http://") && !strings.HasPrefix(input.BaseURL, "https://") {
		return RealmConnection{}, fmt.Errorf("%w: base_url must be an HTTP or HTTPS URL", ErrValidation)
	}
	var ciphertext, nonce []byte
	var version string
	var err error
	encrypted := input.AdminSecret != ""
	if encrypted {
		if s.keyring == nil {
			return RealmConnection{}, credentials.ErrNoKeys
		}
		ciphertext, nonce, version, err = s.keyring.Encrypt([]byte(input.AdminSecret), id)
		if err != nil {
			return RealmConnection{}, err
		}
	}
	input.BaseURL = strings.TrimRight(input.BaseURL, "/")
	return s.repo.UpdateConnection(ctx, id, input, encrypted, ciphertext, nonce, version)
}

func (s *Service) connectionClient(ctx context.Context, id string, allowDisabled bool) (*keycloak.Client, RealmConnection, error) {
	credential, err := s.repo.GetConnectionCredential(ctx, id)
	if err != nil {
		return nil, RealmConnection{}, err
	}
	if !credential.Enabled && !allowDisabled {
		return nil, RealmConnection{}, fmt.Errorf("%w: realm connection is disabled", ErrUnavailable)
	}
	if s.keyring == nil {
		return nil, RealmConnection{}, credentials.ErrNoKeys
	}
	var secret []byte
	if credential.Legacy != "" {
		ciphertext, nonce, version, encryptErr := s.keyring.Encrypt([]byte(credential.Legacy), credential.ID)
		if encryptErr != nil {
			return nil, RealmConnection{}, encryptErr
		}
		if err := s.repo.SaveMigratedCredential(ctx, credential.ID, ciphertext, nonce, version); err != nil {
			return nil, RealmConnection{}, err
		}
		secret = []byte(credential.Legacy)
	} else {
		secret, err = s.keyring.Decrypt(credential.Ciphertext, credential.Nonce, credential.KeyVersion, credential.ID)
		if err != nil {
			return nil, RealmConnection{}, fmt.Errorf("decrypt realm credential: %w", err)
		}
	}
	return keycloak.NewClient(credential.BaseURL, credential.Realm, credential.AdminClientID, string(secret)), credential.RealmConnection, nil
}

func (s *Service) TestConnection(ctx context.Context, id string) (RealmConnection, error) {
	kc, _, err := s.connectionClient(ctx, id, true)
	if err == nil {
		_, err = kc.ListClients(ctx, "")
	}
	status, message := "succeeded", ""
	if err != nil {
		status, message = "failed", "target realm authentication failed"
	}
	_ = s.repo.SetConnectionTest(ctx, id, status, message)
	credential, getErr := s.repo.GetConnectionCredential(ctx, id)
	if getErr != nil {
		return RealmConnection{}, getErr
	}
	if err != nil {
		return credential.RealmConnection, fmt.Errorf("%w: connection test failed", ErrUnavailable)
	}
	return credential.RealmConnection, nil
}

func (s *Service) CreateSnapshot(ctx context.Context, appID string, user auth.User) (Snapshot, error) {
	app, err := s.apps.GetByID(ctx, appID)
	if err != nil {
		return Snapshot{}, err
	}
	config := SnapshotConfiguration{
		ClientID: app.Slug, Name: app.Name, Description: app.Description, AppType: app.AppType,
		Enabled: app.Enabled, RedirectURIs: sorted(app.RedirectURIs), WebOrigins: sorted(app.WebOrigins),
		Roles: sorted(app.Roles), ProtocolMappers: defaultMapperConfiguration(), ClientScopes: defaultScopeConfiguration(),
	}
	hash, _, err := canonicalHash(config)
	if err != nil {
		return Snapshot{}, err
	}
	return s.repo.CreateSnapshot(ctx, appID, user.Subject, user.Username, config, hash)
}

func (s *Service) ListSnapshots(ctx context.Context, appID string) ([]Snapshot, error) {
	return s.repo.ListSnapshots(ctx, appID)
}

func (s *Service) ListDeployments(ctx context.Context, appID string) ([]Deployment, error) {
	return s.repo.ListDeployments(ctx, appID)
}

func (s *Service) Deploy(ctx context.Context, appID string, input CreateDeploymentRequest, subject string, approved bool) (Deployment, string, error) {
	snapshot, err := s.snapshotForDeployment(ctx, appID, input.SnapshotID, subject)
	if err != nil {
		return Deployment{}, "", err
	}
	env, err := s.repo.getEnvironment(ctx, input.EnvironmentID)
	if err != nil {
		return Deployment{}, "", err
	}
	if env.Protected && !approved {
		return Deployment{}, "", ErrForbidden
	}
	if err := validateOverrides(input.Overrides); err != nil {
		return Deployment{}, "", err
	}
	connection, err := s.repo.GetConnectionCredential(ctx, input.RealmConnectionID)
	if err != nil {
		return Deployment{}, "", err
	}
	if connection.EnvironmentID != env.ID || !env.Enabled || !connection.Enabled {
		return Deployment{}, "", fmt.Errorf("%w: connection is not enabled for the selected environment", ErrUnavailable)
	}
	deployment, err := s.repo.UpsertDeployment(ctx, appID, input, snapshot)
	if err != nil {
		return Deployment{}, "", err
	}
	jobID, err := s.createJob(ctx, deployment, snapshot.ID, "deploy_application")
	if err != nil {
		return Deployment{}, "", err
	}
	if err := s.applySnapshot(ctx, deployment, snapshot, subject); err != nil {
		s.failJob(ctx, jobID, deployment.ID, err)
		_, _ = s.db.Exec(ctx, `UPDATE application_deployments SET status='failed',updated_at=NOW() WHERE id=$1`, deployment.ID)
		return deployment, jobID, err
	}
	s.succeedJob(ctx, jobID)
	updated, err := s.repo.GetDeployment(ctx, deployment.ID)
	return updated, jobID, err
}

func (s *Service) Promote(ctx context.Context, appID string, input PromotionRequest, subject string, approved bool) (Deployment, string, error) {
	source, err := s.repo.GetDeployment(ctx, input.SourceDeploymentID)
	if err != nil {
		return Deployment{}, "", err
	}
	if source.ApplicationID != appID || source.Status != "deployed" || source.SnapshotID == "" {
		return Deployment{}, "", ErrConflict
	}
	destination, err := s.repo.getEnvironment(ctx, input.DestinationEnvironmentID)
	if err != nil {
		return Deployment{}, "", err
	}
	previous, err := s.repo.PreviousEnvironmentDeployment(ctx, appID, destination.PromotionOrder)
	if err != nil || previous.EnvironmentID != source.EnvironmentID {
		return Deployment{}, "", fmt.Errorf("%w: promotions must use the immediately preceding deployed environment", ErrConflict)
	}
	if destination.Protected && !approved {
		return Deployment{}, "", ErrForbidden
	}
	latest, err := s.repo.LatestSnapshot(ctx, appID)
	if err != nil {
		return Deployment{}, "", err
	}
	if latest.ID != source.SnapshotID {
		return Deployment{}, "", fmt.Errorf("%w: source deployment does not use the latest snapshot", ErrConflict)
	}
	return s.Deploy(ctx, appID, CreateDeploymentRequest{
		EnvironmentID: input.DestinationEnvironmentID, RealmConnectionID: input.RealmConnectionID,
		SnapshotID: source.SnapshotID, Overrides: input.Overrides,
	}, subject, approved)
}

func (s *Service) Rollback(ctx context.Context, deploymentID, subject string, approved bool) (Deployment, string, error) {
	deployment, err := s.repo.GetDeployment(ctx, deploymentID)
	if err != nil {
		return Deployment{}, "", err
	}
	if deployment.Protected && !approved {
		return Deployment{}, "", ErrForbidden
	}
	if deployment.PreviousSnapshotID == "" {
		return Deployment{}, "", fmt.Errorf("%w: deployment has no previous snapshot", ErrConflict)
	}
	return s.Deploy(ctx, deployment.ApplicationID, CreateDeploymentRequest{
		EnvironmentID: deployment.EnvironmentID, RealmConnectionID: deployment.RealmConnectionID,
		SnapshotID: deployment.PreviousSnapshotID, Overrides: deployment.Overrides,
	}, subject, true)
}

func (s *Service) snapshotForDeployment(ctx context.Context, appID, snapshotID, subject string) (Snapshot, error) {
	var snapshot Snapshot
	var err error
	if snapshotID == "" {
		snapshot, err = s.repo.LatestSnapshot(ctx, appID)
		if errors.Is(err, ErrNotFound) {
			return s.CreateSnapshot(ctx, appID, auth.User{Subject: subject, Username: subject})
		}
	} else {
		snapshot, err = s.repo.GetSnapshot(ctx, snapshotID)
	}
	if err == nil && snapshot.ApplicationID != appID {
		return Snapshot{}, ErrConflict
	}
	return snapshot, err
}

func (s *Service) applySnapshot(ctx context.Context, deployment Deployment, snapshot Snapshot, subject string) error {
	kc, _, err := s.connectionClient(ctx, deployment.RealmConnectionID, false)
	if err != nil {
		return err
	}
	config, err := applyOverrides(snapshot.Configuration, deployment.Overrides)
	if err != nil {
		return err
	}
	request := toKeycloakRequest(config)
	client, err := kc.CreateOrGetClient(ctx, request)
	if err != nil {
		return fmt.Errorf("%w: target Keycloak operation failed", ErrUnavailable)
	}
	client.ClientID, client.Name, client.Description = config.ClientID, config.Name, config.Description
	client.Enabled = config.Enabled
	client.PublicClient = config.AppType == "frontend" || config.AppType == "mobile"
	client.StandardFlowEnabled = client.PublicClient
	client.ServiceAccountsEnabled = config.AppType == "machine_to_machine"
	client.RedirectURIs, client.WebOrigins = config.RedirectURIs, config.WebOrigins
	if err := kc.UpdateClient(ctx, client.ID, client); err != nil {
		return fmt.Errorf("%w: target Keycloak update failed", ErrUnavailable)
	}
	if err := kc.SyncClientRoles(ctx, client.ID, config.Roles); err != nil {
		return fmt.Errorf("%w: target role synchronization failed", ErrUnavailable)
	}
	if err := kc.CreateDefaultProtocolMappers(ctx, client.ID); err != nil {
		return fmt.Errorf("%w: target protocol-mapper synchronization failed", ErrUnavailable)
	}
	if err := syncManagedScopes(ctx, kc, client.ID, config.ClientScopes); err != nil {
		return fmt.Errorf("%w: target client-scope synchronization failed", ErrUnavailable)
	}
	if err := s.repo.MarkDeployed(ctx, deployment.ID, client.ID, subject); err != nil {
		return err
	}
	return nil
}

func (s *Service) CheckDrift(ctx context.Context, deploymentID, subject string) (DriftRun, error) {
	deployment, err := s.repo.GetDeployment(ctx, deploymentID)
	if err != nil {
		return DriftRun{}, err
	}
	jobID, err := s.createJob(ctx, deployment, deployment.SnapshotID, "check_configuration_drift")
	if err != nil {
		return DriftRun{}, err
	}
	runID, err := s.repo.CreateDriftRun(ctx, deploymentID, subject)
	if err != nil {
		s.failJob(ctx, jobID, deploymentID, err)
		return DriftRun{}, err
	}
	snapshot, err := s.repo.GetSnapshot(ctx, deployment.SnapshotID)
	if err != nil {
		return s.finishDriftError(ctx, runID, deploymentID, jobID, err)
	}
	kc, _, err := s.connectionClient(ctx, deployment.RealmConnectionID, false)
	if err != nil {
		return s.finishDriftError(ctx, runID, deploymentID, jobID, err)
	}
	client, err := kc.GetClient(ctx, deployment.KeycloakClientUUID)
	if err != nil {
		return s.finishDriftError(ctx, runID, deploymentID, jobID, err)
	}
	roles, err := kc.ListClientRoles(ctx, client.ID)
	if err != nil {
		return s.finishDriftError(ctx, runID, deploymentID, jobID, err)
	}
	mappers, err := kc.ListClientProtocolMappers(ctx, client.ID)
	if err != nil {
		return s.finishDriftError(ctx, runID, deploymentID, jobID, err)
	}
	scopeAssignments, err := kc.GetClientScopeAssignments(ctx, client.ID)
	if err != nil {
		return s.finishDriftError(ctx, runID, deploymentID, jobID, err)
	}
	desired, err := applyOverrides(snapshot.Configuration, deployment.Overrides)
	if err != nil {
		return s.finishDriftError(ctx, runID, deploymentID, jobID, err)
	}
	actualRoles := make([]string, 0, len(roles))
	for _, role := range roles {
		actualRoles = append(actualRoles, role.Name)
	}
	actual := SnapshotConfiguration{
		ClientID: client.ClientID, Name: client.Name, Description: client.Description, Enabled: client.Enabled,
		AppType: inferAppType(client), RedirectURIs: sorted(client.RedirectURIs),
		WebOrigins: sorted(client.WebOrigins), Roles: sorted(actualRoles),
		ProtocolMappers: managedMapperConfiguration(mappers, desired.ProtocolMappers),
		ClientScopes:    managedScopeConfiguration(scopeAssignments, desired.ClientScopes),
	}
	desiredHash, desiredJSON, _ := canonicalHash(desired)
	actualHash, actualJSON, _ := canonicalHash(actual)
	findings := compareJSON(desiredJSON, actualJSON)
	status := "in_sync"
	if len(findings) > 0 {
		status = "drifted"
	}
	if err := s.repo.CompleteDriftRun(ctx, runID, deploymentID, status, desiredHash, actualHash, "", findings); err != nil {
		s.failJob(ctx, jobID, deploymentID, err)
		return DriftRun{}, err
	}
	s.succeedJob(ctx, jobID)
	if status == "drifted" && s.notifications != nil {
		_ = s.notifications.SendAdmins(ctx, notifications.Message{Type: "drift_detected", Title: "Configuration drift detected",
			Body: deployment.ApplicationName + " differs in " + deployment.Environment, ResourceType: "deployment",
			ResourceID: deployment.ID, ApplicationID: deployment.ApplicationID, DeduplicationKey: "drift:" + runID})
	}
	runs, err := s.repo.ListDriftRuns(ctx, deploymentID)
	if err != nil || len(runs) == 0 {
		return DriftRun{}, err
	}
	return runs[0], nil
}

func (s *Service) finishDriftError(ctx context.Context, runID, deploymentID, jobID string, cause error) (DriftRun, error) {
	_ = s.repo.CompleteDriftRun(ctx, runID, deploymentID, "error", "", "", "target realm check failed", nil)
	s.failJob(ctx, jobID, deploymentID, cause)
	if s.notifications != nil {
		deployment, _ := s.repo.GetDeployment(ctx, deploymentID)
		_ = s.notifications.SendAdmins(ctx, notifications.Message{
			Type: "drift_check_failed", Title: "Configuration drift check failed",
			Body: "The target realm could not be checked.", ResourceType: "drift_run",
			ResourceID: runID, ApplicationID: deployment.ApplicationID,
			DeduplicationKey: "drift-error:" + runID,
		})
	}
	return DriftRun{}, fmt.Errorf("%w: drift check failed", ErrUnavailable)
}

func (s *Service) ListDriftRuns(ctx context.Context, deploymentID string) ([]DriftRun, error) {
	return s.repo.ListDriftRuns(ctx, deploymentID)
}

func (s *Service) RunDriftScheduler(ctx context.Context, interval time.Duration, concurrency int) {
	if interval <= 0 {
		return
	}
	if concurrency < 1 {
		concurrency = 1
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runScheduledDriftChecks(ctx, concurrency)
		}
	}
}

func (s *Service) RunSecretExpiryScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rows, err := s.db.Query(ctx, `UPDATE secret_deliveries sd
				SET secret_ciphertext=NULL,secret_nonce=NULL
				FROM application_deployments d
				WHERE sd.deployment_id=d.id AND sd.consumed_at IS NULL
				  AND sd.expires_at<=NOW() AND sd.secret_ciphertext IS NOT NULL
				RETURNING sd.id,sd.recipient_subject,d.application_id`)
			if err != nil {
				continue
			}
			for rows.Next() {
				var id, recipient, appID string
				if rows.Scan(&id, &recipient, &appID) == nil && s.notifications != nil {
					_ = s.notifications.Send(ctx, notifications.Message{
						RecipientSubject: recipient, Type: "secret_rotation_expired",
						Title:        "Rotated client secret expired",
						Body:         "The one-time secret was not retrieved before expiry.",
						ResourceType: "secret_delivery", ResourceID: id, ApplicationID: appID,
						DeduplicationKey: "secret-expired:" + id,
					})
				}
			}
			rows.Close()
		}
	}
}

func (s *Service) runScheduledDriftChecks(ctx context.Context, concurrency int) {
	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return
	}
	defer conn.Release()
	var locked bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock(8042026)`).Scan(&locked); err != nil || !locked {
		return
	}
	defer conn.Exec(context.Background(), `SELECT pg_advisory_unlock(8042026)`)
	deployments, err := s.repo.ListDeployments(ctx, "")
	if err != nil {
		return
	}
	sem := make(chan struct{}, concurrency)
	var done = make(chan struct{}, len(deployments))
	started := 0
	for _, deployment := range deployments {
		if deployment.Status != "deployed" {
			continue
		}
		started++
		sem <- struct{}{}
		go func(id string) {
			defer func() {
				<-sem
				done <- struct{}{}
			}()
			_, _ = s.CheckDrift(ctx, id, "scheduler")
		}(deployment.ID)
	}
	for i := 0; i < started; i++ {
		<-done
	}
}

func (s *Service) Reconcile(ctx context.Context, deploymentID, subject string, approved bool) (Deployment, string, error) {
	deployment, err := s.repo.GetDeployment(ctx, deploymentID)
	if err != nil {
		return Deployment{}, "", err
	}
	if !approved {
		return Deployment{}, "", ErrForbidden
	}
	snapshot, err := s.repo.GetSnapshot(ctx, deployment.SnapshotID)
	if err != nil {
		return Deployment{}, "", err
	}
	jobID, err := s.createJob(ctx, deployment, snapshot.ID, "reconcile_drift")
	if err != nil {
		return Deployment{}, "", err
	}
	if err := s.applySnapshot(ctx, deployment, snapshot, subject); err != nil {
		s.failJob(ctx, jobID, deployment.ID, err)
		return deployment, jobID, err
	}
	_, verifyErr := s.CheckDrift(ctx, deploymentID, subject)
	if verifyErr != nil {
		s.failJob(ctx, jobID, deployment.ID, verifyErr)
		return deployment, jobID, verifyErr
	}
	s.succeedJob(ctx, jobID)
	if s.notifications != nil {
		_ = s.notifications.SendAdmins(ctx, notifications.Message{
			Type: "drift_resolved", Title: "Configuration drift resolved",
			Body:         deployment.ApplicationName + " is in sync in " + deployment.Environment,
			ResourceType: "deployment", ResourceID: deployment.ID, ApplicationID: deployment.ApplicationID,
			DeduplicationKey: "drift-resolved:" + deployment.ID + ":" + snapshot.ID,
		})
	}
	updated, err := s.repo.GetDeployment(ctx, deploymentID)
	return updated, jobID, err
}

func (s *Service) RotateSecret(ctx context.Context, deploymentID, recipient string, approved bool) (SecretDelivery, string, error) {
	if !approved {
		return SecretDelivery{}, "", ErrForbidden
	}
	deployment, err := s.repo.GetDeployment(ctx, deploymentID)
	if err != nil {
		return SecretDelivery{}, "", err
	}
	snapshot, err := s.repo.GetSnapshot(ctx, deployment.SnapshotID)
	if err != nil {
		return SecretDelivery{}, "", err
	}
	if snapshot.Configuration.AppType == "frontend" || snapshot.Configuration.AppType == "mobile" {
		return SecretDelivery{}, "", fmt.Errorf("%w: public clients do not have rotatable secrets", ErrValidation)
	}
	kc, _, err := s.connectionClient(ctx, deployment.RealmConnectionID, false)
	if err != nil {
		return SecretDelivery{}, "", err
	}
	jobID, err := s.createJob(ctx, deployment, snapshot.ID, "rotate_client_secret")
	if err != nil {
		return SecretDelivery{}, "", err
	}
	secret, err := kc.RotateClientSecret(ctx, deployment.KeycloakClientUUID)
	if err != nil {
		s.failJob(ctx, jobID, deployment.ID, err)
		return SecretDelivery{}, jobID, fmt.Errorf("%w: secret rotation failed; status is unknown and was not retried", ErrUnavailable)
	}
	ciphertext, nonce, version, err := s.keyring.Encrypt([]byte(secret), "secret-delivery:"+deployment.ID+":"+recipient)
	if err != nil {
		s.failJob(ctx, jobID, deployment.ID, err)
		return SecretDelivery{}, jobID, err
	}
	delivery, err := s.repo.CreateSecretDelivery(ctx, deployment.ID, recipient, ciphertext, nonce, version, time.Now().Add(s.deliveryTTL))
	if err != nil {
		s.failJob(ctx, jobID, deployment.ID, err)
		return SecretDelivery{}, jobID, err
	}
	s.succeedJob(ctx, jobID)
	if s.notifications != nil {
		_ = s.notifications.Send(ctx, notifications.Message{RecipientSubject: recipient, Type: "secret_rotation_ready",
			Title: "Rotated client secret ready", Body: "Retrieve the secret before it expires.",
			ResourceType: "secret_delivery", ResourceID: delivery.ID, ApplicationID: deployment.ApplicationID,
			DeduplicationKey: "secret-ready:" + delivery.ID})
	}
	return delivery, jobID, nil
}

func (s *Service) ConsumeSecret(ctx context.Context, id, recipient string) (string, SecretDelivery, error) {
	ciphertext, nonce, version, delivery, err := s.repo.ConsumeSecret(ctx, id, recipient)
	if err != nil {
		return "", delivery, err
	}
	secret, err := s.keyring.Decrypt(ciphertext, nonce, version, "secret-delivery:"+delivery.DeploymentID+":"+recipient)
	if err != nil {
		return "", delivery, err
	}
	if s.notifications != nil {
		deployment, _ := s.repo.GetDeployment(ctx, delivery.DeploymentID)
		_ = s.notifications.Send(ctx, notifications.Message{
			RecipientSubject: recipient, Type: "secret_rotation_consumed",
			Title: "Rotated client secret consumed", Body: "The one-time secret was retrieved and deleted.",
			ResourceType: "secret_delivery", ResourceID: delivery.ID, ApplicationID: deployment.ApplicationID,
			DeduplicationKey: "secret-consumed:" + delivery.ID,
		})
	}
	return string(secret), delivery, nil
}

func (s *Service) ExecuteApproved(ctx context.Context, action, appID string, payload json.RawMessage, subject string) (string, error) {
	switch action {
	case "deploy_application":
		var input CreateDeploymentRequest
		if err := json.Unmarshal(payload, &input); err != nil {
			return "", fmt.Errorf("%w: invalid deployment payload", ErrValidation)
		}
		_, jobID, err := s.Deploy(ctx, appID, input, subject, true)
		return jobID, err
	case "promote_application":
		var input PromotionRequest
		if err := json.Unmarshal(payload, &input); err != nil {
			return "", fmt.Errorf("%w: invalid promotion payload", ErrValidation)
		}
		_, jobID, err := s.Promote(ctx, appID, input, subject, true)
		return jobID, err
	case "rollback_deployment", "reconcile_drift", "rotate_client_secret":
		var input struct {
			DeploymentID string `json:"deployment_id"`
		}
		if err := json.Unmarshal(payload, &input); err != nil || input.DeploymentID == "" {
			return "", fmt.Errorf("%w: deployment_id is required", ErrValidation)
		}
		if action == "rollback_deployment" {
			_, jobID, err := s.Rollback(ctx, input.DeploymentID, subject, true)
			return jobID, err
		}
		if action == "reconcile_drift" {
			_, jobID, err := s.Reconcile(ctx, input.DeploymentID, subject, true)
			return jobID, err
		}
		_, jobID, err := s.RotateSecret(ctx, input.DeploymentID, subject, true)
		return jobID, err
	default:
		return "", ErrConflict
	}
}

func (s *Service) createJob(ctx context.Context, deployment Deployment, snapshotID, action string) (string, error) {
	var id string
	err := s.db.QueryRow(ctx, `INSERT INTO provisioning_jobs(application_id,action,status,started_at,
		environment_id,realm_connection_id,deployment_id,snapshot_id)
		VALUES($1,$2,'running',NOW(),$3,$4,$5,$6) RETURNING id`, deployment.ApplicationID, action,
		deployment.EnvironmentID, deployment.RealmConnectionID, deployment.ID, snapshotID).Scan(&id)
	return id, err
}

func (s *Service) succeedJob(ctx context.Context, id string) {
	_, _ = s.db.Exec(ctx, `UPDATE provisioning_jobs SET status='succeeded',completed_at=NOW(),updated_at=NOW() WHERE id=$1`, id)
}

func (s *Service) failJob(ctx context.Context, jobID, deploymentID string, err error) {
	_, _ = s.db.Exec(ctx, `UPDATE provisioning_jobs SET status='failed',error_message=$2,completed_at=NOW(),updated_at=NOW() WHERE id=$1`, jobID, safeError(err))
	_ = deploymentID
}

func toKeycloakRequest(config SnapshotConfiguration) keycloak.CreateClientRequest {
	public := config.AppType == "frontend" || config.AppType == "mobile"
	req := keycloak.CreateClientRequest{ClientID: config.ClientID, Name: config.Name, Description: config.Description,
		Protocol: "openid-connect", PublicClient: public, StandardFlowEnabled: public,
		ServiceAccountsEnabled: config.AppType == "machine_to_machine", RedirectURIs: config.RedirectURIs,
		WebOrigins: config.WebOrigins, Enabled: config.Enabled}
	if public {
		req.Attributes = map[string]string{"pkce.code.challenge.method": "S256"}
	}
	return req
}

func inferAppType(client keycloak.KeycloakClient) string {
	if client.PublicClient {
		return "frontend"
	}
	if client.ServiceAccountsEnabled {
		return "machine_to_machine"
	}
	return "backend"
}

func applyOverrides(config SnapshotConfiguration, raw json.RawMessage) (SnapshotConfiguration, error) {
	if len(raw) == 0 || string(raw) == "{}" {
		return config, nil
	}
	var overrides struct {
		RedirectURIs *[]string `json:"redirect_uris"`
		WebOrigins   *[]string `json:"web_origins"`
		Enabled      *bool     `json:"enabled"`
	}
	if err := json.Unmarshal(raw, &overrides); err != nil {
		return config, fmt.Errorf("%w: invalid overrides", ErrValidation)
	}
	if overrides.RedirectURIs != nil {
		config.RedirectURIs = sorted(*overrides.RedirectURIs)
	}
	if overrides.WebOrigins != nil {
		config.WebOrigins = sorted(*overrides.WebOrigins)
	}
	if overrides.Enabled != nil {
		config.Enabled = *overrides.Enabled
	}
	return config, nil
}

func validateOverrides(raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return fmt.Errorf("%w: overrides must be an object", ErrValidation)
	}
	for name := range fields {
		if name != "redirect_uris" && name != "web_origins" && name != "enabled" {
			return fmt.Errorf("%w: override %q is not allowed", ErrValidation, name)
		}
	}
	return nil
}

func sorted(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	if result == nil {
		return []string{}
	}
	return result
}

func defaultMapperConfiguration() []MapperConfiguration {
	defaults := keycloak.DefaultProtocolMappers()
	result := make([]MapperConfiguration, 0, len(defaults))
	for _, mapper := range defaults {
		result = append(result, MapperConfiguration{
			Name: mapper.Name, Protocol: mapper.Protocol,
			ProtocolMapper: mapper.ProtocolMapper, Config: mapper.Config,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func managedMapperConfiguration(actual []keycloak.ProtocolMapper, desired []MapperConfiguration) []MapperConfiguration {
	names := map[string]bool{}
	for _, mapper := range desired {
		names[mapper.Name] = true
	}
	result := []MapperConfiguration{}
	for _, mapper := range actual {
		if names[mapper.Name] {
			result = append(result, MapperConfiguration{
				Name: mapper.Name, Protocol: mapper.Protocol,
				ProtocolMapper: mapper.ProtocolMapper, Config: mapper.Config,
			})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func defaultScopeConfiguration() []ScopeConfiguration {
	return []ScopeConfiguration{
		{Name: "email", Type: "default"},
		{Name: "profile", Type: "default"},
		{Name: "roles", Type: "default"},
		{Name: "web-origins", Type: "default"},
	}
}

func syncManagedScopes(ctx context.Context, kc *keycloak.Client, clientID string, desired []ScopeConfiguration) error {
	if len(desired) == 0 {
		return nil
	}
	assignments, err := kc.GetClientScopeAssignments(ctx, clientID)
	if err != nil {
		return err
	}
	all := append(append(append([]keycloak.ClientScope{}, assignments.Default...), assignments.Optional...), assignments.Available...)
	for _, wanted := range desired {
		for _, scope := range all {
			if scope.Name != wanted.Name {
				continue
			}
			if wanted.Type == "optional" {
				if err := kc.AssignOptionalClientScope(ctx, clientID, scope.ID); err != nil {
					return err
				}
			} else if err := kc.AssignDefaultClientScope(ctx, clientID, scope.ID); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func managedScopeConfiguration(actual keycloak.ClientScopeAssignments, desired []ScopeConfiguration) []ScopeConfiguration {
	wanted := map[string]string{}
	for _, scope := range desired {
		wanted[scope.Name] = scope.Type
	}
	result := []ScopeConfiguration{}
	for _, group := range []struct {
		items []keycloak.ClientScope
		kind  string
	}{{actual.Default, "default"}, {actual.Optional, "optional"}} {
		for _, scope := range group.items {
			if wanted[scope.Name] == group.kind {
				result = append(result, ScopeConfiguration{Name: scope.Name, Type: group.kind})
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func canonicalHash(value any) (string, json.RawMessage, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), payload, nil
}

func compareJSON(desiredRaw, actualRaw []byte) []DriftFinding {
	var desired, actual map[string]any
	_ = json.Unmarshal(desiredRaw, &desired)
	_ = json.Unmarshal(actualRaw, &actual)
	keys := make([]string, 0, len(desired))
	for key := range desired {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	findings := []DriftFinding{}
	for _, key := range keys {
		d, _ := json.Marshal(desired[key])
		a, ok := actual[key]
		if !ok {
			findings = append(findings, DriftFinding{Path: "/" + key, ChangeType: "missing", DesiredValue: d})
			continue
		}
		actualValue, _ := json.Marshal(a)
		if string(d) != string(actualValue) {
			findings = append(findings, DriftFinding{Path: "/" + key, ChangeType: "changed", DesiredValue: d, ActualValue: actualValue})
		}
	}
	return findings
}

func safeError(err error) string {
	switch {
	case errors.Is(err, ErrUnavailable):
		return "target realm operation failed"
	case errors.Is(err, ErrValidation), errors.Is(err, ErrConflict):
		return err.Error()
	default:
		return "workflow failed"
	}
}
