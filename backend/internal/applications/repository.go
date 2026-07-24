package applications

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Application, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, COALESCE(description, ''), app_type,
		       COALESCE(owner_name, ''), COALESCE(owner_email, ''),
		       status, source, enabled, config_version,
		       COALESCE(keycloak_client_uuid, ''),
		       COALESCE(keycloak_client_id, ''),
		       provisioned_at, created_at, updated_at
		FROM applications
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []Application{}

	for rows.Next() {
		var app Application

		if err := rows.Scan(
			&app.ID,
			&app.Name,
			&app.Slug,
			&app.Description,
			&app.AppType,
			&app.OwnerName,
			&app.OwnerEmail,
			&app.Status,
			&app.Source,
			&app.Enabled,
			&app.ConfigVersion,
			&app.KeycloakClientUUID,
			&app.KeycloakClientID,
			&app.ProvisionedAt,
			&app.CreatedAt,
			&app.UpdatedAt,
		); err != nil {
			return nil, err
		}

		apps = append(apps, app)
	}

	return apps, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id string) (Application, error) {
	var app Application

	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(description, ''), app_type,
		       COALESCE(owner_name, ''), COALESCE(owner_email, ''),
		       status, source, enabled, config_version,
		       COALESCE(keycloak_client_uuid, ''),
		       COALESCE(keycloak_client_id, ''),
		       COALESCE(keycloak_client_secret, ''),
		       provisioned_at, created_at, updated_at
		FROM applications
		WHERE id = $1
	`, id).Scan(
		&app.ID,
		&app.Name,
		&app.Slug,
		&app.Description,
		&app.AppType,
		&app.OwnerName,
		&app.OwnerEmail,
		&app.Status,
		&app.Source,
		&app.Enabled,
		&app.ConfigVersion,
		&app.KeycloakClientUUID,
		&app.KeycloakClientID,
		&app.KeycloakClientSecret,
		&app.ProvisionedAt,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Application{}, ErrNotFound
		}
		return Application{}, err
	}

	app.RedirectURIs, _ = r.getRedirectURIs(ctx, id)
	app.WebOrigins, _ = r.getWebOrigins(ctx, id)
	app.Roles, _ = r.getRoles(ctx, id)

	return app, nil
}

func (r *Repository) Update(ctx context.Context, id string, req UpdateApplicationRequest) (Application, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Application{}, err
	}
	defer tx.Rollback(ctx)

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	tag, err := tx.Exec(ctx, `
		UPDATE applications
		SET name = $2, slug = $3, description = $4, app_type = $5,
		    owner_name = $6, owner_email = $7, enabled = $8,
		    keycloak_client_id = CASE WHEN keycloak_client_uuid IS NOT NULL THEN $3 ELSE keycloak_client_id END,
		    config_version = config_version + 1,
		    updated_at = NOW()
		WHERE id = $1
	`, id, req.Name, req.Slug, req.Description, req.AppType, req.OwnerName, req.OwnerEmail, enabled)
	if err != nil {
		if isUniqueViolation(err) {
			return Application{}, fmt.Errorf("%w: slug or Keycloak client is already managed", ErrConflict)
		}
		return Application{}, err
	}
	if tag.RowsAffected() == 0 {
		return Application{}, ErrNotFound
	}

	for _, table := range []string{"application_redirect_uris", "application_web_origins", "application_roles"} {
		if _, err := tx.Exec(ctx, "DELETE FROM "+table+" WHERE application_id = $1", id); err != nil {
			return Application{}, err
		}
	}
	if err := insertRelated(ctx, tx, id, req.RedirectURIs, req.WebOrigins, req.Roles); err != nil {
		return Application{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Application{}, err
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) EnsureIdentityAvailable(ctx context.Context, id, slug string) error {
	var conflict bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM applications
			WHERE id <> $1
			  AND (slug = $2 OR keycloak_client_id = $2)
		)
	`, id, slug).Scan(&conflict); err != nil {
		return err
	}
	if conflict {
		return fmt.Errorf("%w: slug or Keycloak client ID %q is already managed", ErrConflict, slug)
	}
	return nil
}

