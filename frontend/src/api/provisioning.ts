import { api, type ApiResponse } from "./client";
import type { ProvisioningJob } from "../types/provisioning";

export async function listProvisioningJobs() {
  const res = await api.get<ApiResponse<ProvisioningJob[]>>("/jobs");
  return res.data.data;
}

export async function getProvisioningJob(id: string) {
  const res = await api.get<ApiResponse<ProvisioningJob>>(`/jobs/${id}`);
  return res.data.data;
}
