export type ApprovalAction =
  | "provision_application"
  | "update_keycloak_client"
  | "delete_keycloak_client"
  | "deploy_application"
  | "promote_application"
  | "rollback_deployment"
  | "reconcile_drift"
  | "rotate_client_secret";
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
  environment_id?: string;
  realm_connection_id?: string;
  deployment_id?: string;
  snapshot_id?: string;
  reviewed_by_username: string;
  review_comment: string;
  requested_at: string;
  execution_job_id: string;
  execution_error: string;
};
