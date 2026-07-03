import axios from "axios";

const apiAuthToken = import.meta.env.VITE_API_AUTH_TOKEN;

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? "http://localhost:9000/api/v1",
  headers: {
    "Content-Type": "application/json",
    ...(apiAuthToken
      ? {
          Authorization: `Bearer ${apiAuthToken}`,
        }
      : {}),
  },
});

export type ApiResponse<T> = {
  data: T;
};
