package approvals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jabahum/keycloak-onboarder/backend/internal/applications"
	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
	"github.com/jabahum/keycloak-onboarder/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("approval request not found")
	ErrConflict  = errors.New("invalid approval state")
	ErrForbidden = errors.New("approval action forbidden")
)

type Request struct {
	ID                  string          `json:"id"`
	ApplicationID       string          `json:"application_id"`
	ApplicationName     string          `json:"application_name"`
	Action              string          `json:"action"`
	Status              string          `json:"status"`
	RequestedBySubject  string          `json:"requested_by_subject"`
	RequestedByUsername string          `json:"requested_by_username"`
	RequestPayload      json.RawMessage `json:"request_payload"`
	RequestSummary      string          `json:"request_summary"`
	ApplicationVersion  int             `json:"application_version"`
	ReviewedBySubject   string          `json:"reviewed_by_subject"`
	ReviewedByUsername  string          `json:"reviewed_by_username"`
	ReviewComment       string          `json:"review_comment"`
	RequestedAt         time.Time       `json:"requested_at"`
	DecidedAt           *time.Time      `json:"decided_at"`
	CancelledAt         *time.Time      `json:"cancelled_at"`
	ExecutionJobID      string          `json:"execution_job_id"`
	ExecutionError      string          `json:"execution_error"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type SubmitRequest struct {
	Action  string          `json:"action" binding:"required"`
	Payload json.RawMessage `json:"payload"`
	Summary string          `json:"summary"`
}
type DecisionRequest struct {
	Comment string `json:"comment"`
}

type Executor interface {
	Provision(context.Context, string) (string, error)
	Update(context.Context, string, applications.UpdateApplicationRequest) error
	Delete(context.Context, string, bool) error
}

type Service struct {
	db            *pgxpool.Pool
	apps          *applications.Service
	notifications *notifications.Service
	executor      Executor
}

func NewService(db *pgxpool.Pool, apps *applications.Service, n *notifications.Service, e Executor) *Service {
	return &Service{db: db, apps: apps, notifications: n, executor: e}
}

func (s *Service) Submit(ctx context.Context, appID string, user auth.User, input SubmitRequest) (Request, error) {
	switch input.Action {
	case "provision_application", "update_keycloak_client", "delete_keycloak_client":
	default:
		return Request{}, fmt.Errorf("%w: unsupported action", applications.ErrValidation)
	}
	app, err := s.apps.GetByID(ctx, appID)
	if err != nil {
		return Request{}, err
	}
	if input.Action == "update_keycloak_client" && len(input.Payload) == 0 {
		return Request{}, fmt.Errorf("%w: update payload is required", applications.ErrValidation)
	}
	if len(input.Payload) == 0 {
		input.Payload = json.RawMessage(`{}`)
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Request{}, err
	}
	defer tx.Rollback(ctx)
	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO approval_requests(application_id,action,requested_by_subject,requested_by_username,
			request_payload,request_summary,application_version)
		VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id
	`, appID, input.Action, user.Subject, user.Username, input.Payload, input.Summary, app.ConfigVersion).Scan(&id)
	if err != nil {
		return Request{}, fmt.Errorf("%w: an active request already exists", ErrConflict)
	}
	if _, err = tx.Exec(ctx, `UPDATE applications SET status='pending_approval',updated_at=NOW() WHERE id=$1`, appID); err != nil {
		return Request{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Request{}, err
	}
	_ = s.notifications.SendAdmins(ctx, notifications.Message{Type: "approval_submitted", Title: "Approval requested", Body: user.Username + " requested " + input.Action,
		ResourceType: "approval_request", ResourceID: id, ApplicationID: appID, DeduplicationKey: "approval-submitted:" + id})
	return s.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, user auth.User) ([]Request, error) {
	rows, err := s.db.Query(ctx, approvalSelect+`
		WHERE ($1='admin' OR ar.requested_by_subject=$2)
		ORDER BY ar.requested_at DESC LIMIT 100
	`, user.EffectiveRole(), user.Subject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Request{}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Service) Get(ctx context.Context, id string) (Request, error) {
	row := s.db.QueryRow(ctx, approvalSelect+` WHERE ar.id=$1`, id)
	item, err := scan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Request{}, ErrNotFound
	}
	return item, err
}

