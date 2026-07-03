package provisioning

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

func (r *Repository) ListJobs(ctx context.Context) ([]Job, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id,
		       COALESCE(application_id::text, ''),
		       action,
		       status,
		       started_at,
		       completed_at,
		       COALESCE(error_message, ''),
		       created_at,
		       updated_at
		FROM provisioning_jobs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := []Job{}

	for rows.Next() {
		var job Job

		if err := rows.Scan(
			&job.ID,
			&job.ApplicationID,
			&job.Action,
			&job.Status,
			&job.StartedAt,
			&job.CompletedAt,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.UpdatedAt,
		); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

func (r *Repository) GetJobByID(ctx context.Context, id string) (Job, error) {
	var job Job

	err := r.db.QueryRow(ctx, `
		SELECT id,
		       COALESCE(application_id::text, ''),
		       action,
		       status,
		       started_at,
		       completed_at,
		       COALESCE(error_message, ''),
		       created_at,
		       updated_at
		FROM provisioning_jobs
		WHERE id = $1
	`, id).Scan(
		&job.ID,
		&job.ApplicationID,
		&job.Action,
		&job.Status,
		&job.StartedAt,
		&job.CompletedAt,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return Job{}, err
	}

	steps, err := r.ListSteps(ctx, id)
	if err != nil {
		return Job{}, err
	}

	job.Steps = steps

	return job, nil
}

func (r *Repository) CreateJob(ctx context.Context, req CreateJobRequest) (Job, error) {
	var id string

	err := r.db.QueryRow(ctx, `
		INSERT INTO provisioning_jobs (
			application_id,
			action,
			status
		)
		VALUES ($1, $2, 'pending')
		RETURNING id
	`, req.ApplicationID, req.Action).Scan(&id)

	if err != nil {
		return Job{}, err
	}

	return r.GetJobByID(ctx, id)
}

func (r *Repository) MarkJobRunning(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE provisioning_jobs
		SET status = 'running',
		    started_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`, id)

	return err
}

func (r *Repository) MarkJobSucceeded(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE provisioning_jobs
		SET status = 'succeeded',
		    completed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`, id)

	return err
}

func (r *Repository) MarkJobFailed(ctx context.Context, id string, message string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE provisioning_jobs
		SET status = 'failed',
		    error_message = $2,
		    completed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`, id, message)

	return err
}

func (r *Repository) CreateStep(ctx context.Context, req CreateStepRequest) (JobStep, error) {
	var id string

	err := r.db.QueryRow(ctx, `
		INSERT INTO provisioning_job_steps (
			job_id,
			step_name,
			status
		)
		VALUES ($1, $2, 'pending')
		RETURNING id
	`, req.JobID, req.StepName).Scan(&id)

	if err != nil {
		return JobStep{}, err
	}

	return r.GetStepByID(ctx, id)
}

func (r *Repository) GetStepByID(ctx context.Context, id string) (JobStep, error) {
	var step JobStep

	err := r.db.QueryRow(ctx, `
		SELECT id,
		       job_id,
		       step_name,
		       status,
		       COALESCE(error_message, ''),
		       started_at,
		       completed_at,
		       created_at
		FROM provisioning_job_steps
		WHERE id = $1
	`, id).Scan(
		&step.ID,
		&step.JobID,
		&step.StepName,
		&step.Status,
		&step.ErrorMessage,
		&step.StartedAt,
		&step.CompletedAt,
		&step.CreatedAt,
	)

	return step, err
}

func (r *Repository) ListSteps(ctx context.Context, jobID string) ([]JobStep, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id,
		       job_id,
		       step_name,
		       status,
		       COALESCE(error_message, ''),
		       started_at,
		       completed_at,
		       created_at
		FROM provisioning_job_steps
		WHERE job_id = $1
		ORDER BY created_at ASC
	`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	steps := []JobStep{}

	for rows.Next() {
		var step JobStep

		if err := rows.Scan(
			&step.ID,
			&step.JobID,
			&step.StepName,
			&step.Status,
			&step.ErrorMessage,
			&step.StartedAt,
			&step.CompletedAt,
			&step.CreatedAt,
		); err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, rows.Err()
}

func (r *Repository) MarkStepRunning(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE provisioning_job_steps
		SET status = 'running',
		    started_at = NOW()
		WHERE id = $1
	`, id)

	return err
}

func (r *Repository) MarkStepSucceeded(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE provisioning_job_steps
		SET status = 'succeeded',
		    completed_at = NOW()
		WHERE id = $1
	`, id)

	return err
}

func (r *Repository) MarkStepFailed(ctx context.Context, id string, message string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE provisioning_job_steps
		SET status = 'failed',
		    error_message = $2,
		    completed_at = NOW()
		WHERE id = $1
	`, id, message)

	return err
}
