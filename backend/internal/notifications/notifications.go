package notifications

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Title         string          `json:"title"`
	Message       string          `json:"message"`
	ResourceType  string          `json:"resource_type"`
	ResourceID    string          `json:"resource_id"`
	ApplicationID string          `json:"application_id"`
	ReadAt        *time.Time      `json:"read_at"`
	CreatedAt     time.Time       `json:"created_at"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
}

type Message struct {
	RecipientSubject                                                             string
	Type, Title, Body, ResourceType, ResourceID, ApplicationID, DeduplicationKey string
	Metadata                                                                     any
}

type Sender interface {
	Send(context.Context, Message) error
}

type Service struct{ db *pgxpool.Pool }

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) UpsertUser(ctx context.Context, user auth.User) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO user_profiles(subject,username,email,display_name,effective_role,last_seen_at)
		VALUES($1,$2,$3,$4,$5,NOW())
		ON CONFLICT(subject) DO UPDATE SET username=EXCLUDED.username,email=EXCLUDED.email,
		display_name=EXCLUDED.display_name,effective_role=EXCLUDED.effective_role,last_seen_at=NOW()
	`, user.Subject, user.Username, user.Email, user.DisplayName, user.EffectiveRole())
	return err
}

func (s *Service) Send(ctx context.Context, msg Message) error {
	metadata, _ := json.Marshal(msg.Metadata)
	_, err := s.db.Exec(ctx, `
		INSERT INTO notifications(recipient_subject,type,title,message,resource_type,resource_id,
			application_id,deduplication_key,metadata)
		VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,'')::uuid,NULLIF($8,''),$9)
		ON CONFLICT(recipient_subject,deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING
	`, msg.RecipientSubject, msg.Type, msg.Title, msg.Body, msg.ResourceType, msg.ResourceID,
		msg.ApplicationID, msg.DeduplicationKey, metadata)
	return err
}

func (s *Service) SendAdmins(ctx context.Context, msg Message) error {
	metadata, _ := json.Marshal(msg.Metadata)
	_, err := s.db.Exec(ctx, `
		INSERT INTO notifications(recipient_subject,type,title,message,resource_type,resource_id,
			application_id,deduplication_key,metadata)
		SELECT subject,$1,$2,$3,$4,$5,NULLIF($6,'')::uuid,
		       CASE WHEN $7='' THEN NULL ELSE $7||':'||subject END,$8
		FROM user_profiles WHERE effective_role='admin'
		ON CONFLICT(recipient_subject,deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING
	`, msg.Type, msg.Title, msg.Body, msg.ResourceType, msg.ResourceID, msg.ApplicationID, msg.DeduplicationKey, metadata)
	return err
}

func (s *Service) List(ctx context.Context, subject string) ([]Notification, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id,type,title,message,COALESCE(resource_type,''),COALESCE(resource_id,''),
		       COALESCE(application_id::text,''),read_at,created_at,COALESCE(metadata,'null'::jsonb)
		FROM notifications WHERE recipient_subject=$1 ORDER BY created_at DESC LIMIT 100
	`, subject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Notification{}
	for rows.Next() {
		var item Notification
		if err := rows.Scan(&item.ID, &item.Type, &item.Title, &item.Message,
			&item.ResourceType, &item.ResourceID, &item.ApplicationID, &item.ReadAt, &item.CreatedAt, &item.Metadata); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Service) UnreadCount(ctx context.Context, subject string) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE recipient_subject=$1 AND read_at IS NULL`, subject).Scan(&n)
	return n, err
}
func (s *Service) MarkRead(ctx context.Context, subject, id string) error {
	tag, err := s.db.Exec(ctx, `UPDATE notifications SET read_at=COALESCE(read_at,NOW()) WHERE id=$1 AND recipient_subject=$2`, id, subject)
	if err == nil && tag.RowsAffected() == 0 {
		return context.Canceled
	}
	return err
}
func (s *Service) MarkAllRead(ctx context.Context, subject string) error {
	_, err := s.db.Exec(ctx, `UPDATE notifications SET read_at=NOW() WHERE recipient_subject=$1 AND read_at IS NULL`, subject)
	return err
}
