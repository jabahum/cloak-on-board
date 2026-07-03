package provisioning

import "time"

type Job struct {
	ID            string     `json:"id"`
	ApplicationID string     `json:"application_id"`
	Action        string     `json:"action"`
	Status        string     `json:"status"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	ErrorMessage  string     `json:"error_message"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Steps []JobStep `json:"steps,omitempty"`
}

type JobStep struct {
	ID           string     `json:"id"`
	JobID        string     `json:"job_id"`
	StepName     string     `json:"step_name"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
}
