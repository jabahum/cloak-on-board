ALTER TABLE applications
ADD COLUMN config_version INTEGER NOT NULL DEFAULT 1;

UPDATE applications
SET status = CASE
    WHEN status = 'provisioned' THEN 'provisioned'
    WHEN status = 'failed' THEN 'failed'
    ELSE 'draft'
END;

CREATE TABLE user_profiles (
    subject VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    display_name VARCHAR(255),
    effective_role VARCHAR(20) NOT NULL,
    last_seen_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    occurred_at TIMESTAMP NOT NULL DEFAULT NOW(),
    actor_subject VARCHAR(255),
    actor_username VARCHAR(255),
    actor_email VARCHAR(255),
    actor_role VARCHAR(20),
    action VARCHAR(150) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    application_id UUID REFERENCES applications(id) ON DELETE SET NULL,
    request_id VARCHAR(100),
    result VARCHAR(20) NOT NULL,
    status_code INTEGER NOT NULL,
    source_ip VARCHAR(100),
    user_agent TEXT,
    before_data JSONB,
    after_data JSONB,
    metadata JSONB,
    error_message TEXT
);

CREATE INDEX audit_logs_occurred_at_idx ON audit_logs (occurred_at DESC);
CREATE INDEX audit_logs_actor_idx ON audit_logs (actor_subject);
CREATE INDEX audit_logs_action_idx ON audit_logs (action);
CREATE INDEX audit_logs_resource_idx ON audit_logs (resource_type, resource_id);
CREATE INDEX audit_logs_application_idx ON audit_logs (application_id);
CREATE INDEX audit_logs_result_idx ON audit_logs (result);

CREATE TABLE approval_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID REFERENCES applications(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    requested_by_subject VARCHAR(255) NOT NULL,
    requested_by_username VARCHAR(255) NOT NULL,
    request_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    request_summary TEXT,
    application_version INTEGER NOT NULL,
    reviewed_by_subject VARCHAR(255),
    reviewed_by_username VARCHAR(255),
    review_comment TEXT,
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),
    decided_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    execution_job_id UUID REFERENCES provisioning_jobs(id) ON DELETE SET NULL,
    execution_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX approval_requests_active_unique
ON approval_requests (application_id, action)
WHERE status IN ('pending', 'approved', 'executing');
CREATE INDEX approval_requests_status_idx ON approval_requests (status, requested_at DESC);
CREATE INDEX approval_requests_requester_idx ON approval_requests (requested_by_subject, requested_at DESC);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipient_subject VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    application_id UUID REFERENCES applications(id) ON DELETE SET NULL,
    deduplication_key VARCHAR(255),
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata JSONB
);

CREATE UNIQUE INDEX notifications_deduplication_unique
ON notifications (recipient_subject, deduplication_key)
WHERE deduplication_key IS NOT NULL;
CREATE INDEX notifications_recipient_idx
ON notifications (recipient_subject, created_at DESC);
CREATE INDEX notifications_unread_idx
ON notifications (recipient_subject, created_at DESC)
WHERE read_at IS NULL;
