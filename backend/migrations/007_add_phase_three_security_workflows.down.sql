DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS approval_requests;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS user_profiles;

ALTER TABLE applications
DROP COLUMN IF EXISTS config_version;