func (r *Repository) Import(ctx context.Context, req CreateApplicationRequest, clientUUID string, clientID string, enabled bool) (Application, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Application{}, err
	}
	defer tx.Rollback(ctx)

	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO applications (
			name, slug, description, app_type, owner_name, owner_email,
			status, source, enabled, keycloak_client_uuid, keycloak_client_id, provisioned_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'provisioned', 'imported', $7, $8, $9, NOW())
		RETURNING id
	`, req.Name, req.Slug, req.Description, req.AppType, req.OwnerName, req.OwnerEmail,
		enabled, clientUUID, clientID).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return Application{}, fmt.Errorf("%w: client ID or Keycloak UUID is already managed", ErrConflict)
		}
		return Application{}, err
	}
	if err := insertRelated(ctx, tx, id, req.RedirectURIs, req.WebOrigins, req.Roles); err != nil {
		return Application{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Application{}, err
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) LinkedClientUUIDs(ctx context.Context) (map[string]bool, error) {
	rows, err := r.db.Query(ctx, `SELECT keycloak_client_uuid FROM applications WHERE keycloak_client_uuid IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = true
	}
	return result, rows.Err()
}

func (r *Repository) Create(ctx context.Context, req CreateApplicationRequest) (Application, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Application{}, err
	}
	defer tx.Rollback(ctx)

	var id string

	err = tx.QueryRow(ctx, `
		INSERT INTO applications (
			name, slug, description, app_type, owner_name, owner_email
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		req.Name,
		req.Slug,
		req.Description,
		req.AppType,
		req.OwnerName,
		req.OwnerEmail,
	).Scan(&id)

	if err != nil {
		return Application{}, err
	}

	for _, uri := range req.RedirectURIs {
		if uri == "" {
			continue
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO application_redirect_uris (application_id, redirect_uri)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, id, uri)

		if err != nil {
			return Application{}, err
		}
	}

	for _, origin := range req.WebOrigins {
		if origin == "" {
			continue
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO application_web_origins (application_id, web_origin)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, id, origin)

		if err != nil {
			return Application{}, err
		}
	}

	for _, role := range req.Roles {
		if role == "" {
			continue
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO application_roles (application_id, role_name)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, id, role)

		if err != nil {
			return Application{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Application{}, err
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE applications
		SET status = $2,
		    updated_at = NOW()
		WHERE id = $1
	`, id, status)

	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM applications WHERE id = $1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (r *Repository) getRedirectURIs(ctx context.Context, appID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT redirect_uri
		FROM application_redirect_uris
		WHERE application_id = $1
		ORDER BY redirect_uri
	`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, rows.Err()
}

type dbExecer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func insertRelated(ctx context.Context, tx dbExecer, id string, redirects, origins, roles []string) error {
	for _, value := range redirects {
		if value == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO application_redirect_uris (application_id, redirect_uri) VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, value); err != nil {
			return err
		}
	}
	for _, value := range origins {
		if value == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO application_web_origins (application_id, web_origin) VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, value); err != nil {
			return err
		}
	}
	for _, value := range roles {
		if value == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO application_roles (application_id, role_name) VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, value); err != nil {
			return err
		}
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *Repository) getWebOrigins(ctx context.Context, appID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT web_origin
		FROM application_web_origins
		WHERE application_id = $1
		ORDER BY web_origin
	`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, rows.Err()
}

func (r *Repository) getRoles(ctx context.Context, appID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT role_name
		FROM application_roles
		WHERE application_id = $1
		ORDER BY role_name
	`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, rows.Err()
}

func (r *Repository) MarkProvisioned(
	ctx context.Context,
	id string,
	clientUUID string,
	clientID string,
	clientSecret string,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE applications
		SET status = 'provisioned',
		    keycloak_client_uuid = $2,
		    keycloak_client_id = $3,
		    keycloak_client_secret = $4,
		    provisioned_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`, id, clientUUID, clientID, clientSecret)

	return err
}
