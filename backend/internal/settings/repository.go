package settings

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

func (r *Repository) Get(ctx context.Context) (Settings, error) {
	var item Settings

	err := r.db.QueryRow(ctx, `
		SELECT id,
		       keycloak_base_url,
		       keycloak_realm,
		       keycloak_admin_client_id,
		       COALESCE(keycloak_admin_client_secret, ''),
		       created_at,
		       updated_at
		FROM settings
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(
		&item.ID,
		&item.KeycloakBaseURL,
		&item.KeycloakRealm,
		&item.KeycloakAdminClientID,
		&item.KeycloakAdminClientSecret,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (r *Repository) Save(ctx context.Context, req SaveSettingsRequest) (Settings, error) {
	var id string

	err := r.db.QueryRow(ctx, `
		INSERT INTO settings (
			keycloak_base_url,
			keycloak_realm,
			keycloak_admin_client_id,
			keycloak_admin_client_secret
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`,
		req.KeycloakBaseURL,
		req.KeycloakRealm,
		req.KeycloakAdminClientID,
		req.KeycloakAdminClientSecret,
	).Scan(&id)

	if err != nil {
		return Settings{}, err
	}

	return r.Get(ctx)
}
