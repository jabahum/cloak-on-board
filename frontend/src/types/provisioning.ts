export type ProvisioningStep = {
  id: string;
  job_id: string;
  step_name: string;
  status: string;
  error_message: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
};

export type ProvisioningJob = {
  id: string;
  application_id: string;
  action: string;
  status: string;
  error_message: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  steps?: ProvisioningStep[];
};
