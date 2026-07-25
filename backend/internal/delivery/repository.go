package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) ListEnvironments(ctx context.Context) ([]Environment, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,slug,promotion_order,protected,enabled,created_at,updated_at
		FROM environments ORDER BY promotion_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Environment{}
	for rows.Next() {
		var item Environment
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.PromotionOrder, &item.Protected,
			&item.Enabled, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateEnvironment(ctx context.Context, input CreateEnvironmentRequest) (Environment, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	var id string
	err := r.db.QueryRow(ctx, `INSERT INTO environments(name,slug,promotion_order,protected,enabled)
		VALUES($1,$2,$3,$4,$5) RETURNING id`, input.Name, input.Slug, input.PromotionOrder, input.Protected, enabled).Scan(&id)
	if unique(err) {
		return Environment{}, ErrConflict
	}
	if err != nil {
		return Environment{}, err
	}
	return r.getEnvironment(ctx, id)
}

func (r *Repository) getEnvironment(ctx context.Context, id string) (Environment, error) {
	var item Environment
	err := r.db.QueryRow(ctx, `SELECT id,name,slug,promotion_order,protected,enabled,created_at,updated_at
		FROM environments WHERE id=$1`, id).Scan(&item.ID, &item.Name, &item.Slug, &item.PromotionOrder,
		&item.Protected, &item.Enabled, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Environment{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) UpdateEnvironment(ctx context.Context, id string, input CreateEnvironmentRequest) (Environment, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	tag, err := r.db.Exec(ctx, `UPDATE environments SET name=$2,slug=$3,promotion_order=$4,
		protected=$5,enabled=$6,updated_at=NOW() WHERE id=$1`, id, input.Name, input.Slug,
		input.PromotionOrder, input.Protected, enabled)
	if unique(err) {
		return Environment{}, ErrConflict
	}
	if err != nil {
		return Environment{}, err
	}
	if tag.RowsAffected() == 0 {
		return Environment{}, ErrNotFound
	}
	return r.getEnvironment(ctx, id)
}

func (r *Repository) DeleteEnvironment(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM environments WHERE id=$1`, id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListConnections(ctx context.Context) ([]RealmConnection, error) {
	rows, err := r.db.Query(ctx, connectionSelect+` ORDER BY e.promotion_order,rc.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []RealmConnection{}
	for rows.Next() {
		item, _, err := scanConnection(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateConnection(ctx context.Context, input CreateConnectionRequest, ciphertext, nonce []byte, version string) (RealmConnection, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	var id string
	err := r.db.QueryRow(ctx, `INSERT INTO realm_connections(environment_id,name,base_url,realm,
		admin_client_id,admin_secret_ciphertext,admin_secret_nonce,encryption_key_version,enabled)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, input.EnvironmentID, input.Name,
		input.BaseURL, input.Realm, input.AdminClientID, ciphertext, nonce, version, enabled).Scan(&id)
	if unique(err) {
		return RealmConnection{}, ErrConflict
	}
	if err != nil {
		return RealmConnection{}, err
	}
	credential, err := r.GetConnectionCredential(ctx, id)
	return credential.RealmConnection, err
}

func (r *Repository) GetConnectionCredential(ctx context.Context, id string) (connectionCredential, error) {
	item, credential, err := scanConnection(r.db.QueryRow(ctx, connectionSelect+` WHERE rc.id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return connectionCredential{}, ErrNotFound
	}
	if err != nil {
		return connectionCredential{}, err
	}
	credential.RealmConnection = item
	return credential, nil
}

func (r *Repository) SaveMigratedCredential(ctx context.Context, id string, ciphertext, nonce []byte, version string) error {
	tag, err := r.db.Exec(ctx, `UPDATE realm_connections SET admin_secret_ciphertext=$2,
		admin_secret_nonce=$3,encryption_key_version=$4,legacy_admin_secret=NULL,updated_at=NOW()
		WHERE id=$1 AND legacy_admin_secret IS NOT NULL`, id, ciphertext, nonce, version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrConflict
	}
	return nil
}

func (r *Repository) UpdateConnection(ctx context.Context, id string, input UpdateConnectionRequest, encrypted bool, ciphertext, nonce []byte, version string) (RealmConnection, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	tag, err := r.db.Exec(ctx, `UPDATE realm_connections SET environment_id=$2,name=$3,base_url=$4,
		realm=$5,admin_client_id=$6,enabled=$7,
		admin_secret_ciphertext=CASE WHEN $8 THEN $9 ELSE admin_secret_ciphertext END,
		admin_secret_nonce=CASE WHEN $8 THEN $10 ELSE admin_secret_nonce END,
		encryption_key_version=CASE WHEN $8 THEN $11 ELSE encryption_key_version END,
		legacy_admin_secret=CASE WHEN $8 THEN NULL ELSE legacy_admin_secret END,
		updated_at=NOW() WHERE id=$1`, id, input.EnvironmentID, input.Name, input.BaseURL,
		input.Realm, input.AdminClientID, enabled, encrypted, ciphertext, nonce, version)
	if unique(err) {
		return RealmConnection{}, ErrConflict
	}
	if err != nil {
		return RealmConnection{}, err
	}
	if tag.RowsAffected() == 0 {
		return RealmConnection{}, ErrNotFound
	}
	credential, err := r.GetConnectionCredential(ctx, id)
	return credential.RealmConnection, err
}

func (r *Repository) SetConnectionTest(ctx context.Context, id, status, message string) error {
	_, err := r.db.Exec(ctx, `UPDATE realm_connections SET last_tested_at=NOW(),
		last_test_status=$2,last_test_error=NULLIF($3,''),updated_at=NOW() WHERE id=$1`, id, status, message)
	return err
}

func (r *Repository) DisableConnection(ctx context.Context, id string) error {
	var used bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM application_deployments
		WHERE realm_connection_id=$1 AND status IN ('deploying','deployed'))`, id).Scan(&used); err != nil {
		return err
	}
	if used {
		return ErrConflict
	}
	tag, err := r.db.Exec(ctx, `UPDATE realm_connections SET enabled=FALSE,updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) CreateSnapshot(ctx context.Context, appID, subject, username string, configuration SnapshotConfiguration, hash string) (Snapshot, error) {
	payload, err := json.Marshal(configuration)
	if err != nil {
		return Snapshot{}, err
	}
	var id string
	err = r.db.QueryRow(ctx, `INSERT INTO application_snapshots(application_id,version,configuration,
		configuration_hash,created_by_subject,created_by_username)
		VALUES($1,(SELECT COALESCE(MAX(version),0)+1 FROM application_snapshots WHERE application_id=$1),
		$2,$3,$4,$5) ON CONFLICT(application_id,configuration_hash) DO NOTHING
		RETURNING id`, appID, payload, hash, subject, username).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = r.db.QueryRow(ctx, `SELECT id FROM application_snapshots
			WHERE application_id=$1 AND configuration_hash=$2`, appID, hash).Scan(&id)
	}
	if err != nil {
		return Snapshot{}, err
	}
	return r.GetSnapshot(ctx, id)
}

func (r *Repository) GetSnapshot(ctx context.Context, id string) (Snapshot, error) {
	var item Snapshot
	var config []byte
	err := r.db.QueryRow(ctx, `SELECT id,application_id,version,configuration,configuration_hash,
		created_by_subject,created_by_username,created_at FROM application_snapshots WHERE id=$1`, id).
		Scan(&item.ID, &item.ApplicationID, &item.Version, &config, &item.ConfigurationHash,
			&item.CreatedBySubject, &item.CreatedByUsername, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Snapshot{}, ErrNotFound
	}
	if err == nil {
		err = json.Unmarshal(config, &item.Configuration)
	}
	return item, err
}

func (r *Repository) LatestSnapshot(ctx context.Context, appID string) (Snapshot, error) {
	var id string
	err := r.db.QueryRow(ctx, `SELECT id FROM application_snapshots WHERE application_id=$1
		ORDER BY version DESC LIMIT 1`, appID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Snapshot{}, ErrNotFound
	}
	if err != nil {
		return Snapshot{}, err
	}
	return r.GetSnapshot(ctx, id)
}

func (r *Repository) ListSnapshots(ctx context.Context, appID string) ([]Snapshot, error) {
	rows, err := r.db.Query(ctx, `SELECT id FROM application_snapshots WHERE application_id=$1 ORDER BY version DESC`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Snapshot{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		item, err := r.GetSnapshot(ctx, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpsertDeployment(ctx context.Context, appID string, input CreateDeploymentRequest, snapshot Snapshot) (Deployment, error) {
	overrides := input.Overrides
	if len(overrides) == 0 {
		overrides = json.RawMessage(`{}`)
	}
	var id string
	err := r.db.QueryRow(ctx, `INSERT INTO application_deployments(application_id,environment_id,
		realm_connection_id,snapshot_id,keycloak_client_id,overrides,status)
		VALUES($1,$2,$3,$4,$5,$6,'not_deployed')
		ON CONFLICT(application_id,environment_id) DO UPDATE SET realm_connection_id=EXCLUDED.realm_connection_id,
		previous_snapshot_id=application_deployments.snapshot_id,snapshot_id=EXCLUDED.snapshot_id,
		keycloak_client_id=EXCLUDED.keycloak_client_id,overrides=EXCLUDED.overrides,updated_at=NOW()
		RETURNING id`, appID, input.EnvironmentID, input.RealmConnectionID, snapshot.ID,
		snapshot.Configuration.ClientID, overrides).Scan(&id)
	if unique(err) {
		return Deployment{}, ErrConflict
	}
	if err != nil {
		return Deployment{}, err
	}
	return r.GetDeployment(ctx, id)
}

func (r *Repository) GetDeployment(ctx context.Context, id string) (Deployment, error) {
	item, err := scanDeployment(r.db.QueryRow(ctx, deploymentSelect+` WHERE d.id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Deployment{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) GetDeploymentForEnvironment(ctx context.Context, appID, environmentID string) (Deployment, error) {
	item, err := scanDeployment(r.db.QueryRow(ctx, deploymentSelect+` WHERE d.application_id=$1 AND d.environment_id=$2`, appID, environmentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Deployment{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) ListDeployments(ctx context.Context, appID string) ([]Deployment, error) {
	rows, err := r.db.Query(ctx, deploymentSelect+` WHERE ($1='' OR d.application_id=$1::uuid) ORDER BY e.promotion_order,a.name`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Deployment{}
	for rows.Next() {
		item, err := scanDeployment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) MarkDeployed(ctx context.Context, id, uuid, subject string) error {
	_, err := r.db.Exec(ctx, `UPDATE application_deployments SET keycloak_client_uuid=$2,status='deployed',
		drift_status='in_sync',deployed_at=NOW(),deployed_by_subject=$3,last_checked_at=NOW(),updated_at=NOW() WHERE id=$1`,
		id, uuid, subject)
	return err
}

func (r *Repository) PreviousEnvironmentDeployment(ctx context.Context, appID string, order int) (Deployment, error) {
	item, err := scanDeployment(r.db.QueryRow(ctx, deploymentSelect+`
		WHERE d.application_id=$1 AND e.promotion_order < $2 AND d.status='deployed'
		ORDER BY e.promotion_order DESC LIMIT 1`, appID, order))
	if errors.Is(err, pgx.ErrNoRows) {
		return Deployment{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) CreateDriftRun(ctx context.Context, deploymentID, subject string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `INSERT INTO drift_runs(deployment_id,initiated_by_subject)
		VALUES($1,$2) RETURNING id`, deploymentID, subject).Scan(&id)
	return id, err
}

func (r *Repository) CompleteDriftRun(ctx context.Context, runID, deploymentID, status, desiredHash, actualHash, errorMessage string, findings []DriftFinding) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, finding := range findings {
		if _, err := tx.Exec(ctx, `INSERT INTO drift_findings(drift_run_id,path,change_type,desired_value,actual_value)
			VALUES($1,$2,$3,$4,$5)`, runID, finding.Path, finding.ChangeType,
			nullableJSON(finding.DesiredValue), nullableJSON(finding.ActualValue)); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE drift_runs SET status=$2,desired_hash=NULLIF($3,''),
		actual_hash=NULLIF($4,''),error_message=NULLIF($5,''),completed_at=NOW() WHERE id=$1`,
		runID, status, desiredHash, actualHash, errorMessage); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE application_deployments SET drift_status=$2,last_checked_at=NOW(),
		updated_at=NOW() WHERE id=$1`, deploymentID, status); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ListDriftRuns(ctx context.Context, deploymentID string) ([]DriftRun, error) {
	rows, err := r.db.Query(ctx, `SELECT id,deployment_id,status,COALESCE(desired_hash,''),COALESCE(actual_hash,''),
		COALESCE(initiated_by_subject,''),COALESCE(error_message,''),started_at,completed_at
		FROM drift_runs WHERE ($1='' OR deployment_id=$1::uuid) ORDER BY started_at DESC LIMIT 100`, deploymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []DriftRun{}
	for rows.Next() {
		var item DriftRun
		if err := rows.Scan(&item.ID, &item.DeploymentID, &item.Status, &item.DesiredHash, &item.ActualHash,
			&item.InitiatedBySubject, &item.ErrorMessage, &item.StartedAt, &item.CompletedAt); err != nil {
			return nil, err
		}
		item.Findings, _ = r.listFindings(ctx, item.ID)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) listFindings(ctx context.Context, runID string) ([]DriftFinding, error) {
	rows, err := r.db.Query(ctx, `SELECT id,drift_run_id,path,change_type,desired_value,actual_value,created_at
		FROM drift_findings WHERE drift_run_id=$1 ORDER BY path`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []DriftFinding{}
	for rows.Next() {
		var item DriftFinding
		if err := rows.Scan(&item.ID, &item.DriftRunID, &item.Path, &item.ChangeType,
			&item.DesiredValue, &item.ActualValue, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateSecretDelivery(ctx context.Context, deploymentID, recipient string, ciphertext, nonce []byte, version string, expires time.Time) (SecretDelivery, error) {
	var item SecretDelivery
	err := r.db.QueryRow(ctx, `INSERT INTO secret_deliveries(deployment_id,recipient_subject,
		secret_ciphertext,secret_nonce,encryption_key_version,expires_at)
		VALUES($1,$2,$3,$4,$5,$6) RETURNING id,deployment_id,expires_at,consumed_at,created_at`,
		deploymentID, recipient, ciphertext, nonce, version, expires).
		Scan(&item.ID, &item.DeploymentID, &item.ExpiresAt, &item.ConsumedAt, &item.CreatedAt)
	return item, err
}

func (r *Repository) ConsumeSecret(ctx context.Context, id, recipient string) (ciphertext, nonce []byte, version string, delivery SecretDelivery, err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, "", delivery, err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `SELECT id,deployment_id,expires_at,consumed_at,created_at,
		secret_ciphertext,secret_nonce,encryption_key_version FROM secret_deliveries
		WHERE id=$1 AND recipient_subject=$2 FOR UPDATE`, id, recipient).
		Scan(&delivery.ID, &delivery.DeploymentID, &delivery.ExpiresAt, &delivery.ConsumedAt,
			&delivery.CreatedAt, &ciphertext, &nonce, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, "", delivery, ErrNotFound
	}
	if err != nil {
		return nil, nil, "", delivery, err
	}
	if delivery.ConsumedAt != nil {
		return nil, nil, "", delivery, ErrGone
	}
	if !time.Now().Before(delivery.ExpiresAt) {
		if _, updateErr := tx.Exec(ctx, `UPDATE secret_deliveries SET secret_ciphertext=NULL,
			secret_nonce=NULL WHERE id=$1`, id); updateErr != nil {
			return nil, nil, "", delivery, updateErr
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return nil, nil, "", delivery, commitErr
		}
		return nil, nil, "", delivery, ErrGone
	}
	err = tx.QueryRow(ctx, `UPDATE secret_deliveries SET consumed_at=NOW(),secret_ciphertext=NULL,
		secret_nonce=NULL WHERE id=$1 RETURNING consumed_at`, id).Scan(&delivery.ConsumedAt)
	if err != nil {
		return nil, nil, "", delivery, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, nil, "", delivery, err
	}
	return ciphertext, nonce, version, delivery, nil
}

const connectionSelect = `SELECT rc.id,rc.environment_id,e.name,rc.name,rc.base_url,rc.realm,
	rc.admin_client_id,rc.enabled,(rc.admin_secret_ciphertext IS NOT NULL OR rc.legacy_admin_secret IS NOT NULL),
	rc.last_tested_at,COALESCE(rc.last_test_status,''),COALESCE(rc.last_test_error,''),
	rc.created_at,rc.updated_at,rc.admin_secret_ciphertext,rc.admin_secret_nonce,
	COALESCE(rc.encryption_key_version,''),COALESCE(rc.legacy_admin_secret,'')
	FROM realm_connections rc JOIN environments e ON e.id=rc.environment_id`

const deploymentSelect = `SELECT d.id,d.application_id,a.name,d.environment_id,e.name,e.promotion_order,
	e.protected,d.realm_connection_id,rc.realm,COALESCE(d.snapshot_id::text,''),
	COALESCE(d.previous_snapshot_id::text,''),COALESCE(d.keycloak_client_uuid,''),
	d.keycloak_client_id,d.overrides,d.status,d.drift_status,d.deployed_at,
	COALESCE(d.deployed_by_subject,''),d.last_checked_at,d.created_at,d.updated_at
	FROM application_deployments d JOIN applications a ON a.id=d.application_id
	JOIN environments e ON e.id=d.environment_id
	JOIN realm_connections rc ON rc.id=d.realm_connection_id`

type scanner interface{ Scan(...any) error }

func scanConnection(row scanner) (RealmConnection, connectionCredential, error) {
	var item RealmConnection
	var credential connectionCredential
	err := row.Scan(&item.ID, &item.EnvironmentID, &item.Environment, &item.Name, &item.BaseURL,
		&item.Realm, &item.AdminClientID, &item.Enabled, &item.SecretSet, &item.LastTestedAt,
		&item.LastTestStatus, &item.LastTestError, &item.CreatedAt, &item.UpdatedAt,
		&credential.Ciphertext, &credential.Nonce, &credential.KeyVersion, &credential.Legacy)
	return item, credential, err
}

func scanDeployment(row scanner) (Deployment, error) {
	var item Deployment
	err := row.Scan(&item.ID, &item.ApplicationID, &item.ApplicationName, &item.EnvironmentID,
		&item.Environment, &item.PromotionOrder, &item.Protected, &item.RealmConnectionID,
		&item.Realm, &item.SnapshotID, &item.PreviousSnapshotID, &item.KeycloakClientUUID,
		&item.KeycloakClientID, &item.Overrides, &item.Status, &item.DriftStatus, &item.DeployedAt,
		&item.DeployedBySubject, &item.LastCheckedAt, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func nullableJSON(value json.RawMessage) any {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	return value
}

func unique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *Repository) String() string { return fmt.Sprintf("delivery.Repository(%p)", r.db) }
