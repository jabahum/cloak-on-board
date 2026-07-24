DROP INDEX IF EXISTS applications_keycloak_client_id_unique;
DROP INDEX IF EXISTS applications_keycloak_client_uuid_unique;

ALTER TABLE provisioning_jobs
DROP CONSTRAINT provisioning_jobs_application_id_fkey,
ADD CONSTRAINT provisioning_jobs_application_id_fkey
FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE applications
DROP COLUMN IF EXISTS enabled,
DROP COLUMN IF EXISTS source;
