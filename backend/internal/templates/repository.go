package templates

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

func (r *Repository) List(ctx context.Context) ([]Template, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, COALESCE(description, ''), app_type,
		       default_roles, default_redirect_uris, default_web_origins,
		       default_scopes, default_mappers, client_config,
		       created_at, updated_at
		FROM onboarding_templates
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Template{}

	for rows.Next() {
		var item Template

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.AppType,
			&item.DefaultRoles,
			&item.DefaultRedirectURIs,
			&item.DefaultWebOrigins,
			&item.DefaultScopes,
			&item.DefaultMappers,
			&item.ClientConfig,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id string) (Template, error) {
	var item Template

	err := r.db.QueryRow(ctx, `
		SELECT id, name, COALESCE(description, ''), app_type,
		       default_roles, default_redirect_uris, default_web_origins,
		       default_scopes, default_mappers, client_config,
		       created_at, updated_at
		FROM onboarding_templates
		WHERE id = $1
	`, id).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.AppType,
		&item.DefaultRoles,
		&item.DefaultRedirectURIs,
		&item.DefaultWebOrigins,
		&item.DefaultScopes,
		&item.DefaultMappers,
		&item.ClientConfig,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (r *Repository) SeedDefaults(ctx context.Context) error {
	for _, item := range DefaultTemplates() {
		_, err := r.db.Exec(ctx, `
			INSERT INTO onboarding_templates (
				name, description, app_type,
				default_roles, default_redirect_uris, default_web_origins,
				default_scopes, default_mappers, client_config
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (name) DO NOTHING
		`,
			item.Name,
			item.Description,
			item.AppType,
			item.DefaultRoles,
			item.DefaultRedirectURIs,
			item.DefaultWebOrigins,
			item.DefaultScopes,
			item.DefaultMappers,
			item.ClientConfig,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
