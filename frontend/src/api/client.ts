import axios from "axios";
import { getAccessToken, refreshToken } from "../auth/keycloak";

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? "http://localhost:9000/api/v1",
  headers: {
    "Content-Type": "application/json",
  },
});

api.interceptors.request.use(async (config) => {
  try {
    await refreshToken();
  } catch {
    window.location.assign("/");
    throw new Error("Your session expired. Please sign in again.");
  }

  const token = getAccessToken();

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
});

export type ApiResponse<T> = {
  data: T;
};
