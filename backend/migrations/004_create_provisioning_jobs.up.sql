CREATE TABLE provisioning_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    application_id UUID REFERENCES applications(id) ON DELETE CASCADE,

    action VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    started_at TIMESTAMP,
    completed_at TIMESTAMP,

    error_message TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE provisioning_job_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    job_id UUID NOT NULL REFERENCES provisioning_jobs(id) ON DELETE CASCADE,

    step_name VARCHAR(150) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    error_message TEXT,

    started_at TIMESTAMP,
    completed_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_provisioning_jobs_application_id ON provisioning_jobs(application_id);
CREATE INDEX idx_provisioning_jobs_status ON provisioning_jobs(status);
CREATE INDEX idx_provisioning_job_steps_job_id ON provisioning_job_steps(job_id);