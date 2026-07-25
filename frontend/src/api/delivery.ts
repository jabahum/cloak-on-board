import { api, type ApiResponse } from "./client";
import type {
  Deployment,
  DriftRun,
  Environment,
  RealmConnection,
  Snapshot,
} from "../types/delivery";

export async function listEnvironments() {
  const response = await api.get<ApiResponse<Environment[]>>("/environments");
  return response.data.data;
}

export async function createEnvironment(payload: {
  name: string;
  slug: string;
  promotion_order: number;
  protected: boolean;
}) {
  const response = await api.post<ApiResponse<Environment>>("/environments", payload);
  return response.data.data;
}

export async function listRealmConnections() {
  const response = await api.get<ApiResponse<RealmConnection[]>>("/realm-connections");
  return response.data.data;
}

export async function createRealmConnection(payload: {
  environment_id: string;
  name: string;
  base_url: string;
  realm: string;
  admin_client_id: string;
  admin_client_secret: string;
}) {
  const response = await api.post<ApiResponse<RealmConnection>>("/realm-connections", payload);
  return response.data.data;
}

export async function testRealmConnection(id: string) {
  const response = await api.post<ApiResponse<RealmConnection>>(`/realm-connections/${id}/test`);
  return response.data.data;
}

export async function listDeployments(applicationId = "") {
  const response = await api.get<ApiResponse<Deployment[]>>("/deployments", {
    params: applicationId ? { application_id: applicationId } : undefined,
  });
  return response.data.data;
}

export async function listSnapshots(applicationId: string) {
  const response = await api.get<ApiResponse<Snapshot[]>>(`/applications/${applicationId}/snapshots`);
  return response.data.data;
}

export async function createSnapshot(applicationId: string) {
  const response = await api.post<ApiResponse<Snapshot>>(`/applications/${applicationId}/snapshots`);
  return response.data.data;
}

export async function deployApplication(applicationId: string, payload: {
  environment_id: string;
  realm_connection_id: string;
  snapshot_id?: string;
  overrides?: Record<string, unknown>;
}) {
  const response = await api.post<ApiResponse<Deployment>>(`/applications/${applicationId}/deployments`, payload);
  return response.data.data;
}

export async function promoteApplication(applicationId: string, payload: {
  source_deployment_id: string;
  destination_environment_id: string;
  realm_connection_id: string;
  overrides?: Record<string, unknown>;
}) {
  const response = await api.post<ApiResponse<Deployment>>(`/applications/${applicationId}/promotions`, payload);
  return response.data.data;
}

export async function checkDrift(deploymentId: string) {
  const response = await api.post<ApiResponse<DriftRun>>(`/deployments/${deploymentId}/drift-checks`);
  return response.data.data;
}

export async function listDriftRuns(deploymentId = "") {
  const response = await api.get<ApiResponse<DriftRun[]>>("/drift-runs", {
    params: deploymentId ? { deployment_id: deploymentId } : undefined,
  });
  return response.data.data;
}

export async function consumeSecretDelivery(id: string) {
  const response = await api.post<ApiResponse<{ secret: string }>>(`/secret-deliveries/${id}/consume`);
  return response.data.data;
}
