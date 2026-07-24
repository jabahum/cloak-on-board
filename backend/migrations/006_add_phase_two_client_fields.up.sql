ALTER TABLE applications
ADD COLUMN source VARCHAR(20) NOT NULL DEFAULT 'created',
ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT TRUE;

CREATE UNIQUE INDEX applications_keycloak_client_uuid_unique
ON applications (keycloak_client_uuid)
WHERE keycloak_client_uuid IS NOT NULL;

CREATE UNIQUE INDEX applications_keycloak_client_id_unique
ON applications (keycloak_client_id)
WHERE keycloak_client_id IS NOT NULL;

ALTER TABLE provisioning_jobs
DROP CONSTRAINT provisioning_jobs_application_id_fkey,
ADD CONSTRAINT provisioning_jobs_application_id_fkey
FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE SET NULL;
