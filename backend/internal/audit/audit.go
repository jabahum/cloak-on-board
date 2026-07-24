package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Log struct {
	ID            string          `json:"id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	ActorSubject  string          `json:"actor_subject"`
	ActorUsername string          `json:"actor_username"`
	ActorEmail    string          `json:"actor_email"`
	ActorRole     string          `json:"actor_role"`
	Action        string          `json:"action"`
	ResourceType  string          `json:"resource_type"`
	ResourceID    string          `json:"resource_id"`
	ApplicationID string          `json:"application_id"`
	RequestID     string          `json:"request_id"`
	Result        string          `json:"result"`
	StatusCode    int             `json:"status_code"`
	SourceIP      string          `json:"source_ip"`
	UserAgent     string          `json:"user_agent"`
	BeforeData    json.RawMessage `json:"before_data,omitempty"`
	AfterData     json.RawMessage `json:"after_data,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	ErrorMessage  string          `json:"error_message,omitempty"`
}

type Filter struct {
	Actor, Action, ResourceType, ApplicationID, Result, From, To string
	Page, PageSize                                               int
}

type Service struct{ db *pgxpool.Pool }

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) Record(ctx context.Context, log Log) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO audit_logs (
			actor_subject, actor_username, actor_email, actor_role, action,
			resource_type, resource_id, application_id, request_id, result,
			status_code, source_ip, user_agent, before_data, after_data, metadata, error_message
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NULLIF($8,'')::uuid,$9,$10,$11,$12,$13,$14,$15,$16,$17)
	`, log.ActorSubject, log.ActorUsername, log.ActorEmail, log.ActorRole, log.Action,
		log.ResourceType, log.ResourceID, log.ApplicationID, log.RequestID, log.Result,
		log.StatusCode, log.SourceIP, log.UserAgent, nullableJSON(log.BeforeData),
		nullableJSON(log.AfterData), nullableJSON(log.Metadata), log.ErrorMessage)
	return err
}

func (s *Service) List(ctx context.Context, f Filter) ([]Log, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 25
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, occurred_at, COALESCE(actor_subject,''), COALESCE(actor_username,''),
		       COALESCE(actor_email,''), COALESCE(actor_role,''), action, resource_type,
		       COALESCE(resource_id,''), COALESCE(application_id::text,''), COALESCE(request_id,''),
		       result, status_code, COALESCE(source_ip,''), COALESCE(user_agent,''),
		       COALESCE(before_data,'null'::jsonb), COALESCE(after_data,'null'::jsonb),
		       COALESCE(metadata,'null'::jsonb), COALESCE(error_message,'')
		FROM audit_logs
		WHERE ($1 = '' OR actor_subject = $1 OR actor_username ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR action = $2)
		  AND ($3 = '' OR resource_type = $3)
		  AND ($4 = '' OR application_id::text = $4)
		  AND ($5 = '' OR result = $5)
		  AND ($6 = '' OR occurred_at >= $6::timestamp)
		  AND ($7 = '' OR occurred_at <= $7::timestamp)
		ORDER BY occurred_at DESC LIMIT $8 OFFSET $9
	`, f.Actor, f.Action, f.ResourceType, f.ApplicationID, f.Result, f.From, f.To,
		f.PageSize, (f.Page-1)*f.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := []Log{}
	for rows.Next() {
		var item Log
		if err := rows.Scan(&item.ID, &item.OccurredAt, &item.ActorSubject, &item.ActorUsername,
			&item.ActorEmail, &item.ActorRole, &item.Action, &item.ResourceType, &item.ResourceID,
			&item.ApplicationID, &item.RequestID, &item.Result, &item.StatusCode, &item.SourceIP,
			&item.UserAgent, &item.BeforeData, &item.AfterData, &item.Metadata, &item.ErrorMessage); err != nil {
			return nil, err
		}
		logs = append(logs, item)
	}
	return logs, rows.Err()
}

func (s *Service) Get(ctx context.Context, id string) (Log, error) {
	var item Log
	err := s.db.QueryRow(ctx, `
		SELECT id, occurred_at, COALESCE(actor_subject,''), COALESCE(actor_username,''),
		       COALESCE(actor_email,''), COALESCE(actor_role,''), action, resource_type,
		       COALESCE(resource_id,''), COALESCE(application_id::text,''), COALESCE(request_id,''),
		       result, status_code, COALESCE(source_ip,''), COALESCE(user_agent,''),
		       COALESCE(before_data,'null'::jsonb), COALESCE(after_data,'null'::jsonb),
		       COALESCE(metadata,'null'::jsonb), COALESCE(error_message,'')
		FROM audit_logs WHERE id=$1
	`, id).Scan(&item.ID, &item.OccurredAt, &item.ActorSubject, &item.ActorUsername,
		&item.ActorEmail, &item.ActorRole, &item.Action, &item.ResourceType, &item.ResourceID,
		&item.ApplicationID, &item.RequestID, &item.Result, &item.StatusCode, &item.SourceIP,
		&item.UserAgent, &item.BeforeData, &item.AfterData, &item.Metadata, &item.ErrorMessage)
	return item, err
}

func Middleware(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		var metadata json.RawMessage
		if c.Request.Body != nil && strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			body, _ := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			var payload any
			if json.Unmarshal(body, &payload) == nil {
				metadata, _ = json.Marshal(map[string]any{"request": Redact(payload)})
			}
		}
		c.Next()
		user, _ := auth.GetUser(c)
		requestID, _ := c.Get("request_id")
		resource := strings.Trim(strings.Split(strings.TrimPrefix(c.FullPath(), "/"), "/")[0], ":")
		result := "success"
		if c.Writer.Status() >= 400 {
			result = "failure"
		}
		applicationID := ""
		if strings.HasPrefix(c.FullPath(), "/applications/") && c.Request.Method != "DELETE" {
			applicationID = c.Param("id")
		}
		_ = service.Record(c.Request.Context(), Log{
			ActorSubject: user.Subject, ActorUsername: user.Username, ActorEmail: user.Email,
			ActorRole: user.EffectiveRole(), Action: c.Request.Method + " " + c.FullPath(),
			ResourceType: resource, ResourceID: firstNonEmpty(c.Param("mapperId"), c.Param("scopeId"), c.Param("id")),
			ApplicationID: applicationID, RequestID: asString(requestID), Result: result,
			StatusCode: c.Writer.Status(), SourceIP: c.ClientIP(), UserAgent: c.Request.UserAgent(), Metadata: metadata,
		})
	}
}

func Redact(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for key, item := range typed {
			lower := strings.ToLower(key)
			if strings.Contains(lower, "secret") || strings.Contains(lower, "password") ||
				strings.Contains(lower, "token") || lower == "authorization" {
				out[key] = "[REDACTED]"
			} else {
				out[key] = Redact(item)
			}
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = Redact(item)
		}
		return out
	default:
		return value
	}
}

func nullableJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
func asString(v any) string { s, _ := v.(string); return s }
func ParseInt(value string, fallback int) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}
