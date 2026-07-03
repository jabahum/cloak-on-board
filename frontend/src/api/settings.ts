import { api, type ApiResponse } from "./client";
import type { SaveSettingsPayload, Settings } from "../types/settings";

export async function getSettings() {
  const res = await api.get<ApiResponse<Settings>>("/settings");
  return res.data.data;
}

export async function saveSettings(payload: SaveSettingsPayload) {
  const res = await api.put<ApiResponse<Settings>>("/settings", payload);
  return res.data.data;
}
