export type ApprovalAction =
  | "provision_application"
  | "update_keycloak_client"
  | "delete_keycloak_client";
export type ApprovalRequest = {
  id: string;
  application_id: string;
  application_name: string;
  action: ApprovalAction;
  status: string;
  requested_by_subject: string;
  requested_by_username: string;
  request_payload: Record<string, unknown>;
  request_summary: string;
  application_version: number;
  reviewed_by_username: string;
  review_comment: string;
  requested_at: string;
  execution_job_id: string;
  execution_error: string;
};
