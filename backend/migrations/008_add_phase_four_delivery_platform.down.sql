ALTER TABLE approval_requests
    DROP COLUMN IF EXISTS snapshot_id,
    DROP COLUMN IF EXISTS deployment_id,
    DROP COLUMN IF EXISTS realm_connection_id,
    DROP COLUMN IF EXISTS environment_id;
ALTER TABLE provisioning_jobs
    DROP COLUMN IF EXISTS snapshot_id,
    DROP COLUMN IF EXISTS deployment_id,
    DROP COLUMN IF EXISTS realm_connection_id,
    DROP COLUMN IF EXISTS environment_id;
DROP TABLE IF EXISTS secret_deliveries;
DROP TABLE IF EXISTS drift_findings;
DROP TABLE IF EXISTS drift_runs;
DROP TABLE IF EXISTS application_deployments;
DROP TABLE IF EXISTS application_snapshots;
DROP FUNCTION IF EXISTS prevent_application_snapshot_update();
DROP TABLE IF EXISTS realm_connections;
DROP TABLE IF EXISTS environments;
