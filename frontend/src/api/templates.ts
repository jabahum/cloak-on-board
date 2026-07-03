import { api, type ApiResponse } from "./client";
import type { OnboardingTemplate } from "../types/template";

export async function listTemplates() {
  const res = await api.get<ApiResponse<OnboardingTemplate[]>>("/templates");
  return res.data.data;
}

export async function seedTemplates() {
  const res =
    await api.post<ApiResponse<{ message: string }>>("/templates/seed");
  return res.data.data;
}
