package provisioning

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListJobs(ctx context.Context) ([]Job, error) {
	return s.repo.ListJobs(ctx)
}

func (s *Service) GetJobByID(ctx context.Context, id string) (Job, error) {
	if strings.TrimSpace(id) == "" {
		return Job{}, errors.New("job id is required")
	}

	return s.repo.GetJobByID(ctx, id)
}

func (s *Service) CreateJob(ctx context.Context, req CreateJobRequest) (Job, error) {
	req.ApplicationID = strings.TrimSpace(req.ApplicationID)
	req.Action = strings.TrimSpace(req.Action)

	if req.ApplicationID == "" {
		return Job{}, errors.New("application id is required")
	}

	if req.Action == "" {
		req.Action = "provision_application"
	}

	return s.repo.CreateJob(ctx, req)
}

func (s *Service) CreateStep(ctx context.Context, req CreateStepRequest) (JobStep, error) {
	req.JobID = strings.TrimSpace(req.JobID)
	req.StepName = strings.TrimSpace(req.StepName)

	if req.JobID == "" {
		return JobStep{}, errors.New("job id is required")
	}

	if req.StepName == "" {
		return JobStep{}, errors.New("step name is required")
	}

	return s.repo.CreateStep(ctx, req)
}

func (s *Service) MarkJobRunning(ctx context.Context, id string) error {
	return s.repo.MarkJobRunning(ctx, id)
}

func (s *Service) MarkJobSucceeded(ctx context.Context, id string) error {
	return s.repo.MarkJobSucceeded(ctx, id)
}

func (s *Service) MarkJobFailed(ctx context.Context, id string, message string) error {
	return s.repo.MarkJobFailed(ctx, id, message)
}

func (s *Service) MarkStepRunning(ctx context.Context, id string) error {
	return s.repo.MarkStepRunning(ctx, id)
}

func (s *Service) MarkStepSucceeded(ctx context.Context, id string) error {
	return s.repo.MarkStepSucceeded(ctx, id)
}

func (s *Service) MarkStepFailed(ctx context.Context, id string, message string) error {
	return s.repo.MarkStepFailed(ctx, id, message)
}
