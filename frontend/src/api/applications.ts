import { api, type ApiResponse } from "./client";
import type {
  Application,
  CreateApplicationPayload,
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