func (s *Service) Approve(ctx context.Context, id string, user auth.User, comment string) (Request, error) {
	req, err := s.Get(ctx, id)
	if err != nil {
		return Request{}, err
	}
	if req.RequestedBySubject == user.Subject {
		return Request{}, fmt.Errorf("%w: requesters cannot approve their own request", ErrForbidden)
	}
	if req.Status != "pending" {
		return Request{}, ErrConflict
	}
	app, err := s.apps.GetByID(ctx, req.ApplicationID)
	if err != nil {
		return Request{}, err
	}
	if app.ConfigVersion != req.ApplicationVersion {
		return Request{}, fmt.Errorf("%w: application changed after submission", ErrConflict)
	}
	tag, err := s.db.Exec(ctx, `UPDATE approval_requests SET status='executing',reviewed_by_subject=$2,
		reviewed_by_username=$3,review_comment=$4,decided_at=NOW(),updated_at=NOW()
		WHERE id=$1 AND status='pending'`, id, user.Subject, user.Username, comment)
	if err != nil {
		return Request{}, err
	}
	if tag.RowsAffected() != 1 {
		return Request{}, ErrConflict
	}
	_ = s.apps.UpdateStatus(ctx, req.ApplicationID, "approved")
	jobID, execErr := s.execute(ctx, req)
	status := "succeeded"
	errorMessage := ""
	if execErr != nil {
		status = "failed"
		errorMessage = execErr.Error()
		_ = s.apps.UpdateStatus(ctx, req.ApplicationID, "failed")
	}
	_, _ = s.db.Exec(ctx, `UPDATE approval_requests SET status=$2,execution_job_id=NULLIF($3,'')::uuid,
		execution_error=$4,updated_at=NOW() WHERE id=$1`, id, status, jobID, errorMessage)
	notificationApplicationID := req.ApplicationID
	if req.Action == "delete_keycloak_client" && status == "succeeded" {
		notificationApplicationID = ""
	}
	_ = s.notifications.Send(ctx, notifications.Message{RecipientSubject: req.RequestedBySubject, Type: "approval_" + status,
		Title: "Approval " + status, Body: req.Action + " " + status, ResourceType: "approval_request", ResourceID: id,
		ApplicationID: notificationApplicationID, DeduplicationKey: "approval-result:" + id + ":" + status})
	if execErr != nil {
		return s.Get(ctx, id)
	}
	return s.Get(ctx, id)
}

