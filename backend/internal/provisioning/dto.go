package provisioning

type CreateJobRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
	Action        string `json:"action" binding:"required"`
}

type CreateStepRequest struct {
	JobID    string `json:"job_id" binding:"required"`
	StepName string `json:"step_name" binding:"required"`
}
