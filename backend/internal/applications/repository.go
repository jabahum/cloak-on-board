package applications

import (
	"context"

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
		       status, created_at, updated_at
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
		       status, created_at, updated_at
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
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err != nil {
		return Application{}, err
	}

	app.RedirectURIs, _ = r.getRedirectURIs(ctx, id)
	app.WebOrigins, _ = r.getWebOrigins(ctx, id)
	app.Roles, _ = r.getRoles(ctx, id)

	return app, nil
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
	_, err := r.db.Exec(ctx, `DELETE FROM applications WHERE id = $1`, id)
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