func (s *Service) Reject(ctx context.Context, id string, user auth.User, comment string) (Request, error) {
	if comment == "" {
		return Request{}, fmt.Errorf("%w: rejection comment is required", applications.ErrValidation)
	}
	req, err := s.Get(ctx, id)
	if err != nil {
		return Request{}, err
	}
	if req.RequestedBySubject == user.Subject {
		return Request{}, ErrForbidden
	}
	tag, err := s.db.Exec(ctx, `UPDATE approval_requests SET status='rejected',reviewed_by_subject=$2,
		reviewed_by_username=$3,review_comment=$4,decided_at=NOW(),updated_at=NOW()
		WHERE id=$1 AND status='pending'`, id, user.Subject, user.Username, comment)
	if err != nil {
		return Request{}, err
	}
	if tag.RowsAffected() != 1 {
		return Request{}, ErrConflict
	}
	_ = s.apps.UpdateStatus(ctx, req.ApplicationID, "rejected")
	_ = s.notifications.Send(ctx, notifications.Message{RecipientSubject: req.RequestedBySubject, Type: "approval_rejected",
		Title: "Approval rejected", Body: comment, ResourceType: "approval_request", ResourceID: id, ApplicationID: req.ApplicationID, DeduplicationKey: "approval-rejected:" + id})
	return s.Get(ctx, id)
}
func (s *Service) Cancel(ctx context.Context, id string, user auth.User, comment string) (Request, error) {
	req, err := s.Get(ctx, id)
	if err != nil {
		return Request{}, err
	}
	if req.RequestedBySubject != user.Subject && user.EffectiveRole() != "admin" {
		return Request{}, ErrForbidden
	}
	tag, err := s.db.Exec(ctx, `UPDATE approval_requests SET status='cancelled',review_comment=$2,cancelled_at=NOW(),updated_at=NOW() WHERE id=$1 AND status='pending'`, id, comment)
	if err != nil {
		return Request{}, err
	}
	if tag.RowsAffected() != 1 {
		return Request{}, ErrConflict
	}
	_ = s.apps.UpdateStatus(ctx, req.ApplicationID, "draft")
	_ = s.notifications.Send(ctx, notifications.Message{RecipientSubject: req.RequestedBySubject, Type: "approval_cancelled",
		Title: "Approval cancelled", Body: req.Action + " was cancelled", ResourceType: "approval_request",
		ResourceID: id, ApplicationID: req.ApplicationID, DeduplicationKey: "approval-cancelled:" + id})
	return s.Get(ctx, id)
}
func (s *Service) Retry(ctx context.Context, id string, user auth.User) (Request, error) {
	req, err := s.Get(ctx, id)
	if err != nil {
		return Request{}, err
	}
	if req.Status != "failed" {
		return Request{}, ErrConflict
	}
	tag, err := s.db.Exec(ctx, `UPDATE approval_requests SET status='executing',execution_error=NULL,updated_at=NOW() WHERE id=$1 AND status='failed'`, id)
	if err != nil || tag.RowsAffected() != 1 {
		return Request{}, ErrConflict
	}
	jobID, execErr := s.execute(ctx, req)
	status := "succeeded"
	msg := ""
	if execErr != nil {
		status = "failed"
		msg = execErr.Error()
	}
	_, _ = s.db.Exec(ctx, `UPDATE approval_requests SET status=$2,execution_job_id=NULLIF($3,'')::uuid,execution_error=$4,updated_at=NOW() WHERE id=$1`, id, status, jobID, msg)
	applicationID := req.ApplicationID
	if req.Action == "delete_keycloak_client" && status == "succeeded" {
		applicationID = ""
	}
	_ = s.notifications.Send(ctx, notifications.Message{RecipientSubject: req.RequestedBySubject, Type: "approval_" + status,
		Title: "Approval " + status, Body: req.Action + " " + status, ResourceType: "approval_request",
		ResourceID: id, ApplicationID: applicationID, DeduplicationKey: "approval-retry:" + id + ":" + status})
	return s.Get(ctx, id)
}
func (s *Service) execute(ctx context.Context, req Request) (string, error) {
	switch req.Action {
	case "provision_application":
		return s.executor.Provision(ctx, req.ApplicationID)
	case "update_keycloak_client":
		var input applications.UpdateApplicationRequest
		if err := json.Unmarshal(req.RequestPayload, &input); err != nil {
			return "", err
		}
		return "", s.executor.Update(ctx, req.ApplicationID, input)
	case "delete_keycloak_client":
		return "", s.executor.Delete(ctx, req.ApplicationID, true)
	}
	return "", ErrConflict
}

const approvalSelect = `
	SELECT ar.id,COALESCE(ar.application_id::text,''),COALESCE(a.name,''),ar.action,ar.status,
	ar.requested_by_subject,ar.requested_by_username,ar.request_payload,COALESCE(ar.request_summary,''),
	ar.application_version,COALESCE(ar.reviewed_by_subject,''),COALESCE(ar.reviewed_by_username,''),
	COALESCE(ar.review_comment,''),ar.requested_at,ar.decided_at,ar.cancelled_at,
	COALESCE(ar.execution_job_id::text,''),COALESCE(ar.execution_error,''),ar.created_at,ar.updated_at
	FROM approval_requests ar LEFT JOIN applications a ON a.id=ar.application_id`

type scanner interface{ Scan(...any) error }

func scan(row scanner) (Request, error) {
	var r Request
	err := row.Scan(&r.ID, &r.ApplicationID, &r.ApplicationName, &r.Action, &r.Status,
		&r.RequestedBySubject, &r.RequestedByUsername, &r.RequestPayload, &r.RequestSummary, &r.ApplicationVersion,
		&r.ReviewedBySubject, &r.ReviewedByUsername, &r.ReviewComment, &r.RequestedAt, &r.DecidedAt, &r.CancelledAt,
		&r.ExecutionJobID, &r.ExecutionError, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}
