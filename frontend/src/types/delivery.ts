export type Environment = {
  id: string;
  name: string;
  slug: string;
  promotion_order: number;
  protected: boolean;
  enabled: boolean;
};

export type RealmConnection = {
  id: string;
  environment_id: string;
  environment: string;
  name: string;
  base_url: string;
  realm: string;
  admin_client_id: string;
  enabled: boolean;
  secret_set: boolean;
  last_tested_at?: string;
  last_test_status: string;
  last_test_error?: string;
};

export type Snapshot = {
  id: string;
  application_id: string;
  version: number;
  configuration_hash: string;
  created_by_username: string;
  created_at: string;
};

export type Deployment = {
  id: string;
  application_id: string;
  application_name: string;
  environment_id: string;
  environment: string;
  promotion_order: number;
  protected: boolean;
  realm_connection_id: string;
  realm: string;
  snapshot_id: string;
  previous_snapshot_id?: string;
  keycloak_client_uuid?: string;
  keycloak_client_id: string;
  status: string;
  drift_status: string;
  deployed_at?: string;
};

export type DriftFinding = {
  id: string;
  path: string;
  change_type: string;
  desired_value?: unknown;
  actual_value?: unknown;
};

export type DriftRun = {
  id: string;
  deployment_id: string;
  status: string;
  started_at: string;
  completed_at?: string;
  findings: DriftFinding[];
};

