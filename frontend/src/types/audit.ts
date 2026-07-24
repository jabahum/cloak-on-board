export type AuditLog = {
  id: string;
  occurred_at: string;
  actor_username: string;
  actor_role: string;
  action: string;
  resource_type: string;
  resource_id: string;
  application_id: string;
  request_id: string;
  result: string;
  status_code: number;
  before_data?: unknown;
  after_data?: unknown;
  metadata?: unknown;
  error_message?: string;
};
