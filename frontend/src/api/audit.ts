import { api, type ApiResponse } from "./client";
import type { AuditLog } from "../types/audit";
export async function listAuditLogs(params:Record<string,string|number>={}){const r=await api.get<ApiResponse<AuditLog[]>>("/audit-logs",{params});return r.data.data}
