import { api, type ApiResponse } from "./client";
import type {
  Application,
  ClientScopeAssignments,
  CreateApplicationPayload,
  ImportApplicationPayload,
  KeycloakClient,
  ProtocolMapper,
  ProtocolMapperPayload,
  UpdateApplicationPayload,
} from "../types/application";
import type { ProvisioningJob } from "../types/provisioning";

export async function listApplications() {
  const res = await api.get<ApiResponse<Application[]>>("/applications");
  return res.data.data;
}

export async function getApplication(id: string) {
  const res = await api.get<ApiResponse<Application>>(`/applications/${id}`);
  return res.data.data;
}

export async function createApplication(payload: CreateApplicationPayload) {
  const res = await api.post<ApiResponse<Application>>(
    "/applications",
    payload,
  );
  return res.data.data;
}

export async function provisionApplication(id: string) {
  const res = await api.post<ApiResponse<ProvisioningJob>>(
    `/applications/${id}/provision`,
  );
  return res.data.data;
}

export async function updateApplication(
  id: string,
  payload: UpdateApplicationPayload,
) {
  const res = await api.put<ApiResponse<Application>>(
    `/applications/${id}`,
    payload,
  );
  return res.data.data;
}

export async function deleteApplication(id: string, deleteKeycloak: boolean) {
  await api.delete(`/applications/${id}`, {
    params: { delete_keycloak: deleteKeycloak },
  });
}

export async function listKeycloakClients(search = "") {
  const res = await api.get<ApiResponse<KeycloakClient[]>>(
    "/keycloak/clients",
    { params: { search } },
  );
  return res.data.data;
}

export async function importApplication(payload: ImportApplicationPayload) {
  const res = await api.post<ApiResponse<Application>>(
    "/applications/import",
    payload,
  );
  return res.data.data;
}

export async function getClientScopes(id: string) {
  const res = await api.get<ApiResponse<ClientScopeAssignments>>(
    `/applications/${id}/client-scopes`,
  );
  return res.data.data;
}

export async function assignClientScope(
  id: string,
  scopeId: string,
  type: "default" | "optional",
) {
  await api.put(`/applications/${id}/client-scopes/${scopeId}`, { type });
}

export async function removeClientScope(
  id: string,
  scopeId: string,
  type: "default" | "optional",
) {
  await api.delete(`/applications/${id}/client-scopes/${scopeId}`, {
    params: { type },
  });
}

export async function listProtocolMappers(id: string) {
  const res = await api.get<ApiResponse<ProtocolMapper[]>>(
    `/applications/${id}/protocol-mappers`,
  );
  return res.data.data;
}

export async function createProtocolMapper(
  id: string,
  payload: ProtocolMapperPayload,
) {
  await api.post(`/applications/${id}/protocol-mappers`, payload);
}

export async function updateProtocolMapper(
  id: string,
  mapperId: string,
  payload: ProtocolMapperPayload,
) {
  await api.put(
    `/applications/${id}/protocol-mappers/${mapperId}`,
    payload,
  );
}

export async function deleteProtocolMapper(id: string, mapperId: string) {
  await api.delete(`/applications/${id}/protocol-mappers/${mapperId}`);
}
